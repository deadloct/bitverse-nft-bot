package create_message

import (
	"github.com/bwmarrin/discordgo"
	"github.com/deadloct/bitverse-nft-bot/internal"
	log "github.com/sirupsen/logrus"
)

type HandlerFunc func(s *discordgo.Session, m *discordgo.MessageCreate)

type BaseMessageHandler struct {
	botUser  *discordgo.User
	handlers map[string]HandlerFunc
}

func NewBaseMessageHandler(cm *internal.ClientManager, botUser *discordgo.User) *BaseMessageHandler {
	ordersHandler := NewOrdersHandler(cm)
	handlers := map[string]HandlerFunc{
		"hero":   NewAssetMessageHandler(internal.BitVerseCollections["hero"], cm).HandleMessage,
		"portal": NewAssetMessageHandler(internal.BitVerseCollections["portal"], cm).HandleMessage,
		"orders": ordersHandler.HandleMessage,
	}

	return &BaseMessageHandler{
		botUser:  botUser,
		handlers: handlers,
	}
}

func (h *BaseMessageHandler) HandleMessage(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == h.botUser.ID {
		return
	}

	if !internal.IsBotCommand(m.Content) {
		return
	}

	parts := internal.ParseBotCommand(m.Content)
	if _, ok := h.handlers[parts[0]]; !ok {
		s.ChannelMessageSend(m.ChannelID, "Unrecognized command")
		return
	}

	log.Debugf("processing message: %#v", m.Content)
	h.handlers[parts[0]](s, m)
}
