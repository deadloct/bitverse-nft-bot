package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/bwmarrin/discordgo"
	"github.com/deadloct/bitverse-nft-bot/internal/cmd"

	log "github.com/sirupsen/logrus"
)

func main() {
	log.Info("verbose logs enabled")
	log.SetLevel(log.DebugLevel)

	session, err := discordgo.New("Bot " + os.Getenv("BITVERSE_NFT_BOT_AUTH_TOKEN"))
	if err != nil {
		log.Panic(err)
	}

	// Listen for server (guild) messages only
	session.Identify.Intents = discordgo.IntentsGuildMessages

	slash := cmd.NewSlashCommands(session)
	if err := slash.Start(); err != nil {
		log.Panic(err)
	}
	defer slash.Stop()

	log.Info("Bot is now running. Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc

	log.Info("Bot exiting...")
}
