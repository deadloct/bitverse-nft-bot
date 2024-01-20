package handlers

import (
	"fmt"
	"strings"

	"github.com/deadloct/bitverse-nft-bot/internal/data"

	"github.com/deadloct/immutablex-go-lib/utils"
)

const (
	OrderPrefixRarible         = "https://rarible.com/token/immutablex"
	OrderPrefixImmutableMarket = "https://market.immutable.com"
	OrderTokenTroveURLFormat   = "https://tokentrove.com/collection/%s/imx-%s"
)

func GetImmutascanUserURL(address string) string {
	return strings.Join([]string{utils.ImmutascanURL, "address", address}, "/")
}

func GetImmutascanAssetURL(tokenAddress string, tokenID string) string {
	return strings.Join([]string{
		utils.ImmutascanURL,
		"address",
		tokenAddress,
		tokenID,
	}, "/")
}

func GetTokenTroveAssetURL(tokenAddress string, tokenID string) string {
	switch tokenAddress {
	case data.BitVerseCollections["hero"].Address:
		return fmt.Sprintf(OrderTokenTroveURLFormat, "BitverseHeroes", tokenID)
	case data.BitVerseCollections["portal"].Address:
		return "Portals are grouped by rarity on TokenTrove not individually"
	default:
		return ""
	}
}

// Samples:
// https://rarible.com/token/immutablex/0x6465ef3009f3c474774f4afb607a5d600ea71d95:834?tab=overview
// https://market.immutable.com/collections/0x6465ef3009f3c474774f4afb607a5d600ea71d95/assets/834
// https://immutascan.io/address/0x6465ef3009f3c474774f4afb607a5d600ea71d95/1046
// https://tokentrove.com/collection/BitverseHeroes/imx-4922
type OrderURLs struct {
	Rarible         string
	TokenTrove      string
	ImmutableMarket string
	Immutascan      string
}

func GetOrderURLs(tokenAddress string, tokenID string) OrderURLs {
	return OrderURLs{
		Rarible:         fmt.Sprintf("%s/%s:%s", OrderPrefixRarible, tokenAddress, tokenID),
		ImmutableMarket: fmt.Sprintf("%s/collections/%s/assets/%s", OrderPrefixImmutableMarket, tokenAddress, tokenID),
		Immutascan:      GetImmutascanAssetURL(tokenAddress, tokenID),
		TokenTrove:      GetTokenTroveAssetURL(tokenAddress, tokenID),
	}
}
