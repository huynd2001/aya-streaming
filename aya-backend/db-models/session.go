package db_models

import (
	"gorm.io/gorm"
	"time"
)

type GORMSession struct {
	gorm.Model
	Id         uint `gorm:"primaryKey"`
	Discord    string
	Twitch     string
	Youtube    string
	CreateTime time.Time
	UpdateTime time.Time
	IsOn       bool
	IsDelete   bool
}

type Session struct {
	Id         uint      `json:"id"`
	Discord    *Discord  `json:"discord,omitempty"`
	Twitch     *Twitch   `json:"twitch,omitempty"`
	Youtube    *Youtube  `json:"youtube,omitempty"`
	CreateTime time.Time `json:"createTime"`
	UpdateTime time.Time `json:"UpdateTime"`
	IsOn       bool      `json:"isOn"`
	IsDelete   bool      `json:"isDelete"`
}

type Discord struct {
	GuildId   string `json:"guildId,omitempty"`
	ChannelId string `json:"channelId,omitempty"`
}

type Twitch struct {
}

type Youtube struct {
}
