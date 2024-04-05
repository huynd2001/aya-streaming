package models

import (
	"encoding/json"
	"gorm.io/gorm"
	"time"
)

type GORMSession struct {
	gorm.Model
	ID       uint
	Discord  string
	Twitch   string
	Youtube  string
	CreateAt time.Time
	UpdateAt time.Time
	IsOn     bool
	IsDelete bool
	OwnerID  uint
	User     User `gorm:"references:Id"`
}

type Session struct {
	ID       uint      `json:"id"`
	Discord  *Discord  `json:"discord,omitempty"`
	Twitch   *Twitch   `json:"twitch,omitempty"`
	Youtube  *Youtube  `json:"youtube,omitempty"`
	CreateAt time.Time `json:"createTime"`
	UpdateAt time.Time `json:"updateTime"`
	IsOn     bool      `json:"isOn"`
	IsDelete bool      `json:"isDelete"`
	OwnerID  uint      `json:"ownerId"`
}

type Discord struct {
	GuildId   string `json:"guildId,omitempty"`
	ChannelId string `json:"channelId,omitempty"`
}

type Twitch struct {
}

type Youtube struct {
}

func convertGormToSession(gS GORMSession) Session {
	var err error

	var discord *Discord
	err = json.Unmarshal([]byte(gS.Discord), discord)
	if err != nil {
		discord = nil
	}

	var twitch *Twitch
	err = json.Unmarshal([]byte(gS.Twitch), twitch)
	if err != nil {
		twitch = nil
	}

	var youtube *Youtube
	err = json.Unmarshal([]byte(gS.Youtube), youtube)
	if err != nil {
		youtube = nil
	}

	return Session{
		ID:       gS.ID,
		CreateAt: gS.CreateAt,
		UpdateAt: gS.UpdateAt,
		IsDelete: gS.IsDelete,
		IsOn:     gS.IsOn,
		Discord:  discord,
		Youtube:  youtube,
		Twitch:   twitch,
		OwnerID:  gS.OwnerID,
	}

}

func convertSessionToGorm(s Session) GORMSession {
	discordBytes, err := json.Marshal(&s.Discord)
	var discordStr string
	if err != nil {
		discordStr = "{}"
	} else {
		discordStr = string(discordBytes)
	}

	youtubeBytes, err := json.Marshal(&s.Youtube)
	var youtubeStr string
	if err != nil {
		youtubeStr = "{}"
	} else {
		youtubeStr = string(youtubeBytes)
	}

	twitchBytes, err := json.Marshal(&s.Twitch)
	var twitchStr string
	if err != nil {
		twitchStr = "{}"
	} else {
		twitchStr = string(twitchBytes)
	}

	return GORMSession{
		ID:       s.ID,
		CreateAt: s.CreateAt,
		UpdateAt: s.UpdateAt,
		IsDelete: s.IsDelete,
		IsOn:     s.IsOn,
		Discord:  discordStr,
		Youtube:  youtubeStr,
		Twitch:   twitchStr,
		OwnerID:  s.OwnerID,
	}
}
