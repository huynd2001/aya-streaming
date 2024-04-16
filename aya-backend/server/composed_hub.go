package main

import (
	"aya-backend/server/hubs"
	"aya-backend/server/service"
)

type MessageHub struct {
	hubs.SessionResourceHub
	discordHub *hubs.DiscordResourceHub
	youtubeHub *hubs.YoutubeResourceHub
}

type HubResourceInfo struct {
	ResourceType service.Source
	SpecificInfo any
}

func (m *MessageHub) GetSessionId(resourceInfo any) []string {
	hubResourceInfo, ok := resourceInfo.(HubResourceInfo)
	if !ok {
		return []string{}
	}
	switch hubResourceInfo.ResourceType {
	case service.Discord:
		return m.discordHub.GetSessionId(hubResourceInfo.SpecificInfo)
	case service.Youtube:
		return m.youtubeHub.GetSessionId(hubResourceInfo.SpecificInfo)
	default:
		return []string{}
	}
}

func (m *MessageHub) RemoveSession(sessionId string) {
	m.discordHub.RemoveSession(sessionId)
	m.youtubeHub.RemoveSession(sessionId)
}

func (m *MessageHub) AddSession(sessionId string, resourceInfo any) {
	hubResourceInfo, ok := resourceInfo.(HubResourceInfo)
	if !ok {
		return
	}
	switch hubResourceInfo.ResourceType {
	case service.Discord:
		m.discordHub.AddSession(sessionId, hubResourceInfo.SpecificInfo)
	case service.Youtube:
		m.youtubeHub.AddSession(sessionId, hubResourceInfo.SpecificInfo)
	default:
		return
	}
}
