package hubs

import (
	youtubesource "aya-backend/server/service/youtube"
	"sync"
)

type YoutubeResourceHub struct {
	SessionResourceHub
	mutex           sync.RWMutex
	channel2Session map[string]map[string]bool
	session2Channel map[string]map[string]bool
}

func NewYoutubeResourceHub() *YoutubeResourceHub {
	return &YoutubeResourceHub{
		channel2Session: make(map[string]map[string]bool),
		session2Channel: make(map[string]map[string]bool),
	}
}

func (hub *YoutubeResourceHub) GetSessionId(resourceInfo any) []string {
	hub.mutex.RLock()
	defer hub.mutex.RUnlock()
	ytResourceInfo, ok := resourceInfo.(youtubesource.YoutubeInfo)
	if !ok {
		return []string{}
	}
	youtubeChannelId := ytResourceInfo.YoutubeChannelId
	if hub.channel2Session[youtubeChannelId] == nil {
		return []string{}
	}
	sessions := make([]string, len(hub.channel2Session[youtubeChannelId]))
	idx := 0
	for sessionId := range hub.channel2Session[youtubeChannelId] {
		sessions[idx] = sessionId
		idx++
	}
	return sessions
}

func (hub *YoutubeResourceHub) RemoveSession(sessionId string) {
	hub.mutex.Lock()
	defer hub.mutex.Unlock()
	if hub.session2Channel[sessionId] == nil {
		// Do not have to do anything
		return
	}
	for channelId := range hub.session2Channel[sessionId] {
		delete(hub.channel2Session[channelId], sessionId)
	}
	delete(hub.session2Channel, sessionId)
	return
}

func (hub *YoutubeResourceHub) AddSession(sessionId string, resourceInfo any) {
	hub.mutex.Lock()
	defer hub.mutex.Unlock()
	ytResourceInfo, ok := resourceInfo.(youtubesource.YoutubeInfo)
	if !ok {
		// do nothing
		return
	}
	youtubeChannelId := ytResourceInfo.YoutubeChannelId
	if hub.session2Channel[sessionId] == nil {
		hub.session2Channel[sessionId] = make(map[string]bool)
	}
	if hub.channel2Session[youtubeChannelId] == nil {
		hub.channel2Session[youtubeChannelId] = make(map[string]bool)
	}
	hub.session2Channel[sessionId][youtubeChannelId] = true
	hub.channel2Session[youtubeChannelId][sessionId] = true

}
