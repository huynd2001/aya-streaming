package test_source

import (
	"fmt"
	"streemly-backend/service"
	"time"
)

type TestEmitter struct {
	updateEmitter chan service.MessageUpdate
}

func (testEmitter TestEmitter) UpdateEmitter() chan service.MessageUpdate {
	return testEmitter.updateEmitter
}

func NewEmitter() service.ChatEmitter {

	messageUpdates := make(chan service.MessageUpdate)

	go func() {
		i := 0
		for {
			messageUpdates <- service.MessageUpdate{
				Source: service.TestSource,
				Update: service.New,
				Message: service.Message{
					Source: service.TestSource,
					Id:     fmt.Sprint("{}", i),
					Author: service.Author{
						Username: "Gamers",
						IsAdmin:  true,
						IsBot:    false,
						Color:    "",
					},
					Content:    []service.MessagePart{},
					Attachment: []string{},
				},
			}
			time.Sleep(1 * time.Second)
			i++
		}
	}()

	go func() {
		i := 0
		for {
			messageUpdates <- service.MessageUpdate{
				Source: service.TestSource,
				Update: service.Delete,
				Message: service.Message{
					Source: service.TestSource,
					Id:     fmt.Sprint("{}", i),
					Author: service.Author{
						Username: "Gamers",
						IsAdmin:  true,
						IsBot:    false,
						Color:    "",
					},
					Content:    []service.MessagePart{},
					Attachment: []string{},
				},
			}
			time.Sleep(2 * time.Second)
			i++
		}
	}()

	return TestEmitter{
		updateEmitter: messageUpdates,
	}

}
