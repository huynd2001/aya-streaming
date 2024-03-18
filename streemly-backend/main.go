package main

import (
	"fmt"
	"github.com/lpernett/godotenv"
	"log"
	"os"
	"os/signal"
	discordsource "streemly-backend/service/discord"
	testsource "streemly-backend/service/test_source"
	"syscall"
)

func main() {

	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	fmt.Println("Hello, world!")
	testEmitter := testsource.NewEmitter()

	discordToken := os.Getenv("DISCORD_TOKEN")
	discordEmitter := discordsource.NewEmitter(discordToken)

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)

	for {
		select {
		case testMsg := <-testEmitter.UpdateEmitter():
			fmt.Println("Message from test source!")
			fmt.Println("{}", testMsg)
		case discordMsg := <-discordEmitter.UpdateEmitter():
			fmt.Println("Message from discord!")
			fmt.Println("{}", discordMsg)
		case _ = <-sc:
			fmt.Println("End Session!")
			_ = discordEmitter.DiscordClient.Close()
			return
		}
	}
}
