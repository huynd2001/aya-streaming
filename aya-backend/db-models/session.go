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
	var source string
	var err error
	if err = json.Unmarshal(data, &source); err != nil {
		return err
	}
	switch r.ResourceType {
	case service.Discord:
		_, ok := r.ResourceInfo.(discordsource.DiscordInfo)
		if !ok {
			return fmt.Errorf("resource of type 'discord', but cannot parse the info")
		} else {
			return nil
		}
	case service.Youtube:
		_, ok := r.ResourceInfo.(youtubesource.YoutubeInfo)
		if !ok {
			return fmt.Errorf("resource of type 'youtube', but cannot parse the info")
		} else {
			return nil
		}
	default:
		return fmt.Errorf("resource of type '%v' is not supported", r.ResourceType)
	}
}

func (session *GORMSession) BeforeCreate(tx *gorm.DB) (err error) {
	session.UUID = uuid.New()
	return
}
