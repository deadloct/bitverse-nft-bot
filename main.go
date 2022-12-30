package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/bwmarrin/discordgo"
	"github.com/deadloct/bitverse-nft-bot/internal"
	"github.com/deadloct/bitverse-nft-bot/internal/handlers/create_message"
	log "github.com/sirupsen/logrus"
)

func main() {
	log.Info("verbose logs enabled")
	log.SetLevel(log.DebugLevel)

	discord, err := discordgo.New("Bot " + os.Getenv("BITVERSE_NFT_BOT_AUTH_TOKEN"))
	if err != nil {
		log.Panic(err)
	}

	cm := internal.NewClientManager()
	if err := cm.Start(); err != nil {
		log.Panic(err)
	}
	defer cm.Stop()

	botUser, err := discord.User("@me")
	if err != nil {
		log.Panic(err)
	}

	createMessageHandler := create_message.NewBaseMessageHandler(cm, botUser)
	discord.AddHandler(createMessageHandler.HandleMessage)

	if err := discord.Open(); err != nil {
		log.Panic(err)
	}

	log.Info("Bot is now running. Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc

	log.Info("Bot exiting...")
}
