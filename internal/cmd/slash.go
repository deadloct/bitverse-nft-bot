package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/bwmarrin/discordgo"
	"github.com/deadloct/bitverse-nft-bot/internal/api"
	"github.com/deadloct/bitverse-nft-bot/internal/data"
	"github.com/deadloct/bitverse-nft-bot/internal/handlers"
	"github.com/deadloct/immutablex-cli/lib/orders"
	log "github.com/sirupsen/logrus"
)

const (
	MaxOrderCount     = 5
	DefaultOrderCount = 3
)

type SlashCommands struct {
	clientsManager     *api.ClientsManager
	heroesHandler      *handlers.AssetMessageHandler
	ordersHandler      *handlers.OrdersHandler
	portalsHandler     *handlers.AssetMessageHandler
	registeredCommands []*discordgo.ApplicationCommand
	session            *discordgo.Session
	started            bool
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

	// Remove commands
	for _, v := range s.registeredCommands {
		err := s.session.ApplicationCommandDelete(s.session.State.User.ID, "", v.ID)
		if err != nil {
			log.Errorf("Cannot delete '%v' command: %v", v.Name, err)
		}
	}

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
						"Return this many records (Default %v, Max: %v)",
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
			},
		},
	}

	s.registeredCommands = make([]*discordgo.ApplicationCommand, len(commands))
	for i, v := range commands {
		cmd, err := s.session.ApplicationCommandCreate(s.session.State.User.ID, "", v)
		if err != nil {
			log.Panicf("Cannot create '%v' command: %v", v.Name, err)
		}
		s.registeredCommands[i] = cmd
	}
}

func (s *SlashCommands) commandHandler(sess *discordgo.Session, i *discordgo.InteractionCreate) {
	var response *discordgo.InteractionResponseData

	options := i.ApplicationCommandData().Options

	v := i.ApplicationCommandData().Name
	switch v {
	case "hero":
		id := options[0].IntValue()
		response = s.heroesHandler.HandleCommand(fmt.Sprint(id))

	case "portal":
		id := options[0].IntValue()
		response = s.portalsHandler.HandleCommand(fmt.Sprint(id))

	case "market":
		cfg := &orders.ListOrdersConfig{
			PageSize:         DefaultOrderCount,
			SellTokenAddress: data.BitVerseCollections["hero"].Address,
			Status:           "active",
			OrderBy:          "buy_quantity_with_fees",
			Direction:        "asc",
		}

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
				if pageSize > MaxOrderCount {
					pageSize = MaxOrderCount
				} else if pageSize < 1 {
					pageSize = 1
				}

				cfg.PageSize = pageSize

			case "order-by":
				cfg.OrderBy = option.StringValue()
				if cfg.OrderBy == "" {
					cfg.OrderBy = "buy_quantity_with_fees"
				}

			case "rarity":
				metadata["Rarity"] = []string{option.StringValue()}

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

		if len(metadata) > 0 {
			data, err := json.Marshal(metadata)
			if err != nil {
				log.Errorf("error serializing sell metadata %#v to json: %v", metadata, err)
			} else {
				cfg.SellMetadata = string(data[:])
			}
		}

		log.Debugf("Get orders for cfg %#v", cfg)
		response = s.ordersHandler.HandleCommand(cfg)

	default:
		response = &discordgo.InteractionResponseData{
			Content: fmt.Sprintf("name %s is unrecognized", v),
		}
	}

	sess.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: response,
	})
}
