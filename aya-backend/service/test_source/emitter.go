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

				Update: service.New,
				Message: service.Message{
					Source: service.TestSource,
					Id:     fmt.Sprintf("%d", i),
					Author: service.Author{
						Username: "Gamers",
						IsAdmin:  true,
						IsBot:    false,
						Color:    "#ffffff",
					},
					MessageParts: []service.MessagePart{
						{
							Content: fmt.Sprintf("Bot#%d: Hello from server", i),
						},
					},
					Attachments: []string{},
				},
			}

			go func() {
				a := i
				time.Sleep(1 * time.Second * 30)
				messageUpdates <- service.MessageUpdate{
					Update: service.Delete,
					Message: service.Message{
						Source: service.TestSource,
						Id:     fmt.Sprintf("%d", a),
						Author: service.Author{
							Username: "Gamers",
							IsAdmin:  true,
							IsBot:    false,
							Color:    "#ffffff",
						},
						MessageParts: []service.MessagePart{},
						Attachments:  []string{},
					},
				}
			}()

			time.Sleep(1 * time.Second * 10)
			i++
		}
	}()

	fmt.Println("New Test Emitter created!")

	return &TestEmitter{
		updateEmitter: messageUpdates,
	}

}
