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

const MaxRecords = 20

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

	var output []string
	for _, order := range result {
		output = append(output, h.getOutputForOrder(order))
	}

	return &discordgo.InteractionResponseData{Content: strings.Join(output, "\n")}
}

func (h *OrdersHandler) getOutputForOrder(order imxapi.Order) string {
	url := strings.Join([]string{lib.ImmutascanURL, "order", fmt.Sprint(order.OrderId)}, "/")
	ethPrice := h.getPrice(order)
	fiatPrice := ethPrice * h.coinbase.LastSpotPrice
	return fmt.Sprintf(`Order:
- Status: %s
- Price With Fees: %f ETH / %.2f USD
- User: %s
- Date: %s
- Order Details: <%s>`, order.Status, ethPrice, fiatPrice, order.User, order.GetUpdatedTimestamp(), url)
}

func (h *OrdersHandler) getPrice(order imxapi.Order) float64 {
	amount, err := strconv.Atoi(order.GetBuy().Data.QuantityWithFees)
	if err != nil {
		return 0
	}

	decimals := int(*order.GetBuy().Data.Decimals)
	return float64(amount) * math.Pow10(-1*decimals)
}
