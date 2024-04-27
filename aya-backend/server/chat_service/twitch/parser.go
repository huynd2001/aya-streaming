package twitch_source

import (
	"aya-backend/server/chat_service"
	"github.com/fatih/color"
	"github.com/gempir/go-twitch-irc/v4"
)

type TwitchMessageParser struct {
}

func (parser *TwitchMessageParser) ParseMessage(twitchMsg twitch.PrivateMessage) chat_service.MessageUpdate {
	color.Cyan("%#v\n", twitchMsg)
	return chat_service.MessageUpdate{
		UpdateTime: twitchMsg.Time,
		Update:     chat_service.New,
		Message: chat_service.Message{
			Source: chat_service.Twitch,
			Id:     twitchMsg.ID,
			Author: chat_service.Author{
				Username: twitchMsg.User.Name,
			},
			MessageParts: []chat_service.MessagePart{
				{
					Content: twitchMsg.Message,
				},
			},
		},
		ExtraFields: TwitchInfo{
			TwitchChannelName: twitchMsg.Channel,
		},
	}
}
