package models

import (
	"aya-backend/server/service"
	discordsource "aya-backend/server/service/discord"
	youtubesource "aya-backend/server/service/youtube"
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
	ResourceType service.Source `json:"resourceType"`
	ResourceInfo any            `json:"resourceInfo"`
}

func (r *Resource) UnmarshalJSON(data []byte) error {
	var resource struct {
		ResourceType    service.Source `json:"resourceType"`
		ResourceInfoStr string         `json:"resourceInfo"`
	}
	var err error
	if err = json.Unmarshal(data, &resource); err != nil {
		return err
	}

	r.ResourceType = resource.ResourceType

	switch r.ResourceType {
	case service.Discord:
		var discordInfo discordsource.DiscordInfo
		err := json.Unmarshal([]byte(resource.ResourceInfoStr), &discordInfo)
		if err != nil {
			return fmt.Errorf("resource of type 'discord', but cannot parse the info: %s", err.Error())
		} else {
			r.ResourceInfo = discordInfo
			return nil
		}
	case service.Youtube:
		var youtubeInfo youtubesource.YoutubeInfo
		err := json.Unmarshal([]byte(resource.ResourceInfoStr), &youtubeInfo)
		if err != nil {
			return fmt.Errorf("resource of type 'youtube', but cannot parse the info: %s", err.Error())
		} else {
			r.ResourceInfo = youtubeInfo
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
