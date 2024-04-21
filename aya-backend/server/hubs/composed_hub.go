package hubs

import (
	models "aya-backend/db-models"
	"aya-backend/server/service"
)

type MessageHub struct {
	SessionResourceHub
	discordHub *DiscordResourceHub
	youtubeHub *YoutubeResourceHub
}

func NewMessageHub() *MessageHub {
	return &MessageHub{
		discordHub: NewDiscordResourceHub(),
		youtubeHub: NewYoutubeResourceHub(),
	}
}

func (m *MessageHub) GetSessionId(resourceInfo any) []string {
	hubResourceInfo, ok := resourceInfo.(models.Resource)
	if !ok {
		return []string{}
	}
	switch hubResourceInfo.ResourceType {
	case service.Discord:
		return m.discordHub.GetSessionId(hubResourceInfo.ResourceInfo)
	case service.Youtube:
		return m.youtubeHub.GetSessionId(hubResourceInfo.ResourceInfo)
	default:
		return []string{}
	}
}

func (m *MessageHub) RemoveSession(sessionId string) {
	m.discordHub.RemoveSession(sessionId)
	m.youtubeHub.RemoveSession(sessionId)
}

func (m *MessageHub) AddSession(sessionId string, resourceInfo any) {
	hubResourceInfo, ok := resourceInfo.(models.Resource)
	if !ok {
		return
	}
	switch hubResourceInfo.ResourceType {
	case service.Discord:
		m.discordHub.AddSession(sessionId, hubResourceInfo.ResourceInfo)
	case service.Youtube:
		m.youtubeHub.AddSession(sessionId, hubResourceInfo.ResourceInfo)
	default:
		return
	}
}
