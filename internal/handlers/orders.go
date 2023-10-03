package handlers

import (
	"context"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/deadloct/bitverse-nft-bot/internal/api"
	"github.com/deadloct/immutablex-go-lib/coinbase"
	"github.com/deadloct/immutablex-go-lib/orders"
	"github.com/deadloct/immutablex-go-lib/utils"
	imxapi "github.com/immutable/imx-core-sdk-golang/imx/api"
	log "github.com/sirupsen/logrus"
)

const (
	MaxContentLength    = 1900
	MetadataHeroName    = "BHQ - Hero Name"
	MetadataHeroLevel   = "BHQ - Level"
	ImmutableUSDCSymbol = "ERC20"
)

type Metadata map[string]interface{}

type OrdersHandler struct {
	cm       *api.ClientsManager
	coinbase *coinbase.CoinbaseClient
}

func NewOrdersHandler(cm *api.ClientsManager) *OrdersHandler {
	return &OrdersHandler{
		cm:       cm,
		coinbase: coinbase.GetCoinbaseClientInstance(),
	}
}

func (h *OrdersHandler) HandleCommand(
	cfg *orders.ListOrdersConfig,
	format string,
	currency coinbase.FiatSymbol,
) *discordgo.InteractionResponseData {

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	result, err := h.cm.OrdersClient.ListOrders(ctx, cfg)
	if err != nil {
		log.Error(err)
		return &discordgo.InteractionResponseData{Content: "Unable to fetch orders for the provided query"}
	}

	if len(result) == 0 {
		return &discordgo.InteractionResponseData{Content: "No results found"}
	}

	metadata := make(map[string]Metadata, len(result))
	for _, r := range result {
		data := r.Sell.GetData()
		asset, err := h.cm.AssetsClient.GetAsset(ctx, data.GetTokenAddress(), data.GetTokenId(), false)
		if err != nil {
			log.Errorf("unable to retrieve asset %v: %v", data.GetTokenId(), err)
			continue
		}

		metadata[data.GetTokenId()] = asset.GetMetadata()
	}

	switch format {
	case "summary":
		var summaries []string
		for _, order := range result {
			summaries = append(summaries, h.getSummaryForOrder(order, currency, metadata))
		}

		first := fmt.Sprintf("%v results:", len(result))
		summaries = append([]string{first}, summaries...)

		var content string
		for i, summary := range summaries {
			if len(summary)+len(content) > MaxContentLength {
				content += "\n... (Max Discord length reached)"
				break
			}

			if i == 0 {
				content += summary
			} else {
				content += "\n\n" + summary
			}
		}

		return &discordgo.InteractionResponseData{Content: content}

	default:
		var embeds []*discordgo.MessageEmbed
		for _, order := range result {
			embeds = append(embeds, h.getEmbedForOrder(order, currency, metadata))
		}

		return &discordgo.InteractionResponseData{
			Content: fmt.Sprintf("%v Results", len(result)),
			Embeds:  embeds,
		}
	}
}

func (h *OrdersHandler) FormatPrice(price float64, fiat coinbase.FiatSymbol) string {
	var symbol string

	switch fiat {
	case coinbase.FiatEUR:
		symbol = "€"
	case coinbase.FiatGBP:
		symbol = "£"
	default:
		symbol = "$"
	}

	log.Debugf("asked to format price %0.2f in currency %v", price, fiat)

	return fmt.Sprintf("%s%0.2f", symbol, price)
}

func (h *OrdersHandler) getSummaryForOrder(order imxapi.Order, fiatType coinbase.FiatSymbol, metadata map[string]Metadata) string {
	data := order.Sell.GetData()
	tokenID := data.GetTokenId()
	collection := data.GetTokenAddress()
	name := order.Sell.Data.Properties.GetName()
	if name == "" {
		name = "Item " + tokenID
	}

	urls := GetOrderURLs(collection, tokenID)
	cryptoPrice := h.getPrice(order)
	cryptoSymbol := h.getCryptoSymbol(order.GetBuy().Type)
	fiatPrice := cryptoPrice * h.coinbase.RetrieveSpotPrice(cryptoSymbol, fiatType)
	priceStr := fmt.Sprintf("%f %s / %s", cryptoPrice, cryptoSymbol, h.FormatPrice(fiatPrice, fiatType))

	return fmt.Sprintf("• __%s__ (%s -- Confirm Fees on Web)\n  Hero Name: %s\n  Link: <%s>", name, priceStr, h.getHeroName(tokenID, metadata), urls.Immutascan)
}

func (h *OrdersHandler) getEmbedForOrder(order imxapi.Order, fiatType coinbase.FiatSymbol, metadata map[string]Metadata) *discordgo.MessageEmbed {
	data := order.Sell.GetData()
	tokenID := data.GetTokenId()
	collection := data.GetTokenAddress()
	user := order.GetUser()
	name := order.Sell.Data.Properties.GetName()
	if name == "" {
		name = "Item " + tokenID
	}
	urls := GetOrderURLs(collection, tokenID)
	orderID := order.OrderId
	orderURL := strings.Join([]string{utils.ImmutascanURL, "order", fmt.Sprint(orderID)}, "/")

	cryptoPrice := h.getPrice(order)
	cryptoSymbol := h.getCryptoSymbol(order.GetBuy().Type)
	fiatPrice := cryptoPrice * h.coinbase.RetrieveSpotPrice(cryptoSymbol, fiatType)
	priceStr := fmt.Sprintf("%f %s / %s", cryptoPrice, cryptoSymbol, h.FormatPrice(fiatPrice, fiatType))

	imageURL := data.Properties.GetImageUrl()
	title := fmt.Sprintf("%s (%s -- Confirm Fees on Web)", name, priceStr)

	fields := []*discordgo.MessageEmbedField{
		{Name: "Hero Name", Value: h.getHeroName(tokenID, metadata)},
		{Name: "Stats", Value: order.Status},
		{Name: "Owner", Value: GetImmutascanUserURL(user) + "?tab=1&forSale=true"},
		{Name: "Immutable Market Listing", Value: urls.ImmutableMarket},
		{Name: "Immutascan Listing", Value: urls.Immutascan},
		{Name: "Gamestop Listing", Value: urls.Gamestop},
		{Name: "Rarible Listing", Value: urls.Rarible},
		{Name: "Record of Listing", Value: orderURL},
	}

	return &discordgo.MessageEmbed{
		Title:     title,
		URL:       urls.Immutascan,
		Fields:    fields,
		Timestamp: order.GetUpdatedTimestamp(),
		Image:     &discordgo.MessageEmbedImage{URL: imageURL},
	}
}

func (h *OrdersHandler) getPrice(order imxapi.Order) float64 {
	// Deprecated field, but updates not yet available in imx's go lib.
	price := order.GetBuy().Data.QuantityWithFees
	log.Debugf("price of order with fees: %v", price)
	amount, err := strconv.Atoi(price)
	if err != nil {
		log.Errorf("error converting price to integer, using 0 price: %v", err)
		return 0
	}

	decimals := int(*order.GetBuy().Data.Decimals)
	return float64(amount) * math.Pow10(-1*decimals)
}

func (h *OrdersHandler) getHeroName(tokenID string, metaMap map[string]Metadata) string {
	heroName := "(Unknown)"
	if m, ok := metaMap[tokenID]; ok {
		if n, ok := m[MetadataHeroName]; ok {
			if name, ok := n.(string); ok && name != "" {
				heroName = name
			}
		}
	}

	return heroName
}

func (h *OrdersHandler) getCryptoSymbol(str string) coinbase.CryptoSymbol {
	if str == ImmutableUSDCSymbol {
		return coinbase.CryptoUSDC
	}

	return coinbase.CryptoSymbol(str)
}
