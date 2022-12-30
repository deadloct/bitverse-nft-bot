package main

import (
	"context"
	"fmt"
	"math"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/deadloct/immutablex-cli/lib"
	"github.com/deadloct/immutablex-cli/lib/collections"
	"github.com/deadloct/immutablex-cli/lib/orders"
	"github.com/immutable/imx-core-sdk-golang/imx/api"
	log "github.com/sirupsen/logrus"
)

const (
	MaxRecords = 20
	Trigger    = "!"
)

var BotUser *discordgo.User

func getPrice(order api.Order) float64 {
	amount, err := strconv.Atoi(order.GetBuy().Data.QuantityWithFees)
	if err != nil {
		return 0
	}

	decimals := int(*order.GetBuy().Data.Decimals)
	return float64(amount) * math.Pow10(-1*decimals)
}

func messageHandler(s *discordgo.Session, m *discordgo.MessageCreate) {
	log.Debug("Received message...")

	if m.Author.ID == BotUser.ID {
		log.Debug("Author is the same as the bot user, dropping message")
		return
	}

	if !strings.HasPrefix(m.Content, Trigger) {
		log.Debug("Message is not a command, dropping")
		return
	}

	cm := NewClientManager()
	if err := cm.Start(); err != nil {
		log.Panic(err)
	}
	defer cm.Stop()

	parts := strings.Split(m.Content, " ")
	switch strings.ToLower(parts[0]) {
	case "!hero":
		if len(parts) < 2 {
			s.ChannelMessageSend(m.ChannelID, "Please provide a hero ID")
			return
		}

		tokenID := parts[1]
		if _, err := strconv.ParseFloat(tokenID, 64); err != nil {
			s.ChannelMessageSend(m.ChannelID, "The provided token ID is invalid")
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		shortcut := collections.NewShortcuts().GetShortcutByName("hero")

		asset, err := cm.AssetsClient.GetAsset(ctx, shortcut.Addr, tokenID, true)
		if err != nil {
			log.Error(err)
			s.ChannelMessageSend(m.ChannelID, "Unable to find asset for token ID %s")
			return
		}

		url := strings.Join([]string{
			lib.ImmutascanURL,
			"address",
			shortcut.Addr,
			tokenID,
		}, "/")
		str := fmt.Sprintf("**Hero %s**\nLink: %s\nOwner: %s", tokenID, url, asset.User)

		msg := &discordgo.MessageSend{Content: str}
		if asset.ImageUrl.IsSet() {
			msg.Embeds = []*discordgo.MessageEmbed{
				{Image: &discordgo.MessageEmbedImage{URL: *asset.ImageUrl.Get()}},
			}
		}

		s.ChannelMessageSendComplex(m.ChannelID, msg)

	case "!cheapest":
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
		result, err := cm.OrdersClient.ListOrders(ctx, cfg)
		if err != nil {
			log.Error(err)
			s.ChannelMessageSend(m.ChannelID, "Unable to fetch orders for token %s")
			return
		}

		var output []string
		for _, order := range result {
			url := strings.Join([]string{lib.ImmutascanURL, "order", fmt.Sprint(order.OrderId)}, "/")
			ethPrice := getPrice(order)
			fiatPrice := ethPrice * lib.GetCoinbaseClientInstance().LastSpotPrice
			str := fmt.Sprintf(`Order:
- Status: %s
- Price With Fees: %f ETH / %.2f USD
- User: %s
- Date: %s
- Immutascan: %s%s`, order.Status, ethPrice, fiatPrice, order.User, order.GetUpdatedTimestamp(), url, "\n\n")

			output = append(output, str)
		}

		s.ChannelMessageSend(m.ChannelID, strings.Join(output, "\n"))

	default:
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Unrecognized command %s", parts[0]))
	}
}

func main() {
	discord, err := discordgo.New("Bot " + os.Getenv("BITVERSE_NFT_BOT_AUTH_TOKEN"))
	if err != nil {
		log.Panic(err)
	}

	BotUser, err = discord.User("@me")
	if err != nil {
		log.Panic(err.Error())
	}

	discord.AddHandler(messageHandler)
	if err := discord.Open(); err != nil {
		log.Panic(err)
	}

	log.Info("Bot is now running. Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc

	log.Info("Bot exiting...")
}
