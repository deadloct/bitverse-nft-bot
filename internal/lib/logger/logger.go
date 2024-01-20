package logger

import (
	"github.com/bwmarrin/discordgo"
	log "github.com/sirupsen/logrus"
)

func Debug(sess *discordgo.Session, i *discordgo.Interaction, args ...interface{}) {
	log.WithFields(getFields(sess, i)).Debug(args...)
}

func Debugf(sess *discordgo.Session, i *discordgo.Interaction, format string, args ...interface{}) {
	log.WithFields(getFields(sess, i)).Debugf(format, args...)
}

func Info(sess *discordgo.Session, i *discordgo.Interaction, args ...interface{}) {
	log.WithFields(getFields(sess, i)).Info(args...)
}

func Infof(sess *discordgo.Session, i *discordgo.Interaction, format string, args ...interface{}) {
	log.WithFields(getFields(sess, i)).Infof(format, args...)
}

func Warn(sess *discordgo.Session, i *discordgo.Interaction, args ...interface{}) {
	log.WithFields(getFields(sess, i)).Warn(args...)
}

func Warnf(sess *discordgo.Session, i *discordgo.Interaction, format string, args ...interface{}) {
	log.WithFields(getFields(sess, i)).Warnf(format, args...)
}

func Error(sess *discordgo.Session, i *discordgo.Interaction, args ...interface{}) {
	log.WithFields(getFields(sess, i)).Error(args...)
}

func Errorf(sess *discordgo.Session, i *discordgo.Interaction, format string, args ...interface{}) {
	log.WithFields(getFields(sess, i)).Errorf(format, args...)
}

func Panic(sess *discordgo.Session, i *discordgo.Interaction, args ...interface{}) {
	log.WithFields(getFields(sess, i)).Panic(args...)
}

func Panicf(sess *discordgo.Session, i *discordgo.Interaction, format string, args ...interface{}) {
	log.WithFields(getFields(sess, i)).Panicf(format, args...)
}

func Fatal(sess *discordgo.Session, i *discordgo.Interaction, args ...interface{}) {
	log.WithFields(getFields(sess, i)).Fatal(args...)
}

func Fatalf(sess *discordgo.Session, i *discordgo.Interaction, format string, args ...interface{}) {
	log.WithFields(getFields(sess, i)).Fatalf(format, args...)
}

func getFields(sess *discordgo.Session, i *discordgo.Interaction) log.Fields {
	guild, err := sess.Guild(i.GuildID)
	if err != nil {
		log.Errorf("could not retrieve guild info for guild %v: %v", i.GuildID, err)
	}
	var guildName string
	if guild != nil {
		guildName = guild.Name
	}

	discordChannel, err := sess.Channel(i.ChannelID)
	if err != nil {
		log.Errorf("could not retrieve channel info for channel %v: %v", i.ChannelID, err)
	}
	var discordChannelName string
	if discordChannel != nil {
		discordChannelName = discordChannel.Name
	}

	return log.Fields{
		"guild_name":   guildName,
		"channel_name": discordChannelName,
		"channel_id":   i.ChannelID,
		"guild_id":     i.GuildID,
		"guild_locale": i.GuildLocale,
		"request_id":   i.ID,
	}
}
