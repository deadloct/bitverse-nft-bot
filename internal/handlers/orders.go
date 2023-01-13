package handlers

import (
	"context"
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/deadloct/bitverse-nft-bot/internal/api"
	"github.com/deadloct/immutablex-go-lib/coinbase"
	"github.com/deadloct/immutablex-go-lib/orders"
	"github.com/deadloct/immutablex-go-lib/utils"
	imxapi "github.com/immutable/imx-core-sdk-golang/imx/api"
	log "github.com/sirupsen/logrus"
)

const MaxContentLength = 1900

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
	currency coinbase.Currency,
) *discordgo.InteractionResponseData {

	result, err := h.cm.OrdersClient.ListOrders(context.Background(), cfg)
	if err != nil {
		log.Error(err)
		return &discordgo.InteractionResponseData{Content: "Unable to fetch orders for the provided query"}
	}

	if len(result) == 0 {
		return &discordgo.InteractionResponseData{Content: "No results found"}
	}

	switch format {
	case "summary":
		var summaries []string
		for _, order := range result {
			summaries = append(summaries, h.getSummaryForOrder(order, currency))
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
				content += "\n" + summary
			}
		}

		return &discordgo.InteractionResponseData{Content: content}

	default:
		var embeds []*discordgo.MessageEmbed
		for _, order := range result {
			embeds = append(embeds, h.getEmbedForOrder(order, currency))
		}

		return &discordgo.InteractionResponseData{
			Content: fmt.Sprintf("%v Results", len(result)),
			Embeds:  embeds,
		}
	}
}

func (h *OrdersHandler) getSummaryForOrder(order imxapi.Order, currency coinbase.Currency) string {
	data := order.Sell.GetData()
	tokenID := data.GetTokenId()
	collection := data.GetTokenAddress()
	name := order.Sell.Data.Properties.GetName()
	if name == "" {
		name = "Item " + tokenID
	}

	urls := GetOrderURLs(collection, tokenID)
	ethPrice := h.getPrice(order)
	fiatPrice := ethPrice * h.coinbase.RetrieveSpotPrice(currency)
	priceStr := fmt.Sprintf("%f ETH / %s", ethPrice, h.formatPrice(fiatPrice, currency))

	return fmt.Sprintf("• __%s__: <%s> (%s)", name, urls.Immutascan, priceStr)
}

func (h *OrdersHandler) getEmbedForOrder(order imxapi.Order, currency coinbase.Currency) *discordgo.MessageEmbed {
	data := order.Sell.GetData()
	tokenID := data.GetTokenId()
	collection := data.GetTokenAddress()
	name := order.Sell.Data.Properties.GetName()
	if name == "" {
		name = "Item " + tokenID
	}
	urls := GetOrderURLs(collection, tokenID)
	orderID := order.OrderId
	orderURL := strings.Join([]string{utils.ImmutascanURL, "order", fmt.Sprint(orderID)}, "/")

	ethPrice := h.getPrice(order)
	fiatPrice := ethPrice * h.coinbase.RetrieveSpotPrice(currency)
	priceStr := fmt.Sprintf("%f ETH / %s", ethPrice, h.formatPrice(fiatPrice, currency))

	imageURL := data.Properties.GetImageUrl()
	title := fmt.Sprintf("%s (%s)", name, priceStr)

	return &discordgo.MessageEmbed{
		Title: title,
		URL:   urls.Immutascan,
		Fields: []*discordgo.MessageEmbedField{
			{Name: "Stats", Value: order.Status},
			{Name: "Owner", Value: GetImmutascanUserURL(order.User)},
			{Name: "Immutable Market Listing", Value: urls.ImmutableMarket},
			{Name: "Immutascan Listing", Value: urls.Immutascan},
			{Name: "Gamestop Listing", Value: urls.Gamestop},
			{Name: "Rarible Listing", Value: urls.Rarible},
			{Name: "Record of Listing", Value: orderURL},
		},
		Timestamp: order.GetUpdatedTimestamp(),
		Image:     &discordgo.MessageEmbedImage{URL: imageURL},
	}
}

func (h *OrdersHandler) getPrice(order imxapi.Order) float64 {
	amount, err := strconv.Atoi(order.GetBuy().Data.QuantityWithFees)
	if err != nil {
		return 0
	}

	decimals := int(*order.GetBuy().Data.Decimals)
	return float64(amount) * math.Pow10(-1*decimals)
}

func (h *OrdersHandler) formatPrice(price float64, currency coinbase.Currency) string {
	var symbol string

	switch currency {
	case coinbase.CurrencyEUR:
		symbol = "€"
	case coinbase.CurrencyGBP:
		symbol = "£"
	default:
		symbol = "$"
	}

	return fmt.Sprintf("%s%0.2f", symbol, price)
}
