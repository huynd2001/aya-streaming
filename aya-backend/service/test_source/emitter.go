package test_source

import (
	"aya-backend/service"
	"fmt"
	"time"
)

type TestEmitter struct {
	updateEmitter chan service.MessageUpdate
}

func (testEmitter *TestEmitter) UpdateEmitter() chan service.MessageUpdate {
	return testEmitter.updateEmitter
}

func NewEmitter() *TestEmitter {

	messageUpdates := make(chan service.MessageUpdate)

	go func() {
		i := 0
		for {
			messageUpdates <- service.MessageUpdate{
				Source: service.TestSource,
				Update: service.New,
				Message: service.Message{
					Id: fmt.Sprint("{}", i),
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
			time.Sleep(1 * time.Second * 5)
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
					Id: fmt.Sprint("{}", i),
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
			time.Sleep(2 * time.Second * 5)
			i++
		}
	}()

	return &TestEmitter{
		updateEmitter: messageUpdates,
	}

}
