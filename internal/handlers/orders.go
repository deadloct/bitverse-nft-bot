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

func (h *OrdersHandler) HandleCommand(cfg *orders.ListOrdersConfig) *discordgo.InteractionResponseData {
	result, err := h.cm.OrdersClient.ListOrders(context.Background(), cfg)
	if err != nil {
		log.Error(err)
		return &discordgo.InteractionResponseData{Content: "Unable to fetch orders for the provided query"}
	}

	if len(result) == 0 {
		return &discordgo.InteractionResponseData{Content: "No results found"}
	}

	var embeds []*discordgo.MessageEmbed
	for _, order := range result {
		embeds = append(embeds, h.getEmbedForOrder(order))
	}

	return &discordgo.InteractionResponseData{
		Content: fmt.Sprintf("%v Results", len(result)),
		Embeds:  embeds,
	}
}

func (h *OrdersHandler) getEmbedForOrder(order imxapi.Order) *discordgo.MessageEmbed {
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
	fiatPrice := ethPrice * h.coinbase.RetrieveSpotPrice()
	priceStr := fmt.Sprintf("%f ETH / %.2f USD", ethPrice, fiatPrice)

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
