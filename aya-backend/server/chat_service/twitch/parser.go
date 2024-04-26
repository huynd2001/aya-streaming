package twitch_source

import (
	"aya-backend/server/chat_service"
	"github.com/gempir/go-twitch-irc/v4"
)

type TwitchMessageParser struct {
}

func (parser *TwitchMessageParser) ParseMessage(twitchMsg twitch.PrivateMessage) chat_service.MessageUpdate {
	return chat_service.MessageUpdate{}
}
