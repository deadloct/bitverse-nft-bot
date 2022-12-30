package create_message

import (
	"context"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/deadloct/bitverse-nft-bot/internal"
	"github.com/deadloct/immutablex-cli/lib"
	"github.com/deadloct/immutablex-cli/lib/collections"
	"github.com/deadloct/immutablex-cli/lib/orders"
	"github.com/immutable/imx-core-sdk-golang/imx/api"
	log "github.com/sirupsen/logrus"
)

const MaxRecords = 20

type OrdersHandler struct {
	cm       *internal.ClientManager
	coinbase *lib.CoinbaseClient
}

func NewOrdersHandler(cm *internal.ClientManager) *OrdersHandler {
	return &OrdersHandler{
		cm:       cm,
		coinbase: lib.GetCoinbaseClientInstance(),
	}
}

func (h *OrdersHandler) HandleMessage(s *discordgo.Session, m *discordgo.MessageCreate) {
	parts := strings.Split(m.Content, " ")
	count := 5
	if len(parts) > 1 {
		potentialCount := parts[1]
		if c, err := strconv.ParseFloat(potentialCount, 64); err == nil {
			count = int(math.Min(float64(MaxRecords), c))
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	shortcut := collections.NewShortcuts().GetShortcutByName("hero")

	cfg := &orders.ListOrdersConfig{
		Direction:        "asc",
		SellTokenAddress: shortcut.Addr,
		PageSize:         count,
		Status:           "active",
		OrderBy:          "buy_quantity",
	}
	result, err := h.cm.OrdersClient.ListOrders(ctx, cfg)
	if err != nil {
		log.Error(err)
		s.ChannelMessageSend(m.ChannelID, "Unable to fetch orders for token %s")
		return
	}

	var output []string
	for _, order := range result {
		output = append(output, h.getOutputForOrder(order))
	}

	s.ChannelMessageSend(m.ChannelID, strings.Join(output, "\n\n"))
}

func (h *OrdersHandler) getOutputForOrder(order api.Order) string {
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

func (h *OrdersHandler) getPrice(order api.Order) float64 {
	amount, err := strconv.Atoi(order.GetBuy().Data.QuantityWithFees)
	if err != nil {
		return 0
	}

	decimals := int(*order.GetBuy().Data.Decimals)
	return float64(amount) * math.Pow10(-1*decimals)
}
