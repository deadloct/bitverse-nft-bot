package handlers

import (
	"context"
	"fmt"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/deadloct/bitverse-nft-bot/internal/api"
	"github.com/deadloct/bitverse-nft-bot/internal/data"
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

	title := asset.GetName()
	if title == "" {
		title = fmt.Sprintf("%s %s", h.col.Singular, tokenID)
	}

	url := GetImmutascanAssetURL(h.col.Address, tokenID)
	collectionURL := GetImmutascanUserURL(h.col.Address)
	ownerURL := GetImmutascanUserURL(asset.GetUser())

	return &discordgo.InteractionResponseData{
		Content: title,
		Embeds: []*discordgo.MessageEmbed{
			{
				Image: &discordgo.MessageEmbedImage{URL: *asset.ImageUrl.Get()},
				Fields: []*discordgo.MessageEmbedField{
					{Name: "Status", Value: asset.Status},
					{Name: "Owner", Value: ownerURL},
					{Name: "Token ID", Value: tokenID},
					{Name: "Collection", Value: collectionURL},
				},
				URL:       url,
				Timestamp: asset.GetCreatedAt(),
			},
		},
	}
}
