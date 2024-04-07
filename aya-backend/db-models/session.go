package models

import (
	"gorm.io/gorm"
	"time"
)

type GORMSession struct {
	gorm.Model
	ID       uint
	Discord  string
	Youtube  string
	CreateAt time.Time
	UpdateAt time.Time
	IsOn     bool
	IsDelete bool
	UserID   uint
	User     GORMUser `gorm:"references:ID"`
}

type Discord struct {
	GuildId   string `json:"guildId,omitempty"`
	ChannelId string `json:"channelId,omitempty"`
}

type Twitch struct {
}

type Youtube struct {
}
