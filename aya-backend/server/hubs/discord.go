package hubs

import (
	discordsource "aya-backend/server/service/discord"
	"strings"
	"sync"
)

type DiscordResourceHub struct {
	SessionResourceHub
	mutex                sync.RWMutex
	guildChannel2Session map[string]map[string]bool
	session2GuildChannel map[string]map[string]bool
}

func NewDiscordResourceHub() *DiscordResourceHub {
	return &DiscordResourceHub{
		guildChannel2Session: make(map[string]map[string]bool),
		session2GuildChannel: make(map[string]map[string]bool),
	}
}

func (hub *DiscordResourceHub) GetSessionId(resourceInfo any) []string {
	hub.mutex.RLock()
	defer hub.mutex.RUnlock()
	discordInfo, ok := resourceInfo.(discordsource.DiscordSpecificInfo)
	if !ok {
		return []string{}
	}
	guildId := discordInfo.DiscordGuildId
	channelId := discordInfo.DiscordChannelId
	guildChannel := strings.Join([]string{guildId, channelId}, "/")
	if hub.guildChannel2Session[guildId] == nil {
		return []string{}
	}
	sessions := make([]string, len(hub.guildChannel2Session[guildChannel]))
	idx := 0
	for sessionId := range hub.guildChannel2Session[guildChannel] {
		sessions[idx] = sessionId
	}
	return sessions
}

func (hub *DiscordResourceHub) RemoveSession(sessionId string) {
	hub.mutex.Lock()
	defer hub.mutex.Unlock()
	if hub.session2GuildChannel[sessionId] == nil {
		// Do not have to do anything
		return
	}
	for guildChannel := range hub.session2GuildChannel[sessionId] {
		delete(hub.guildChannel2Session[guildChannel], sessionId)
	}
	delete(hub.session2GuildChannel, sessionId)

}

func (hub *DiscordResourceHub) AddSession(sessionId string, resourceInfo any) {
	hub.mutex.Lock()
	defer hub.mutex.Unlock()
	discordInfo, ok := resourceInfo.(discordsource.DiscordSpecificInfo)
	if !ok {
		// do nothing
		return
	}
	guildId := discordInfo.DiscordGuildId
	channelId := discordInfo.DiscordChannelId
	guildChannel := strings.Join([]string{guildId, channelId}, "/")

	if hub.session2GuildChannel[sessionId] == nil {
		hub.session2GuildChannel[sessionId] = make(map[string]bool)
	}
	if hub.guildChannel2Session[guildChannel] == nil {
		hub.guildChannel2Session[guildChannel] = make(map[string]bool)
	}
	hub.session2GuildChannel[sessionId][guildChannel] = true
	hub.guildChannel2Session[guildChannel][sessionId] = true
}
