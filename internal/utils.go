package internal

import "strings"

const BotCommandPrefix = "!"

func IsBotCommand(content string) bool {
	parts := strings.Fields(content)
	if len(parts) == 0 {
		return false
	}

	return strings.HasPrefix(parts[0], BotCommandPrefix)
}

func ParseBotCommand(content string) []string {
	if !IsBotCommand(content) {
		return nil
	}

	parts := strings.Fields(content)
	parts[0] = strings.ToLower(strings.TrimPrefix(parts[0], BotCommandPrefix))
	return parts
}
