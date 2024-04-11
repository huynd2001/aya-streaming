package models

import (
	"gorm.io/gorm"
)

type GORMSession struct {
	gorm.Model
	ID        uint
	Resources string
	IsOn      bool
	UserID    uint
	User      GORMUser `gorm:"references:ID"`
}

type Discord struct {
	GuildId   string `json:"guild_id,omitempty"`
	ChannelId string `json:"channel_id,omitempty"`
}

type Twitch struct {
}

type Youtube struct {
	ChannelId string `json:"channel_id,omitempty"`
}
