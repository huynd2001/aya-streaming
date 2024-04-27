package hubs

import (
	twitchsource "aya-backend/server-ws/chat_service/twitch"
	"fmt"
	"github.com/fatih/color"
	"sync"
)

type TwitchResourceHub struct {
	SessionResourceHub
	mutex               sync.RWMutex
	channelName2Session map[string]map[string]bool
	session2ChannelName map[string]map[string]bool
	emitter             *twitchsource.TwitchEmitter
}

func NewTwitchResourceHub(emitter *twitchsource.TwitchEmitter) *TwitchResourceHub {
	return &TwitchResourceHub{
		channelName2Session: make(map[string]map[string]bool),
		session2ChannelName: make(map[string]map[string]bool),
		emitter:             emitter,
	}
}

func (hub *TwitchResourceHub) GetSessionId(resourceInfo any) []string {
	hub.mutex.RLock()
	defer hub.mutex.RUnlock()
	twitchInfo, ok := resourceInfo.(twitchsource.TwitchInfo)
	if !ok {
		return []string{}
	}
	youtubeChannelId := twitchInfo.TwitchChannelName
	if hub.channelName2Session[youtubeChannelId] == nil {
		return []string{}
	}
	sessions := make([]string, len(hub.channelName2Session[youtubeChannelId]))
	idx := 0
	for sessionId := range hub.channelName2Session[youtubeChannelId] {
		sessions[idx] = sessionId
		idx++
	}
	return sessions

}

func (hub *TwitchResourceHub) RemoveSession(sessionId string) {
	hub.mutex.Lock()
	defer hub.mutex.Unlock()
	if hub.session2ChannelName[sessionId] == nil {
		// Do not have to do anything
		return
	}
	for channelId := range hub.session2ChannelName[sessionId] {
		delete(hub.channelName2Session[channelId], sessionId)
	}
	delete(hub.session2ChannelName, sessionId)
}

func (hub *TwitchResourceHub) AddSession(sessionId string) {
	hub.mutex.Lock()
	defer hub.mutex.Unlock()
}

func diffTwitch(
	oldResources map[string]bool,
	newResources map[string]bool) (
	similarResources []twitchsource.TwitchInfo,
	removeResources []twitchsource.TwitchInfo,
	addResources []twitchsource.TwitchInfo,
) {
	for oldChannelName := range oldResources {
		if _, ok := newResources[oldChannelName]; !ok {
			removeResources = append(removeResources, twitchsource.TwitchInfo{
				TwitchChannelName: oldChannelName,
			})
		} else {
			similarResources = append(similarResources, twitchsource.TwitchInfo{
				TwitchChannelName: oldChannelName,
			})
		}
	}
	for newChannelName := range newResources {
		if _, ok := oldResources[newChannelName]; !ok {
			addResources = append(addResources, twitchsource.TwitchInfo{
				TwitchChannelName: newChannelName,
			})
		}
	}
	return
}

func (hub *TwitchResourceHub) RegisterSessionResources(sessionId string, resources []twitchsource.TwitchInfo) {
	hub.mutex.Lock()
	defer hub.mutex.Unlock()
	// get the current resources attacked to this session
	var oldResources = hub.session2ChannelName[sessionId]
	if oldResources == nil {
		oldResources = make(map[string]bool)
	}
	newResources := make(map[string]bool)
	for _, resourceInfo := range resources {
		newResources[resourceInfo.TwitchChannelName] = true
	}
	similarRs, removeRs, addRs := diffTwitch(oldResources, newResources)
	for _, similarR := range similarRs {
		fmt.Printf("Twitch: %s->%#v\n", sessionId, similarR)
	}
	for _, removeR := range removeRs {
		red := color.New(color.FgRed).SprintfFunc()
		fmt.Printf("Twitch: %s->%s\n", sessionId, red("-- %#v", removeR))
		hub.emitter.Deregister(sessionId, removeR)
		hub.deregisterSession(sessionId, removeR)
	}
	for _, addR := range addRs {
		green := color.New(color.FgGreen).SprintfFunc()
		fmt.Printf("Twitch: %s->%s\n", sessionId, green("++ %#v", addR))
		hub.emitter.Register(sessionId, addR)
		hub.registerSession(sessionId, addR)
	}
}

func (hub *TwitchResourceHub) registerSession(sessionId string, resourceInfo twitchsource.TwitchInfo) {

	twitchChannelName := resourceInfo.TwitchChannelName
	if hub.session2ChannelName[sessionId] == nil {
		hub.session2ChannelName[sessionId] = make(map[string]bool)
	}
	if hub.channelName2Session[twitchChannelName] == nil {
		hub.channelName2Session[twitchChannelName] = make(map[string]bool)
	}
	hub.session2ChannelName[sessionId][twitchChannelName] = true
	hub.channelName2Session[twitchChannelName][sessionId] = true

}

func (hub *TwitchResourceHub) deregisterSession(sessionId string, resourceInfo twitchsource.TwitchInfo) {
	twitchChannelName := resourceInfo.TwitchChannelName
	if hub.session2ChannelName[sessionId] != nil {
		delete(hub.session2ChannelName[sessionId], twitchChannelName)
	}
	if hub.channelName2Session[twitchChannelName] != nil {
		delete(hub.channelName2Session[twitchChannelName], sessionId)
	}
}
