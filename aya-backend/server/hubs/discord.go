package hubs

import (
	discordsource "aya-backend/server/service/discord"
	"fmt"
	"github.com/fatih/color"
	"strings"
	"sync"
)

type DiscordResourceHub struct {
	SessionResourceHub
	mutex                sync.RWMutex
	guildChannel2Session map[string]map[string]bool
	session2GuildChannel map[string]map[string]bool
	emitter              *discordsource.DiscordEmitter
}

func NewDiscordResourceHub(emitter *discordsource.DiscordEmitter) *DiscordResourceHub {
	return &DiscordResourceHub{
		guildChannel2Session: make(map[string]map[string]bool),
		session2GuildChannel: make(map[string]map[string]bool),
		emitter:              emitter,
	}
}

func (hub *DiscordResourceHub) GetSessionId(resourceInfo any) []string {
	hub.mutex.RLock()
	defer hub.mutex.RUnlock()
	discordInfo, ok := resourceInfo.(discordsource.DiscordInfo)
	if !ok {
		return []string{}
	}
	guildId := discordInfo.DiscordGuildId
	channelId := discordInfo.DiscordChannelId
	guildChannel := fmt.Sprintf("%s/%s", guildId, channelId)
	if hub.guildChannel2Session[guildChannel] == nil {
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

func getDiscordGuildChannel(guildChannel string) (guildId string, channelId string) {
	items := strings.Split(guildChannel, "/")
	if len(items) != 2 {
		guildId = ""
		channelId = ""
	} else {
		guildId = items[0]
		channelId = items[1]
	}
	return
}

func diffDiscord(
	oldResources map[string]bool,
	newResources map[string]bool) (
	similarResources []discordsource.DiscordInfo,
	removeResources []discordsource.DiscordInfo,
	addResources []discordsource.DiscordInfo,
) {
	for oldGuildChannel := range oldResources {
		if _, ok := newResources[oldGuildChannel]; !ok {
			guildId, channelId := getDiscordGuildChannel(oldGuildChannel)
			removeResources = append(removeResources, discordsource.DiscordInfo{
				DiscordGuildId:   guildId,
				DiscordChannelId: channelId,
			})
		} else {
			guildId, channelId := getDiscordGuildChannel(oldGuildChannel)
			similarResources = append(similarResources, discordsource.DiscordInfo{
				DiscordGuildId:   guildId,
				DiscordChannelId: channelId,
			})
		}
	}
	for newGuildChannel := range newResources {
		if _, ok := oldResources[newGuildChannel]; !ok {
			guildId, channelId := getDiscordGuildChannel(newGuildChannel)
			addResources = append(addResources, discordsource.DiscordInfo{
				DiscordGuildId:   guildId,
				DiscordChannelId: channelId,
			})
		}
	}
	return
}

func (hub *DiscordResourceHub) AddSession(sessionId string) {
	hub.mutex.Lock()
	defer hub.mutex.Unlock()
}

func (hub *DiscordResourceHub) RegisterSessionResources(sessionId string, resources []discordsource.DiscordInfo) {
	hub.mutex.Lock()
	defer hub.mutex.Unlock()
	var oldResources = hub.session2GuildChannel[sessionId]
	if oldResources == nil {
		oldResources = make(map[string]bool)
	}
	newResources := make(map[string]bool)
	for _, resource := range resources {
		guildChannel := fmt.Sprintf("%s/%s", resource.DiscordGuildId, resource.DiscordChannelId)
		newResources[guildChannel] = true
	}
	similarRs, removeRs, addRs := diffDiscord(oldResources, newResources)
	for _, similarR := range similarRs {
		fmt.Printf("Discord: %s->%#v\n", sessionId, similarR)
	}
	for _, removeR := range removeRs {
		red := color.New(color.FgRed).SprintfFunc()
		fmt.Printf("Discord: %s->%s\n", sessionId, red("-- %#v", removeRs))
		hub.emitter.Deregister(sessionId, removeR)
		hub.deregisterSession(sessionId, removeR)
	}
	for _, addR := range addRs {
		green := color.New(color.FgGreen).SprintfFunc()
		fmt.Printf("Discord: %s->%s\n", sessionId, green("++ %#v", addR))
		hub.emitter.Register(sessionId, addR)
		hub.registerSession(sessionId, addR)
	}
}

func (hub *DiscordResourceHub) registerSession(sessionId string, resourceInfo discordsource.DiscordInfo) {

	guildId := resourceInfo.DiscordGuildId
	channelId := resourceInfo.DiscordChannelId
	guildChannel := fmt.Sprintf("%s/%s", guildId, channelId)

	if hub.session2GuildChannel[sessionId] == nil {
		hub.session2GuildChannel[sessionId] = make(map[string]bool)
	}
	if hub.guildChannel2Session[guildChannel] == nil {
		hub.guildChannel2Session[guildChannel] = make(map[string]bool)
	}
	hub.session2GuildChannel[sessionId][guildChannel] = true
	hub.guildChannel2Session[guildChannel][sessionId] = true
}

func (hub *DiscordResourceHub) deregisterSession(sessionId string, resourceInfo discordsource.DiscordInfo) {
	guildId := resourceInfo.DiscordGuildId
	channelId := resourceInfo.DiscordChannelId
	guildChannel := fmt.Sprintf("%s/%s", guildId, channelId)

	if hub.session2GuildChannel[sessionId] != nil {
		delete(hub.session2GuildChannel[sessionId], guildChannel)
	}
	if hub.guildChannel2Session[guildChannel] != nil {
		delete(hub.guildChannel2Session[guildChannel], sessionId)
	}
}
