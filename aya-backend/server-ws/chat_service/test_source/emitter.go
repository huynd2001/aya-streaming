package test_source

import (
	. "aya-backend/server-ws/chat_service"
	"fmt"
	"time"
)

type TestEmitter struct {
	ChatEmitter
	updateEmitter chan MessageUpdate
	errorEmitter  chan error
}

func (testEmitter *TestEmitter) UpdateEmitter() chan MessageUpdate {
	return testEmitter.updateEmitter
}

func (testEmitter *TestEmitter) CloseEmitter() error {
	close(testEmitter.updateEmitter)
	return nil
}

func (testEmitter *TestEmitter) ErrorEmitter() chan error {
	return testEmitter.errorEmitter
}

func NewEmitter() *TestEmitter {

	messageUpdates := make(chan MessageUpdate)
	errorEmitter := make(chan error)

	go func() {
		i := 0
		for {
			messageUpdates <- MessageUpdate{
				UpdateTime: time.Now(),
				Update:     New,
				Message: Message{
					Source: TestSource,
					Id:     fmt.Sprintf("%d", i),
					Author: Author{
						Username: "Gamers",
						IsAdmin:  true,
						IsBot:    false,
						Color:    "#ffffff",
					},
					MessageParts: []MessagePart{
						{
							Content: fmt.Sprintf("Bot#%d: Hello from server", i),
						},
					},
					Attachments: []Attachment{},
				},
			}

			go func() {
				a := i
				time.Sleep(1 * time.Second * 30)
				messageUpdates <- MessageUpdate{
					UpdateTime: time.Now(),
					Update:     Delete,
					Message: Message{
						Source: TestSource,
						Id:     fmt.Sprintf("%d", a),
						Author: Author{
							Username: "Gamers",
							IsAdmin:  true,
							IsBot:    false,
							Color:    "#ffffff",
						},
						MessageParts: []MessagePart{},
						Attachments:  []Attachment{},
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
		errorEmitter:  errorEmitter,
	}

}
