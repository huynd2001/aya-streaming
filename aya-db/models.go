package main

import "gorm.io/gorm"

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
	Id        string
	Discord   string
	Twitch    string
	Youtube   string
	StartTime int
}
