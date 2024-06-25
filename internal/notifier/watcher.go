package notifier

import (
	"context"
	"encoding/json"
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
	CheckInterval = time.Minute
	DMTemplate    = `New cheapest NFT:
- name: %v
- price: %v
- rarity: %v
- token id: %v
- immutascan: %v
- immutable market: %v`
)

type Seen struct {
	ID    string
	Price float64
}

type Watcher struct {
	clients   *api.ClientsManager
	coinbase  *coinbase.CoinbaseClient
	rarity    []string
	sender    *DiscordSender
	seens     []Seen
	started   bool
	stop      chan struct{}
	subs      []string
	threshold float64
}

func NewWatcher(cm *api.ClientsManager, session *discordgo.Session, rarity []string, priceThreshold float64) *Watcher {
	return &Watcher{
		clients:   cm,
		coinbase:  coinbase.GetCoinbaseClientInstance(),
		rarity:    rarity,
		sender:    NewDiscordSender(session),
		subs:      strings.Split(config.GetenvStr("SUBSCRIBERS"), ","),
		threshold: priceThreshold,
	}
}

func (w *Watcher) Start() error {
	log.Infof("starting watcher %v/$%v", w.rarity, w.threshold)

	if w.started {
		log.Infof("watcher %v/$%v already started", w.rarity, w.threshold)
		return nil
	}

	if err := w.loop(); err != nil {
		return err
	}

	w.started = true
	log.Infof("watcher %v/$%v started", w.rarity, w.threshold)
	return nil
}

func (w *Watcher) Stop() {
	log.Infof("stopping watcher %v/$%v", w.rarity, w.threshold)
	close(w.stop)
}

func (w *Watcher) loop() error {
	w.stop = make(chan struct{}, 1)
	rarityJSON, err := json.Marshal(w.rarity)
	if err != nil {
		log.Errorf("could not decode rarity: %v", err)
		return err
	}

	cfg := &orders.ListOrdersConfig{
		BuyTokenType:     handlers.TokenTypeETH,
		PageSize:         10,
		SellTokenAddress: data.BitVerseCollections["hero"].Address,
		Status:           "active",
		OrderBy:          "buy_quantity_with_fees",
		Direction:        "asc",
		SellMetadata:     fmt.Sprintf(`{"Rarity": %s}`, rarityJSON),
	}

	w.check(cfg) // first run on startup

	ticker := time.NewTicker(CheckInterval)
	go func() {
		for {
			select {
			case <-w.stop:
				log.Infof("received stop in watcher %v/$%v", w.rarity, w.threshold)
				ticker.Stop()
				return
			case <-ticker.C:
				log.Debugf("checking watcher %v/$%v", w.rarity, w.threshold)
				w.check(cfg)
			}
		}
	}()

	return nil
}

func (w *Watcher) check(cfg *orders.ListOrdersConfig) {
	result, err := w.clients.OrdersClient.ListOrders(context.Background(), cfg)
	if err != nil {
		log.Error(err)
	}

	if len(result) == 0 {
		log.Infof("no results returned for %v/%v", w.rarity, w.threshold)
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

	if fiatPrice <= w.threshold && !w.alreadySeen(tokenID, cryptoPrice) {
		priceStr := fmt.Sprintf("$%0.2f", fiatPrice)
		msg := fmt.Sprintf(DMTemplate, name, priceStr, w.rarity, tokenID, urls.Immutascan, urls.ImmutableMarket)
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

func (w *Watcher) getPrice(order imxapi.Order) float64 {
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

func (w *Watcher) getCryptoSymbol(str string) coinbase.CryptoSymbol {
	if str == "ERC20" {
		return coinbase.CryptoUSDC
	}

	return coinbase.CryptoSymbol(str)
}

func (w *Watcher) alreadySeen(id string, price float64) bool {
	for _, s := range w.seens {
		if s.ID == id && s.Price == price {
			return true
		}
	}

	return false
}
