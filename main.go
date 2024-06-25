package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/bwmarrin/discordgo"
	"github.com/deadloct/bitverse-nft-bot/internal/api"
	"github.com/deadloct/bitverse-nft-bot/internal/cmd"
	"github.com/deadloct/bitverse-nft-bot/internal/config"
	"github.com/deadloct/bitverse-nft-bot/internal/notifier"

	log "github.com/sirupsen/logrus"
)

func main() {
	log.Info("verbose logs enabled")
	log.SetLevel(log.DebugLevel)

	config.LoadEnvFiles()

	session, err := discordgo.New("Bot " + config.GetenvStr("AUTH_TOKEN"))
	if err != nil {
		log.Panic(err)
	}

	// Listen for server (guild) messages only
	session.Identify.Intents = discordgo.IntentsGuildMessages

	cm := api.NewClientsManager()

	// Slash command controller
	slash := cmd.NewSlashCommands(cm, session)
	if err := slash.Start(); err != nil {
		log.Panic(err)
	}
	defer slash.Stop()

	// Loop price watchers
	commonWatcher := notifier.NewWatcher(cm, session, []string{"Common"}, 250.0)
	commonWatcher.Start()
	defer commonWatcher.Stop()

	rareWatcher := notifier.NewWatcher(cm, session, []string{"Rare"}, 550.0)
	rareWatcher.Start()
	defer rareWatcher.Stop()

	epicLegMythWatcher := notifier.NewWatcher(cm, session, []string{"Epic", "Legendary", "Mythic"}, 800.0)
	epicLegMythWatcher.Start()
	defer epicLegMythWatcher.Stop()

	log.Info("Bot is now running. Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc

	log.Info("Bot exiting...")
}
