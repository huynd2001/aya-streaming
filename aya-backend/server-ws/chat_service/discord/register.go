package discord_source

import (
	"fmt"
	"sync"
)

type discordRegister struct {
	mutex           sync.RWMutex
	guildChannelMap map[string]bool
}

func newDiscordRegister() *discordRegister {
	return &discordRegister{
		guildChannelMap: make(map[string]bool),
	}
}

func (register *discordRegister) register(guildId string, channelId string) {
	register.mutex.Lock()
	defer register.mutex.Unlock()
	guildChannel := fmt.Sprintf("%s/%s", guildId, channelId)
	if _, ok := register.guildChannelMap[guildChannel]; ok {
		// guildChannel already exists in the register
		return
	}
	register.guildChannelMap[guildChannel] = true
}

func (register *discordRegister) deregister(guildId string, channelId string) {
	register.mutex.Lock()
	defer register.mutex.Unlock()
	guildChannel := fmt.Sprintf("%s/%s", guildId, channelId)
	delete(register.guildChannelMap, guildChannel)
}

func (register *discordRegister) check(guildId string, channelId string) bool {
	register.mutex.RLock()
	defer register.mutex.RUnlock()
	guildChannel := fmt.Sprintf("%s/%s", guildId, channelId)
	_, ok := register.guildChannelMap[guildChannel]
	return ok
}
