package hubs

import (
	youtubesource "aya-backend/server/service/youtube"
	"fmt"
	"github.com/fatih/color"
	"sync"
)

type YoutubeResourceHub struct {
	SessionResourceHub
	mutex           sync.RWMutex
	channel2Session map[string]map[string]bool
	session2Channel map[string]map[string]bool
	emitter         *youtubesource.YoutubeEmitter
}

func NewYoutubeResourceHub(emitter *youtubesource.YoutubeEmitter) *YoutubeResourceHub {
	return &YoutubeResourceHub{
		channel2Session: make(map[string]map[string]bool),
		session2Channel: make(map[string]map[string]bool),
		emitter:         emitter,
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
}

func (hub *YoutubeResourceHub) AddSession(sessionId string) {
	hub.mutex.Lock()
	defer hub.mutex.Unlock()
}

func diffYoutube(
	oldResources map[string]bool,
	newResources map[string]bool) (
	similarResources []youtubesource.YoutubeInfo,
	removeResources []youtubesource.YoutubeInfo,
	addResources []youtubesource.YoutubeInfo,
) {
	for oldChannelId := range oldResources {
		if _, ok := newResources[oldChannelId]; !ok {
			removeResources = append(removeResources, youtubesource.YoutubeInfo{
				YoutubeChannelId: oldChannelId,
			})
		} else {
			similarResources = append(similarResources, youtubesource.YoutubeInfo{
				YoutubeChannelId: oldChannelId,
			})
		}
	}
	for newChannelId := range newResources {
		if _, ok := oldResources[newChannelId]; !ok {
			addResources = append(addResources, youtubesource.YoutubeInfo{
				YoutubeChannelId: newChannelId,
			})
		}
	}
	return
}

func (hub *YoutubeResourceHub) RegisterSessionResources(sessionId string, resources []youtubesource.YoutubeInfo) {
	hub.mutex.Lock()
	defer hub.mutex.Unlock()
	// get the current resources attacked to this session
	var oldResources = hub.session2Channel[sessionId]
	if oldResources == nil {
		oldResources = make(map[string]bool)
	}
	newResources := make(map[string]bool)
	for _, resourceInfo := range resources {
		newResources[resourceInfo.YoutubeChannelId] = true
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

func (hub *YoutubeResourceHub) registerSession(sessionId string, resourceInfo youtubesource.YoutubeInfo) {

	youtubeChannelId := resourceInfo.YoutubeChannelId
	if hub.session2Channel[sessionId] == nil {
		hub.session2Channel[sessionId] = make(map[string]bool)
	}
	if hub.channel2Session[youtubeChannelId] == nil {
		hub.channel2Session[youtubeChannelId] = make(map[string]bool)
	}
	hub.session2Channel[sessionId][youtubeChannelId] = true
	hub.channel2Session[youtubeChannelId][sessionId] = true

}

func (hub *YoutubeResourceHub) deregisterSession(sessionId string, resourceInfo youtubesource.YoutubeInfo) {
	youtubeChannelId := resourceInfo.YoutubeChannelId
	if hub.session2Channel[sessionId] != nil {
		delete(hub.session2Channel[sessionId], youtubeChannelId)
	}
	if hub.channel2Session[youtubeChannelId] != nil {
		delete(hub.channel2Session[youtubeChannelId], sessionId)
	}
}
