package hubs

import (
	models "aya-backend/db-models"
	"aya-backend/server/db"
	"aya-backend/server/service"
	"aya-backend/server/service/composed"
	discordsource "aya-backend/server/service/discord"
	youtubesource "aya-backend/server/service/youtube"
	"gorm.io/gorm"
	"sync"
	"time"
)

const (
	DATA_RETRIEVAL_INTERVAL = 1 * time.Minute
)

type MessageHub struct {
	SessionResourceHub
	mutex      sync.RWMutex
	discordHub *DiscordResourceHub
	youtubeHub *YoutubeResourceHub

	infoDB *db.InfoDB

	stopSignal chan bool

	registeredSessions map[string]bool
}

func NewMessageHub(emitter *composed.MessageEmitter, gormDB *gorm.DB) *MessageHub {

	msgHub := MessageHub{
		discordHub: NewDiscordResourceHub(emitter.GetDiscordEmitter()),
		youtubeHub: NewYoutubeResourceHub(emitter.GetYoutubeEmitter()),
		infoDB:     db.NewInfoDB(gormDB),
		stopSignal: make(chan bool),
	}

	go func() {
		lastUpdateTime := time.Now()
		for {
			select {
			case <-time.After(DATA_RETRIEVAL_INTERVAL):
				newTime := time.Now()
				resourceInfoMap := msgHub.infoDB.GetResourcesInfo(msgHub.registeredSessions, lastUpdateTime)
				for sessionId, resources := range resourceInfoMap {
					msgHub.RegisterSessionResources(sessionId, resources)
				}
				lastUpdateTime = newTime
			case <-msgHub.stopSignal:
				return
			}
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
		return m.discordHub.GetSessionId(hubResourceInfo.ResourceInfo)
	case service.Youtube:
		return m.youtubeHub.GetSessionId(hubResourceInfo.ResourceInfo)
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
	m.discordHub.RegisterSessionResources(sessionId, discordResources)
	m.youtubeHub.RegisterSessionResources(sessionId, youtubeResources)

}

func (m *MessageHub) AddSession(sessionId string) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.registeredSessions[sessionId] = false
	m.discordHub.AddSession(sessionId)
	m.youtubeHub.AddSession(sessionId)
}
