package handlers

import (
	"strings"

	"github.com/deadloct/immutablex-go-lib/utils"
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
