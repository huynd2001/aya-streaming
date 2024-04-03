package db_models

import (
	"gorm.io/gorm"
	"time"
)

type Discord struct {
	guildId   string
	channelId string
}

type Twitch struct {
}

type Youtube struct {
}

type Session struct {
	gorm.Model
	Id         string
	Discord    string
	Twitch     string
	Youtube    string
	CreateTime time.Time
	UpdateTime time.Time
	EndTime    time.Time
}
