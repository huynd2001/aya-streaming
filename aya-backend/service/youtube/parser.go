package youtube_source

import (
	. "aya-backend/service"
	yt "google.golang.org/api/youtube/v3"
)

type YoutubeMessageParser struct {
}

func (parser *YoutubeMessageParser) ParseAuthor(authorDetails *yt.LiveChatMessageAuthorDetails) Author {
	return Author{
		Username: authorDetails.DisplayName,
		IsAdmin:  authorDetails.IsChatModerator,
		IsBot:    false,
		Color:    "#ffffff",
	}
}

func (parser *YoutubeMessageParser) ParseMessage(msg *yt.LiveChatMessage) Message {
	return Message{
		Source: Youtube,
		Id:     msg.Id,
		Author: parser.ParseAuthor(msg.AuthorDetails),
		MessageParts: []MessagePart{
			{
				Content: msg.Snippet.DisplayMessage,
			},
		},
	}
}
