package hubs

import (
	models "aya-backend/db-models"
	"aya-backend/server/db"
	"aya-backend/server/service"
	"aya-backend/server/service/composed"
	discordsource "aya-backend/server/service/discord"
	youtubesource "aya-backend/server/service/youtube"
	"fmt"
	"gorm.io/gorm"
	"sync"
	"time"
)

const (
	DATA_RETRIEVAL_INTERVAL = 10 * time.Second
)

type MessageHub struct {
	SessionResourceHub
	mutex      sync.RWMutex
	discordHub *DiscordResourceHub
	youtubeHub *YoutubeResourceHub

	infoDB *db.InfoDB

	registeredSessions map[string]bool
}

func NewMessageHub(emitter *composed.MessageEmitter, gormDB *gorm.DB) *MessageHub {

	msgHub := MessageHub{
		discordHub:         NewDiscordResourceHub(emitter.GetDiscordEmitter()),
		youtubeHub:         NewYoutubeResourceHub(emitter.GetYoutubeEmitter()),
		infoDB:             db.NewInfoDB(gormDB),
		registeredSessions: make(map[string]bool),
	}

	go func() {
		lastUpdateTime := time.Now()
		for {
			<-time.After(DATA_RETRIEVAL_INTERVAL)
			newTime := time.Now()
			resourceInfoMap := msgHub.infoDB.GetResourcesInfo(msgHub.registeredSessions, lastUpdateTime)
			if len(resourceInfoMap) > 0 {
				fmt.Println("Changes detected:")
			}
			for sessionId, resources := range resourceInfoMap {
				fmt.Printf("Update session with Id %s\n", sessionId)
				fmt.Printf("New resources info: %s\n", resources)
				msgHub.RegisterSessionResources(sessionId, resources)
			}
			lastUpdateTime = newTime
		}
	}()

	return &msgHub
}

func (m *MessageHub) GetSessionId(resourceInfo any) []string {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	hubResourceInfo, ok := resourceInfo.(models.Resource)
	if !ok {
		return []string{}
	}
	switch hubResourceInfo.ResourceType {
	case service.Discord:
		if m.discordHub != nil {
			return m.discordHub.GetSessionId(hubResourceInfo.ResourceInfo)
		} else {
			return []string{}
		}
	case service.Youtube:
		if m.youtubeHub != nil {
			return m.youtubeHub.GetSessionId(hubResourceInfo.ResourceInfo)
		} else {
			return []string{}
		}
	default:
		return []string{}
	}
}

func (m *MessageHub) RemoveSession(sessionId string) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	delete(m.registeredSessions, sessionId)
	m.discordHub.RemoveSession(sessionId)
	m.youtubeHub.RemoveSession(sessionId)
}

func (m *MessageHub) RegisterSessionResources(sessionId string, resources []models.Resource) {
	var discordResources []discordsource.DiscordInfo
	var youtubeResources []youtubesource.YoutubeInfo
	for _, resource := range resources {
		switch resource.ResourceType {
		case service.Discord:
			discordResource, ok := resource.ResourceInfo.(discordsource.DiscordInfo)
			if ok {
				discordResources = append(discordResources, discordResource)
			}
		case service.Youtube:
			youtubeResource, ok := resource.ResourceInfo.(youtubesource.YoutubeInfo)
			if ok {
				youtubeResources = append(youtubeResources, youtubeResource)
			}
		default:
		}
	}
	if m.discordHub != nil {
		m.discordHub.RegisterSessionResources(sessionId, discordResources)
	}
	if m.youtubeHub != nil {
		m.youtubeHub.RegisterSessionResources(sessionId, youtubeResources)
	}
	m.registeredSessions[sessionId] = true

}

func (m *MessageHub) AddSession(sessionId string) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.registeredSessions[sessionId] = false
	if m.discordHub != nil {
		m.discordHub.AddSession(sessionId)
	}
	if m.youtubeHub != nil {
		m.youtubeHub.AddSession(sessionId)
	}
}
