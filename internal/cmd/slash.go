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
			Name:        "orders",
			Description: "Query orders (sales listings, fulfilled orders, etc) by a variety of filters",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "collection-addr-sell",
					Description: "The collection address of a sell order",
					Required:    false,
					Choices: []*discordgo.ApplicationCommandOptionChoice{
						{Name: "Heroes", Value: data.BitVerseCollections["hero"].Address},
						{Name: "Portals", Value: data.BitVerseCollections["portal"].Address},
					},
				},
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "collection-addr-buy",
					Description: "The collection address of a buy order",
					Required:    false,
					Choices: []*discordgo.ApplicationCommandOptionChoice{
						{Name: "Heroes", Value: data.BitVerseCollections["hero"].Address},
						{Name: "Portals", Value: data.BitVerseCollections["portal"].Address},
					},
				},
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "status",
					Description: "Status of the order",
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
					Name:        "rarity-buy",
					Description: "Filter by NFT rarity for a buy order",
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
					Name:        "rarity-sell",
					Description: "Filter by NFT rarity for a sell order",
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
					Description: "Choose the field to sort the results by",
					Required:    false,
					Choices: []*discordgo.ApplicationCommandOptionChoice{
						{Name: "Created At", Value: "created_at"},
						{Name: "Expired At", Value: "expired_at"},
						{Name: "Sell Quantity", Value: "sell_quantity"},
						{Name: "Buy Quantity", Value: "buy_quantity"},
						{Name: "Buy Quantity With Fees", Value: "buy_quantity_with_fees"},
						{Name: "Updated At", Value: "updated_at"},
					},
				},
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "sort-order",
					Description: "Sort direction",
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

	case "orders":
		// TODO: work out a good way to provide more results without spamming
		cfg := &orders.ListOrdersConfig{PageSize: 5}
		buyMetadata := make(map[string][]string)
		sellMetadata := make(map[string][]string)
		for _, option := range options {
			switch option.Name {
			case "collection-addr-sell":
				cfg.SellTokenAddress = option.StringValue()
			case "collection-addr-buy":
				cfg.BuyTokenAddress = option.StringValue()
			case "status":
				cfg.Status = option.StringValue()
			case "rarity-buy":
				buyMetadata["Rarity"] = []string{option.StringValue()}
			case "rarity-sell":
				sellMetadata["Rarity"] = []string{option.StringValue()}
			case "order-by":
				cfg.OrderBy = option.StringValue()
			case "sort-order":
				cfg.Direction = option.StringValue()
			case "user":
				cfg.User = option.StringValue()
			default:
				continue
			}
		}

		if len(buyMetadata) > 0 {
			data, err := json.Marshal(buyMetadata)
			if err != nil {
				log.Errorf("error serializing buy metadata %#v to json: %v", buyMetadata, err)
			} else {
				cfg.BuyMetadata = string(data[:])
			}
		}

		if len(sellMetadata) > 0 {
			data, err := json.Marshal(sellMetadata)
			if err != nil {
				log.Errorf("error serializing sell metadata %#v to json: %v", sellMetadata, err)
			} else {
				cfg.SellMetadata = string(data[:])
			}
		}

		if cfg.BuyTokenAddress == "" && cfg.SellTokenAddress == "" {
			cfg.SellTokenAddress = data.BitVerseCollections["hero"].Address
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
