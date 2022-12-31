package handlers

import (
	"strings"

	"github.com/deadloct/immutablex-cli/lib"
)

func GetImmutascanUserURL(address string) string {
	return strings.Join([]string{lib.ImmutascanURL, "address", address}, "/")
}

func GetImmutascanAssetURL(tokenAddress string, tokenID string) string {
	return strings.Join([]string{
		lib.ImmutascanURL,
		"address",
		tokenAddress,
		tokenID,
	}, "/")
}
