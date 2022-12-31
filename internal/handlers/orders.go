package handlers

import (
	"context"
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/deadloct/bitverse-nft-bot/internal/api"
	"github.com/deadloct/immutablex-cli/lib"
	"github.com/deadloct/immutablex-cli/lib/orders"
	imxapi "github.com/immutable/imx-core-sdk-golang/imx/api"
	log "github.com/sirupsen/logrus"
)

type OrdersHandler struct {
	cm       *api.ClientsManager
	coinbase *lib.CoinbaseClient
}

func NewOrdersHandler(cm *api.ClientsManager) *OrdersHandler {
	return &OrdersHandler{
		cm:       cm,
		coinbase: lib.GetCoinbaseClientInstance(),
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
	url := GetImmutascanAssetURL(collection, tokenID)
	orderID := order.OrderId
	orderUrl := strings.Join([]string{lib.ImmutascanURL, "order", fmt.Sprint(orderID)}, "/")

	ethPrice := h.getPrice(order)
	fiatPrice := ethPrice * h.coinbase.LastSpotPrice
	priceStr := fmt.Sprintf("%f ETH / %.2f USD", ethPrice, fiatPrice)

	imageURL := data.Properties.GetImageUrl()
	title := fmt.Sprintf("%s (%s)", name, priceStr)

	return &discordgo.MessageEmbed{
		Title: title,
		URL:   url,
		Fields: []*discordgo.MessageEmbedField{
			{Name: "Stats", Value: order.Status},
			{Name: "Listing", Value: url},
			{Name: "Owner", Value: GetImmutascanUserURL(order.User)},
			{Name: "Order URL", Value: orderUrl},
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
