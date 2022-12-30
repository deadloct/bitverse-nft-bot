package create_message

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/deadloct/bitverse-nft-bot/internal"
	"github.com/deadloct/immutablex-cli/lib"
	log "github.com/sirupsen/logrus"
)

type AssetMessageHandler struct {
	clientManager *internal.ClientManager
	col           internal.BitVerseCollection
}

func NewAssetMessageHandler(col internal.BitVerseCollection, clientManager *internal.ClientManager) *AssetMessageHandler {
	return &AssetMessageHandler{clientManager: clientManager, col: col}
}

func (h *AssetMessageHandler) HandleMessage(s *discordgo.Session, m *discordgo.MessageCreate) {
	parts := strings.Split(m.Content, " ")
	if len(parts) < 2 {
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Please provide a %s ID", h.col.Singular))
		return
	}

	tokenID := parts[1]
	if _, err := strconv.ParseFloat(tokenID, 64); err != nil {
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("The provided %s token ID is invalid", h.col.Singular))
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	asset, err := h.clientManager.AssetsClient.GetAsset(ctx, h.col.Address, tokenID, true)
	if err != nil {
		log.Error(err)
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Error retrieving %s for token ID %s", h.col.Singular, tokenID))
		return
	}

	if asset == nil || asset.TokenId != tokenID {
		log.Error(err)
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Could not find a %s with token ID %s", h.col.Singular, tokenID))
		return
	}

	url := strings.Join([]string{
		lib.ImmutascanURL,
		"address",
		h.col.Address,
		tokenID,
	}, "/")
	str := fmt.Sprintf("**%s %s**\nLink: %s\nOwner: %s\nStatus: %s", h.col.Singular, tokenID, url, asset.User, asset.Status)

	msg := &discordgo.MessageSend{Content: str}
	if asset.ImageUrl.IsSet() {
		msg.Embeds = []*discordgo.MessageEmbed{
			{Image: &discordgo.MessageEmbedImage{URL: *asset.ImageUrl.Get()}},
		}
	}

	s.ChannelMessageSendComplex(m.ChannelID, msg)
}
