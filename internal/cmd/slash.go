package cmd

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/deadloct/bitverse-nft-bot/internal/api"
	"github.com/deadloct/bitverse-nft-bot/internal/data"
	"github.com/deadloct/bitverse-nft-bot/internal/handlers"
	"github.com/deadloct/immutablex-go-lib/coinbase"
	"github.com/deadloct/immutablex-go-lib/orders"
	log "github.com/sirupsen/logrus"
)

const (
	MaxOrderCount     = 5
	DefaultOrderCount = 3
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
			Name:        "hero",
			Description: "Fetches the hero NFT with the provided ID",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionInteger,
					Name:        "id",
					Description: "The hero ID to retrieve",
					Required:    true,
				},
			},
		},
		{
			Name:        "portal",
			Description: "Fetches the portal NFT with the provided ID",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionInteger,
					Name:        "id",
					Description: "The portal ID to retrieve",
					Required:    true,
				},
			},
		},
		{
			Name:        "rates",
			Description: "Shows the conversion rates of ETH to USD, EUR, and GBP",
		},
		{
			Name:        "market",
			Description: "Query market listings",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "collection",
					Description: "The collection of returned listings (default: Heroes)",
					Required:    false,
					Choices: []*discordgo.ApplicationCommandOptionChoice{
						{Name: "Heroes", Value: data.BitVerseCollections["hero"].Address},
						{Name: "Portals", Value: data.BitVerseCollections["portal"].Address},
					},
				},
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "status",
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
					Name:        "rarity",
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
					Name:        "order-by",
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
					Name:        "sort-direction",
					Description: "Sort direction (Default: Ascending)",
					Required:    false,
					Choices: []*discordgo.ApplicationCommandOptionChoice{
						{Name: "Ascending", Value: "asc"},
						{Name: "Descending", Value: "desc"},
					},
				},
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "user",
					Description: "User address that created the order",
					Required:    false,
				},
				{
					Type: discordgo.ApplicationCommandOptionInteger,
					Name: "count",
					Description: fmt.Sprintf(
						"Return this many records (Default %v, Max: %v with detailed formatting)",
						DefaultOrderCount,
						MaxOrderCount,
					),
					Required: false,
				},
				{
					Type:        discordgo.ApplicationCommandOptionInteger,
					Name:        "token-id",
					Description: "The token ID of the listing",
					Required:    false,
				},
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "output-format",
					Description: "Choose the output format (Default: Summary)",
					Required:    false,
					Choices: []*discordgo.ApplicationCommandOptionChoice{
						{Name: "Detailed", Value: "detailed"},
						{Name: "Summary", Value: "summary"},
					},
				},
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "currency",
					Description: "Currency for ETH conversion (Default: USD)",
					Required:    false,
					Choices: []*discordgo.ApplicationCommandOptionChoice{
						{Name: "USD", Value: coinbase.CurrencyUSD},
						{Name: "EUR", Value: coinbase.CurrencyEUR},
						{Name: "GBP", Value: coinbase.CurrencyGBP},
					},
				},
			},
		},
	}

	// Add new commands
	log.Debug("registering slash commands")
	for _, v := range commands {
		if _, err := s.session.ApplicationCommandCreate(s.session.State.User.ID, "", v); err != nil {
			log.Panicf("cannot create command %v: %v", v.Name, err)
		}

		log.Debug("created command %v", v.Name)
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
	case "rates":
		log.Debug("Handling rates command")
		client := coinbase.GetCoinbaseClientInstance()

		currencies := []coinbase.Currency{
			coinbase.CurrencyUSD, coinbase.CurrencyGBP, coinbase.CurrencyEUR,
		}

		currencyStrings := []string{"1 ETH"}
		for _, curr := range currencies {
			v := client.RetrieveSpotPrice(curr)
			currencyStrings = append(currencyStrings, s.ordersHandler.FormatPrice(v, curr))
		}

		log.Debugf("%+v", currencyStrings)
		response = &discordgo.InteractionResponseData{
			Content: strings.Join(currencyStrings, "\n"),
		}

	case "hero":
		log.Debug("Handling hero command")
		id := options[0].IntValue()
		response = s.heroesHandler.HandleCommand(fmt.Sprint(id))

	case "portal":
		log.Debug("Handling portal command")
		id := options[0].IntValue()
		response = s.portalsHandler.HandleCommand(fmt.Sprint(id))

	case "market":
		log.Debug("Handling market command")
		cfg := &orders.ListOrdersConfig{
			PageSize:         DefaultOrderCount,
			SellTokenAddress: data.BitVerseCollections["hero"].Address,
			Status:           "active",
			OrderBy:          "buy_quantity_with_fees",
			Direction:        "asc",
		}

		format := "summary"
		currency := coinbase.CurrencyUSD
		metadata := make(map[string][]string)
		for _, option := range options {
			switch option.Name {
			case "collection":
				cfg.SellTokenAddress = option.StringValue()
				if cfg.SellTokenAddress == "" {
					cfg.SellTokenAddress = data.BitVerseCollections["hero"].Address
				}

			case "count":
				pageSize := int(option.IntValue())
				cfg.PageSize = pageSize

			case "currenocoinbase.GetCoinbaseClientInstance()cy":
				currency = coinbase.Currency(option.StringValue())

			case "order-by":
				cfg.OrderBy = option.StringValue()
				if cfg.OrderBy == "" {
					cfg.OrderBy = "buy_quantity_with_fees"
				}

			case "rarity":
				metadata["Rarity"] = []string{option.StringValue()}

			case "output-format":
				format = option.StringValue()

			case "sort-direction":
				cfg.Direction = option.StringValue()
				if cfg.Direction == "" {
					cfg.Direction = "asc"
				}

			case "status":
				cfg.Status = option.StringValue()
				if cfg.Status == "" {
					cfg.Status = "active"
				}

			case "token-id":
				tokenID := option.IntValue()
				if tokenID > 0 {
					cfg.SellTokenID = fmt.Sprint(tokenID)
				}

			case "user":
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
				log.Errorf("error serializing sell metadata %#v to json: %v", metadata, err)
			} else {
				cfg.SellMetadata = string(data[:])
			}
		}

		log.Debugf("Get orders for cfg %#v", cfg)
		response = s.ordersHandler.HandleCommand(cfg, format, currency)

	default:
		response = &discordgo.InteractionResponseData{
			Content: fmt.Sprintf("name %s is unrecognized", v),
		}
	}

	_, err := sess.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
		Content: &response.Content,
		Embeds:  &response.Embeds,
	})
	if err != nil {
		log.Error(err)
	}
}
