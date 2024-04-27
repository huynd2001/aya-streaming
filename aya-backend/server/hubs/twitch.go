package hubs

import (
	twitch_source "aya-backend/server/chat_service/twitch"
	youtubesource "aya-backend/server/chat_service/youtube"
	"fmt"
	"github.com/fatih/color"
	"sync"
)

type TwitchResourceHub struct {
	SessionResourceHub
	mutex               sync.RWMutex
	channelName2Session map[string]map[string]bool
	session2ChannelName map[string]map[string]bool
	emitter             *twitch_source.TwitchEmitter
}

func NewTwitchResourceHub(emitter *twitch_source.TwitchEmitter) *TwitchResourceHub {
	return &TwitchResourceHub{
		channelName2Session: make(map[string]map[string]bool),
		session2ChannelName: make(map[string]map[string]bool),
		emitter:             emitter,
	}
}

func (hub *TwitchResourceHub) GetSessionId(resourceInfo any) []string {
	hub.mutex.RLock()
	defer hub.mutex.RUnlock()
	twitchInfo, ok := resourceInfo.(twitch_source.TwitchInfo)
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
	similarResources []twitch_source.TwitchInfo,
	removeResources []twitch_source.TwitchInfo,
	addResources []twitch_source.TwitchInfo,
) {
	for oldChannelName := range oldResources {
		if _, ok := newResources[oldChannelName]; !ok {
			removeResources = append(removeResources, twitch_source.TwitchInfo{
				TwitchChannelName: oldChannelName,
			})
		} else {
			similarResources = append(similarResources, twitch_source.TwitchInfo{
				TwitchChannelName: oldChannelName,
			})
		}
	}
	for newChannelName := range newResources {
		if _, ok := oldResources[newChannelName]; !ok {
			addResources = append(addResources, twitch_source.TwitchInfo{
				TwitchChannelName: newChannelName,
			})
		}
	}
	return
}

func (hub *TwitchResourceHub) RegisterSessionResources(sessionId string, resources []twitch_source.TwitchInfo) {
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
	similarRs, removeRs, addRs := diffYoutube(oldResources, newResources)
	for _, similarR := range similarRs {
		fmt.Printf("Youtube: %s->%#v\n", sessionId, similarR)
	}
	for _, removeR := range removeRs {
		red := color.New(color.FgRed).SprintfFunc()
		fmt.Printf("Youtube: %s->%s\n", sessionId, red("-- %#v", removeRs))
		hub.emitter.Deregister(sessionId, removeR)
		hub.deregisterSession(sessionId, removeR)
	}
	for _, addR := range addRs {
		green := color.New(color.FgGreen).SprintfFunc()
		fmt.Printf("Youtube: %s->%s\n", sessionId, green("++ %#v", addR))
		hub.emitter.Register(sessionId, addR)
		hub.registerSession(sessionId, addR)
	}
}

func (hub *TwitchResourceHub) registerSession(sessionId string, resourceInfo youtubesource.YoutubeInfo) {

	youtubeChannelId := resourceInfo.YoutubeChannelId
	if hub.session2ChannelName[sessionId] == nil {
		hub.session2ChannelName[sessionId] = make(map[string]bool)
	}
	if hub.channelName2Session[youtubeChannelId] == nil {
		hub.channelName2Session[youtubeChannelId] = make(map[string]bool)
	}
	hub.session2ChannelName[sessionId][youtubeChannelId] = true
	hub.channelName2Session[youtubeChannelId][sessionId] = true

}

func (hub *TwitchResourceHub) deregisterSession(sessionId string, resourceInfo youtubesource.YoutubeInfo) {
	youtubeChannelId := resourceInfo.YoutubeChannelId
	if hub.session2ChannelName[sessionId] != nil {
		delete(hub.session2ChannelName[sessionId], youtubeChannelId)
	}
	if hub.channelName2Session[youtubeChannelId] != nil {
		delete(hub.channelName2Session[youtubeChannelId], sessionId)
	}
}
