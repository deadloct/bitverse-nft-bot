package notifier

import (
	"context"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/deadloct/bitverse-nft-bot/internal/api"
	"github.com/deadloct/bitverse-nft-bot/internal/config"
	"github.com/deadloct/bitverse-nft-bot/internal/data"
	"github.com/deadloct/bitverse-nft-bot/internal/handlers"
	"github.com/deadloct/immutablex-go-lib/coinbase"
	"github.com/deadloct/immutablex-go-lib/orders"
	imxapi "github.com/immutable/imx-core-sdk-golang/imx/api"
	log "github.com/sirupsen/logrus"
)

const (
	Threshold     float64 = 200
	CheckInterval         = time.Minute
	DMTemplate            = `Cheap NFT available:
- name: %v
- price: %v
- token id: %v
- immutascan: %v
- immutable market: %v`
)

type Seen struct {
	ID    string
	Price float64
}

type Watch struct {
	clients  *api.ClientsManager
	coinbase *coinbase.CoinbaseClient
	sender   *DiscordSender
	seens    []Seen
	started  bool
	stop     chan struct{}
	subs     []string
}

func NewWatch(cm *api.ClientsManager, session *discordgo.Session) *Watch {
	return &Watch{
		clients:  cm,
		coinbase: coinbase.GetCoinbaseClientInstance(),
		sender:   NewDiscordSender(session),
		subs:     strings.Split(config.GetenvStr("SUBSCRIBERS"), ","),
	}
}

func (w *Watch) Start() {
	if w.started {
		return
	}

	w.loop()
	w.started = true
}

func (w *Watch) Stop() {
	close(w.stop)
}

func (w *Watch) loop() chan struct{} {
	w.stop = make(chan struct{}, 1)
	cfg := &orders.ListOrdersConfig{
		BuyTokenType:     handlers.TokenTypeETH,
		PageSize:         10,
		SellTokenAddress: data.BitVerseCollections["hero"].Address,
		Status:           "active",
		OrderBy:          "buy_quantity_with_fees",
		Direction:        "asc",
	}

	w.check(cfg) // first run on startup

	ticker := time.NewTicker(CheckInterval)
	go func() {
		for {
			select {
			case <-w.stop:
				ticker.Stop()
				return
			case <-ticker.C:
				w.check(cfg)
			}
		}
	}()

	return w.stop
}

func (w *Watch) check(cfg *orders.ListOrdersConfig) {
	result, err := w.clients.OrdersClient.ListOrders(context.Background(), cfg)
	if err != nil {
		log.Error(err)
	}

	if len(result) == 0 {
		return
	}

	order := result[0] // just first for now
	data := order.Sell.GetData()
	collection := data.GetTokenAddress()
	tokenID := data.GetTokenId()
	urls := handlers.GetOrderURLs(collection, tokenID)
	name := order.Sell.Data.Properties.GetName()
	if name == "" {
		name = "Item " + tokenID
	}

	cryptoPrice := w.getPrice(order)
	cryptoSymbol := w.getCryptoSymbol(order.GetBuy().Type)
	fiatPrice := cryptoPrice * w.coinbase.RetrieveSpotPrice(cryptoSymbol, coinbase.FiatUSD)
	log.Infof("price of cheapest order with fees: $%0.2f", fiatPrice)

	if fiatPrice <= Threshold && !w.alreadySeen(tokenID, cryptoPrice) {
		priceStr := fmt.Sprintf("$%0.2f", fiatPrice)
		msg := fmt.Sprintf(DMTemplate, name, priceStr, tokenID, urls.Immutascan, urls.ImmutableMarket)
		for _, id := range w.subs {
			if err := w.sender.SendDM(id, msg); err != nil {
				log.Error(err)
				continue
			}

			log.Infof("sent notification about item %v (%v) priced at %v to %v", name, tokenID, priceStr, id)
		}

		w.seens = append(w.seens, Seen{ID: tokenID, Price: cryptoPrice})
		log.Infof("adding %v to seen, no notifications should be sent again", tokenID)
	}
}

func (w *Watch) getPrice(order imxapi.Order) float64 {
	// Deprecated field, but updates not yet available in imx's go lib.
	price := order.GetBuy().Data.QuantityWithFees
	amount, err := strconv.Atoi(price)
	if err != nil {
		log.Errorf("error converting price to integer, using 0 price: %v", err)
		return 0
	}

	decimals := int(*order.GetBuy().Data.Decimals)
	return float64(amount) * math.Pow10(-1*decimals)
}

func (w *Watch) getCryptoSymbol(str string) coinbase.CryptoSymbol {
	if str == "ERC20" {
		return coinbase.CryptoUSDC
	}

	return coinbase.CryptoSymbol(str)
}

func (w *Watch) alreadySeen(id string, price float64) bool {
	for _, s := range w.seens {
		if s.ID == id && s.Price == price {
			return true
		}
	}

	return false
}
