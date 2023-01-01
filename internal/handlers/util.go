package handlers

import (
	"fmt"
	"strings"

	"github.com/deadloct/immutablex-go-lib/utils"
)

const (
	OrderPrefixRarible         = "https://rarible.com/token/immutablex"
	OrderPrefixGamestop        = "https://nft.gamestop.com/token"
	OrderPrefixImmutableMarket = "https://market.immutable.com"
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

// Samples:
// https://rarible.com/token/immutablex/0x6465ef3009f3c474774f4afb607a5d600ea71d95:834?tab=overview
// https://nft.gamestop.com/token/0x6465ef3009f3c474774f4afb607a5d600ea71d95/834
// https://market.immutable.com/collections/0x6465ef3009f3c474774f4afb607a5d600ea71d95/assets/834
// https://immutascan.io/address/0x6465ef3009f3c474774f4afb607a5d600ea71d95/1046
type OrderURLs struct {
	Rarible         string
	Gamestop        string
	ImmutableMarket string
	Immutascan      string
}

func GetOrderURLs(tokenAddress string, tokenID string) OrderURLs {
	return OrderURLs{
		Rarible:         fmt.Sprintf("%s/%s:%s", OrderPrefixRarible, tokenAddress, tokenID),
		Gamestop:        fmt.Sprintf("%s/%s/%s", OrderPrefixGamestop, tokenAddress, tokenID),
		ImmutableMarket: fmt.Sprintf("%s/collections/%s/assets/%s", OrderPrefixImmutableMarket, tokenAddress, tokenID),
		Immutascan:      GetImmutascanAssetURL(tokenAddress, tokenID),
	}
}
