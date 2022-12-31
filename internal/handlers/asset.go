package handlers

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/deadloct/bitverse-nft-bot/internal/api"
	"github.com/deadloct/bitverse-nft-bot/internal/data"
	"github.com/deadloct/immutablex-cli/lib"
	log "github.com/sirupsen/logrus"
)

type AssetMessageHandler struct {
	clientsManager *api.ClientsManager
	col            data.BitVerseCollection
}

func NewAssetMessageHandler(col data.BitVerseCollection, clientsManager *api.ClientsManager) *AssetMessageHandler {
	return &AssetMessageHandler{clientsManager: clientsManager, col: col}
}

func (h *AssetMessageHandler) HandleCommand(tokenID string) *discordgo.InteractionResponseData {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	asset, err := h.clientsManager.AssetsClient.GetAsset(ctx, h.col.Address, tokenID, true)
	if err != nil {
		log.Error(err)
		return &discordgo.InteractionResponseData{Content: fmt.Sprintf("Error retrieving %s for token ID %s", h.col.Singular, tokenID)}
	}

	if asset == nil || asset.TokenId != tokenID {
		log.Error(err)
		return &discordgo.InteractionResponseData{Content: fmt.Sprintf("Could not find a %s with token ID %s", h.col.Singular, tokenID)}
	}

	url := strings.Join([]string{
		lib.ImmutascanURL,
		"address",
		h.col.Address,
		tokenID,
	}, "/")
	content := fmt.Sprintf("**%s %s**\nLink: %s\nOwner: %s\nStatus: %s", h.col.Singular, tokenID, url, asset.User, asset.Status)

	return &discordgo.InteractionResponseData{
		Content: content,
		Embeds: []*discordgo.MessageEmbed{
			{Image: &discordgo.MessageEmbedImage{URL: *asset.ImageUrl.Get()}},
		},
	}
}
