package cmd

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/deadloct/bitverse-nft-bot/internal/api"
	"github.com/deadloct/bitverse-nft-bot/internal/data"
	"github.com/deadloct/bitverse-nft-bot/internal/handlers"
	"github.com/deadloct/bitverse-nft-bot/internal/lib/logger"
	"github.com/deadloct/immutablex-go-lib/coinbase"
	"github.com/deadloct/immutablex-go-lib/orders"
	log "github.com/sirupsen/logrus"
)

const (
	MaxOrderCount     = 5
	DefaultOrderCount = 3

	CMDHero                   = "hero"
	CMDHeroID                 = "id"
	CMDPortal                 = "portal"
	CMDPortalID               = "id"
	CMDRates                  = "rates"
	CMDMarket                 = "market"
	CMDMarketCollection       = "collection"
	CMDMarketStatus           = "status"
	CMDMarketRarity           = "rarity"
	CMDMarketOrderBy          = "order-by"
	CMDMarketSortDirection    = "sort-direction"
	CMDMarketUser             = "user"
	CMDMarketCount            = "count"
	CMDMarketTokenID          = "token-id"
	CMDMarketOutputFormat     = "output-format"
	CMDMarketOutputCurrency   = "output-currency"
	CMDMarketBuyCurrency      = "buy-currency"
	CMDMarketAllBuyCurrencies = "All"
)

type SlashCommands struct {
	clientsManager *api.ClientsManager
	heroesHandler  *handlers.AssetMessageHandler
	ordersHandler  *handlers.OrdersHandler
	portalsHandler *handlers.AssetMessageHandler
	session        *discordgo.Session
	started        bool
}

func NewSlashCommands(session *discordgo.Session) *SlashCommands {
	cm := api.NewClientsManager()
	return &SlashCommands{
		clientsManager: cm,
		heroesHandler:  handlers.NewAssetMessageHandler(data.BitVerseCollections["hero"], cm),
		ordersHandler:  handlers.NewOrdersHandler(cm),
		portalsHandler: handlers.NewAssetMessageHandler(data.BitVerseCollections["portal"], cm),
		session:        session,
	}
}

func (s *SlashCommands) Start() error {
	if s.started {
		return nil
	}

	if err := s.clientsManager.Start(); err != nil {
		return err
	}

	// SlashCommands command handler
	s.session.AddHandler(s.commandHandler)

	// Open up the session
	if err := s.session.Open(); err != nil {
		s.clientsManager.Stop()
		return err
	}

	// Register slash commands
	s.setupCommands()
	s.started = true
	return nil
}

func (s *SlashCommands) Stop() {
	if !s.started {
		return
	}

	s.cleanupCommands()
	s.clientsManager.Stop()
}

func (s *SlashCommands) setupCommands() {
	commands := []*discordgo.ApplicationCommand{
		{
			Name:        CMDHero,
			Description: "Fetches the hero NFT with the provided ID",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionInteger,
					Name:        CMDHeroID,
					Description: "The hero ID to retrieve",
					Required:    true,
				},
			},
		},
		{
			Name:        CMDPortal,
			Description: "Fetches the portal NFT with the provided ID",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionInteger,
					Name:        CMDPortalID,
					Description: "The portal ID to retrieve",
					Required:    true,
				},
			},
		},
		{
			Name:        CMDRates,
			Description: "Shows the conversion rates of ETH to USD, EUR, and GBP",
		},
		{
			Name:        CMDMarket,
			Description: "Query market listings",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        CMDMarketCollection,
					Description: "The collection of returned listings (default: Heroes)",
					Required:    false,
					Choices: []*discordgo.ApplicationCommandOptionChoice{
						{Name: "Heroes", Value: data.BitVerseCollections["hero"].Address},
						{Name: "Portals", Value: data.BitVerseCollections["portal"].Address},
					},
				},
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        CMDMarketStatus,
					Description: "Status of the order (Default: Active)",
					Required:    false,
					Choices: []*discordgo.ApplicationCommandOptionChoice{
						{Name: "Active", Value: "active"},
						{Name: "Filled", Value: "filled"},
						{Name: "Cancelled", Value: "cancelled"},
						{Name: "Expired", Value: "expired"},
						{Name: "Inactive", Value: "inactive"},
					},
				},
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        CMDMarketRarity,
					Description: "Filter by NFT rarity (Default: All)",
					Required:    false,
					Choices: []*discordgo.ApplicationCommandOptionChoice{
						{Name: "Common", Value: "Common"},
						{Name: "Rare", Value: "Rare"},
						{Name: "Epic", Value: "Epic"},
						{Name: "Legendary", Value: "Legendary"},
						{Name: "Mythic", Value: "Mythic"},
					},
				},
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        CMDMarketOrderBy,
					Description: "Choose the field to sort the results by (Default: Price)",
					Required:    false,
					Choices: []*discordgo.ApplicationCommandOptionChoice{
						{Name: "Created At", Value: "created_at"},
						{Name: "Expired At", Value: "expired_at"},
						{Name: "Price", Value: "buy_quantity_with_fees"},
						{Name: "Updated At", Value: "updated_at"},
					},
				},
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        CMDMarketSortDirection,
					Description: "Sort direction (Default: Ascending)",
					Required:    false,
					Choices: []*discordgo.ApplicationCommandOptionChoice{
						{Name: "Ascending", Value: "asc"},
						{Name: "Descending", Value: "desc"},
					},
				},
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        CMDMarketUser,
					Description: "User address that created the order",
					Required:    false,
				},
				{
					Type: discordgo.ApplicationCommandOptionInteger,
					Name: CMDMarketCount,
					Description: fmt.Sprintf(
						"Return this many records (Default %v, Max: %v with detailed formatting)",
						DefaultOrderCount,
						MaxOrderCount,
					),
					Required: false,
				},
				{
					Type:        discordgo.ApplicationCommandOptionInteger,
					Name:        CMDMarketTokenID,
					Description: "The token ID of the listing",
					Required:    false,
				},
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        CMDMarketOutputFormat,
					Description: "Choose the output format (Default: Summary)",
					Required:    false,
					Choices: []*discordgo.ApplicationCommandOptionChoice{
						{Name: "Detailed", Value: "detailed"},
						{Name: "Summary", Value: "summary"},
					},
				},
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        CMDMarketOutputCurrency,
					Description: "Output currency (Default: USD)",
					Required:    false,
					Choices: []*discordgo.ApplicationCommandOptionChoice{
						{Name: "USD", Value: coinbase.FiatUSD},
						{Name: "EUR", Value: coinbase.FiatEUR},
						{Name: "GBP", Value: coinbase.FiatGBP},
					},
				},
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        CMDMarketBuyCurrency,
					Description: "Listing cryptocurrency (Default: ETH)",
					Required:    false,
					Choices: []*discordgo.ApplicationCommandOptionChoice{
						{Name: CMDMarketAllBuyCurrencies, Value: CMDMarketAllBuyCurrencies},
						{Name: "ETH", Value: handlers.TokenTypeETH},
						{Name: "USDC/IMX/Other", Value: handlers.TokenTypeERC20},
					},
				},
			},
		},
	}

	// Add new commands
	log.Debug("registering slash commands")
	for _, v := range commands {
		log.Debugf("registering command %v", v.Name)
		if _, err := s.session.ApplicationCommandCreate(s.session.State.User.ID, "", v); err != nil {
			log.Panicf("cannot create command %v: %v", v.Name, err)
		}

		log.Debugf("created command %v", v.Name)
	}

	log.Debug("finished registering slash commands")
}

func (s *SlashCommands) cleanupCommands() {
	existingCommands, err := s.session.ApplicationCommands(s.session.State.User.ID, "")
	if err != nil {
		log.Errorf("could not retrieve commands to do a pre-startup cleanup")
	}

	log.Debug("cleaning up old slash commands during startup...")
	for _, v := range existingCommands {
		log.Debugf("removing command %v", v.Name)
		if err := s.session.ApplicationCommandDelete(s.session.State.User.ID, "", v.ID); err != nil {
			log.Debugf("unable to remove command %v: %v", v.Name, err)
		} else {
			log.Debugf("removed command %v", v.Name)
		}
	}

	log.Debug("finished old command cleanup")
}

func (s *SlashCommands) commandHandler(sess *discordgo.Session, i *discordgo.InteractionCreate) {
	var response *discordgo.InteractionResponseData

	sess.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{Content: "Loading results..."},
	})

	options := i.ApplicationCommandData().Options

	v := i.ApplicationCommandData().Name
	switch v {
	case CMDRates:
		logger.Info(sess, i.Interaction, "Handling rates command")
		client := coinbase.GetCoinbaseClientInstance()

		cryptos := []coinbase.CryptoSymbol{
			coinbase.CryptoETH, coinbase.CryptoIMX, coinbase.CryptoUSDC,
		}

		fiats := []coinbase.FiatSymbol{
			coinbase.FiatUSD, coinbase.FiatGBP, coinbase.FiatEUR,
		}

		var currencyStrings []string
		for _, crypto := range cryptos {
			str := fmt.Sprintf("1 %s", crypto)
			for _, fiat := range fiats {
				v := client.RetrieveSpotPrice(crypto, fiat)
				str = fmt.Sprintf("%s ≈ %s", str, s.ordersHandler.FormatPrice(v, fiat))
			}
			currencyStrings = append(currencyStrings, str)
		}

		logger.Debugf(sess, i.Interaction, "%+v", currencyStrings)
		response = &discordgo.InteractionResponseData{
			Content: strings.Join(currencyStrings, "\n"),
		}

	case CMDHero:
		logger.Info(sess, i.Interaction, "Handling hero command")
		id := options[0].IntValue()
		response = s.heroesHandler.HandleCommand(fmt.Sprint(id))

	case CMDPortal:
		logger.Info(sess, i.Interaction, "Handling portal command")
		id := options[0].IntValue()
		response = s.portalsHandler.HandleCommand(fmt.Sprint(id))

	case CMDMarket:
		logger.Info(sess, i.Interaction, "Handling market command")
		cfg := &orders.ListOrdersConfig{
			BuyTokenType:     handlers.TokenTypeETH,
			PageSize:         DefaultOrderCount,
			SellTokenAddress: data.BitVerseCollections["hero"].Address,
			Status:           "active",
			OrderBy:          "buy_quantity_with_fees",
			Direction:        "asc",
		}

		format := "summary"
		currency := coinbase.FiatUSD
		metadata := make(map[string][]string)
		for _, option := range options {
			switch option.Name {
			case CMDMarketCollection:
				cfg.SellTokenAddress = option.StringValue()
				if cfg.SellTokenAddress == "" {
					cfg.SellTokenAddress = data.BitVerseCollections["hero"].Address
				}

			case CMDMarketCount:
				pageSize := int(option.IntValue())
				cfg.PageSize = pageSize

			case CMDMarketOutputCurrency:
				currency = coinbase.FiatSymbol(option.StringValue())

			case CMDMarketBuyCurrency:
				switch buyType := option.StringValue(); buyType {
				case CMDMarketAllBuyCurrencies:
					cfg.BuyTokenType = ""
				case handlers.TokenTypeERC20, handlers.TokenTypeETH:
					cfg.BuyTokenType = buyType
				}

			case CMDMarketOrderBy:
				cfg.OrderBy = option.StringValue()
				if cfg.OrderBy == "" {
					cfg.OrderBy = "buy_quantity_with_fees"
				}

			case CMDMarketRarity:
				metadata["Rarity"] = []string{option.StringValue()}

			case CMDMarketOutputFormat:
				format = option.StringValue()

			case CMDMarketSortDirection:
				cfg.Direction = option.StringValue()
				if cfg.Direction == "" {
					cfg.Direction = "asc"
				}

			case CMDMarketStatus:
				cfg.Status = option.StringValue()
				if cfg.Status == "" {
					cfg.Status = "active"
				}

			case CMDMarketTokenID:
				tokenID := option.IntValue()
				if tokenID > 0 {
					cfg.SellTokenID = fmt.Sprint(tokenID)
				}

			case CMDMarketUser:
				cfg.User = option.StringValue()

			default:
				continue
			}
		}

		if cfg.PageSize > MaxOrderCount && format != "summary" {
			cfg.PageSize = MaxOrderCount
		} else if cfg.PageSize < 1 {
			cfg.PageSize = 1
		}

		if len(metadata) > 0 {
			data, err := json.Marshal(metadata)
			if err != nil {
				logger.Errorf(sess, i.Interaction, "error serializing sell metadata %#v to json: %v", metadata, err)
			} else {
				cfg.SellMetadata = string(data[:])
			}
		}

		logger.Debugf(sess, i.Interaction, "Get orders for cfg %#v", cfg)
		response = s.ordersHandler.HandleCommand(cfg, format, currency)

	default:
		logger.Warnf(sess, i.Interaction, "Unknown command: %s", v)
		response = &discordgo.InteractionResponseData{
			Content: fmt.Sprintf("name %s is unrecognized", v),
		}
	}

	_, err := sess.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
		Content: &response.Content,
		Embeds:  &response.Embeds,
	})
	if err != nil {
		logger.Error(sess, i.Interaction, err)
	}
}
