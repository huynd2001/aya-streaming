package main

import (
	"fmt"
	testsource "streemly-backend/service/testsource"
)

func main() {
	fmt.Println("Hello, world!")
	testEmitter := testsource.NewEmitter()

	for {
		newMessage := <-testEmitter.UpdateEmitter()
		fmt.Println("{}", newMessage)
	}
}
