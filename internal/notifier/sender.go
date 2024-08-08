package notifier

import (
	"github.com/bwmarrin/discordgo"
)

type SendingFunc func(str string) (*discordgo.Message, error)

type DiscordSender struct {
	session *discordgo.Session
}

func NewDiscordSender(session *discordgo.Session) *DiscordSender {
	return &DiscordSender{
		session: session,
	}
}

func (s *DiscordSender) SendChannel(channelID, msg string) error {
	_, err := s.session.ChannelMessageSend(channelID, msg)
	return err
}

func (s *DiscordSender) SendDM(userID, msg string) error {
	dmChannel, err := s.session.UserChannelCreate(userID)
	if err != nil {
		return err
	}

	if _, err = s.session.ChannelMessageSend(dmChannel.ID, msg); err != nil {
		return err
	}

	return nil
}
