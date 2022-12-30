package create_message

import (
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/deadloct/bitverse-nft-bot/internal"
)

type HandlerFunc func(s *discordgo.Session, m *discordgo.MessageCreate)

type BaseMessageHandler struct {
	handlers map[string]HandlerFunc
	botUser  *discordgo.User
}

func NewBaseMessageHandler(cm *internal.ClientManager, botUser *discordgo.User) *BaseMessageHandler {
	ordersHandler := NewOrdersHandler(cm)
	handlers := map[string]HandlerFunc{
		"hero":   NewAssetMessageHandler(internal.BitVerseCollections["hero"], cm).HandleMessage,
		"portal": NewAssetMessageHandler(internal.BitVerseCollections["portal"], cm).HandleMessage,
		"orders": ordersHandler.HandleMessage,
	}

	return &BaseMessageHandler{handlers: handlers, botUser: botUser}
}

func (h *BaseMessageHandler) HandleMessage(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == h.botUser.ID {
		return
	}

	parts := strings.Split(strings.TrimSpace(m.Content), " ")
	cmd := strings.ToLower(parts[0])

	if _, ok := h.handlers[cmd]; !ok {
		s.ChannelMessageSend(m.ChannelID, "Unrecognized command")
		return
	}

	h.handlers[cmd](s, m)
}
