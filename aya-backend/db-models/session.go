package models

import (
	"aya-backend/server-ws/chat_service"
	discordsource "aya-backend/server-ws/chat_service/discord"
	twitchsource "aya-backend/server-ws/chat_service/twitch"
	youtubesource "aya-backend/server-ws/chat_service/youtube"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type GORMSession struct {
	gorm.Model
	UUID      uuid.UUID
	Resources string
	IsOn      bool
	UserID    uint
	User      GORMUser `gorm:"references:ID"`
}

type Resource struct {
	ResourceType chat_service.Source `json:"resourceType"`
	ResourceInfo any                 `json:"resourceInfo"`
}

func (r *Resource) UnmarshalJSON(data []byte) error {
	var resource struct {
		ResourceType    chat_service.Source `json:"resourceType"`
		ResourceInfoAny any                 `json:"resourceInfo"`
	}
	var err error
	if err = json.Unmarshal(data, &resource); err != nil {
		return err
	}

	r.ResourceType = resource.ResourceType
	resourceInfoStr, err := json.Marshal(resource.ResourceInfoAny)
	if err != nil {
		return err
	}

	switch r.ResourceType {
	case chat_service.Discord:
		var discordInfo discordsource.DiscordInfo
		err := json.Unmarshal(resourceInfoStr, &discordInfo)
		if err != nil {
			return fmt.Errorf("resource of type 'discord', but cannot parse the info: %s", err.Error())
		} else {
			r.ResourceInfo = discordInfo
			return nil
		}
	case chat_service.Youtube:
		var youtubeInfo youtubesource.YoutubeInfo
		err := json.Unmarshal(resourceInfoStr, &youtubeInfo)
		if err != nil {
			return fmt.Errorf("resource of type 'youtube', but cannot parse the info: %s", err.Error())
		} else {
			r.ResourceInfo = youtubeInfo
			return nil
		}
	case chat_service.Twitch:
		var twitchInfo twitchsource.TwitchInfo
		err := json.Unmarshal(resourceInfoStr, &twitchInfo)
		if err != nil {
			return fmt.Errorf("resource of type 'twitch', but cannot parse the info: %s", err.Error())
		} else {
			r.ResourceInfo = twitchInfo
			return nil
		}
	default:
		return fmt.Errorf("resource of type '%v' is not supported", r.ResourceType)
	}
}

func (session *GORMSession) BeforeCreate(db *gorm.DB) (err error) {
	session.UUID = uuid.New()
	return
}
