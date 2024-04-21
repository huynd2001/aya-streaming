package composed

import (
	models "aya-backend/db-models"
	"aya-backend/server/service"
	discordsource "aya-backend/server/service/discord"
	"aya-backend/server/service/test_source"
	youtubesource "aya-backend/server/service/youtube"
	"errors"
	"fmt"
	"github.com/gorilla/mux"
	"os"
)

const (
	YOUTUBE_API_KEY_ENV       = "YOUTUBE_API_KEY"
	YOUTUBE_CLIENT_ID_ENV     = "YOUTUBE_CLIENT_ID"
	YOUTUBE_CLIENT_SECRET_ENV = "YOUTUBE_CLIENT_SECRET"
	YOUTUBE_FLOW_ENV          = "YOUTUBE_FLOW"

	DISCORD_TOKEN_ENV = "DISCORD_TOKEN"
)

type MessageEmitter struct {
	service.ChatEmitter
	discordEmitter *discordsource.DiscordEmitter
	testEmitter    *test_source.TestEmitter
	youtubeEmitter *youtubesource.YoutubeEmitter

	updateEmitter chan service.MessageUpdate
}

func (messageChannel MessageEmitter) Register(resourceInfo any) {
	// resourceInfo should be of type []Resource
	resources, ok := resourceInfo.([]models.Resource)
	if !ok {
		// do thing since the resource is not of correct type
		fmt.Printf("Cannot register %#v\n", resourceInfo)
		return
	}
	for _, resource := range resources {
		switch resource.ResourceType {
		case service.Discord:
			discordInfo, ok := resource.ResourceInfo.(discordsource.DiscordInfo)
			if !ok {
				fmt.Printf("Cannot register %#v\n", resource.ResourceInfo)
			} else {
				messageChannel.discordEmitter.Register(discordInfo)
			}
		case service.Youtube:
			youtubeInfo, ok := resource.ResourceInfo.(youtubesource.YoutubeInfo)
			if !ok {
				fmt.Printf("Cannot register %#v\n", resource.ResourceInfo)
			} else {
				messageChannel.youtubeEmitter.Register(youtubeInfo)
			}
		default:
			fmt.Printf("Not supporting %s, cannot register %#v\n", resource.ResourceType.String(), resource.ResourceInfo)
		}
	}
}

func (messageChannel MessageEmitter) Deregister(resourceInfo any) {
	// resourceInfo should be of type []Resource
	resources, ok := resourceInfo.([]models.Resource)
	if !ok {
		// do thing since the resource is not of correct type
		fmt.Printf("Cannot deregister %#v\n", resourceInfo)
		return
	}
	for _, resource := range resources {
		switch resource.ResourceType {
		case service.Discord:
			discordInfo, ok := resourceInfo.(discordsource.DiscordInfo)
			if !ok {
				fmt.Printf("Cannot deregister %#v\n", resourceInfo)
			} else {
				messageChannel.discordEmitter.Deregister(discordInfo)
			}
		case service.Youtube:
			youtubeInfo, ok := resourceInfo.(youtubesource.YoutubeInfo)
			if !ok {
				fmt.Printf("Cannot deregister %#v\n", resourceInfo)
			} else {
				messageChannel.youtubeEmitter.Deregister(youtubeInfo)
			}
		default:
			fmt.Printf("Not supporting %s, cannot deregister %#v\n", resource.ResourceType.String(), resource.ResourceInfo)

		}
	}
}

type MessageChannelConfig struct {
	Discord bool
	Test    bool
	Youtube bool
	BaseURL string
	Router  *mux.Router
}

func (messageChannel MessageEmitter) UpdateEmitter() chan service.MessageUpdate {
	return messageChannel.updateEmitter
}

func (messageChannel MessageEmitter) CloseEmitter() error {

	close(messageChannel.updateEmitter)

	var testError error = nil
	var discordError error = nil
	var youtubeError error = nil

	if messageChannel.testEmitter != nil {
		testError = messageChannel.testEmitter.CloseEmitter()
	}

	if messageChannel.discordEmitter != nil {
		discordError = messageChannel.discordEmitter.CloseEmitter()
	}

	if messageChannel.youtubeEmitter != nil {
		testError = messageChannel.youtubeEmitter.CloseEmitter()
	}

	err := errors.Join(testError, discordError, youtubeError)

	if err != nil {
		return fmt.Errorf("error encounter during closing: %w", err)
	} else {
		return nil
	}
}

func NewMessageEmitter(messageChannelConfig *MessageChannelConfig) *MessageEmitter {

	messageChannel := MessageEmitter{
		testEmitter:    nil,
		discordEmitter: nil,
		youtubeEmitter: nil,
	}

	if messageChannelConfig.Test {
		testEmitter := test_source.NewEmitter()
		messageChannel.testEmitter = testEmitter
	}

	if messageChannelConfig.Discord {
		discordToken := os.Getenv(DISCORD_TOKEN_ENV)
		discordEmitter, err := discordsource.NewEmitter(discordToken)

		if err != nil {
			fmt.Printf("Error during creating a discord emitter: %s\n", err.Error())
		}

		messageChannel.discordEmitter = discordEmitter
	}

	if messageChannelConfig.Youtube {
		ytApiKey := os.Getenv(YOUTUBE_API_KEY_ENV)
		ytClientId := os.Getenv(YOUTUBE_CLIENT_ID_ENV)
		ytClientSecret := os.Getenv(YOUTUBE_CLIENT_SECRET_ENV)
		ytFlow := os.Getenv(YOUTUBE_FLOW_ENV)

		ytEmitterConfig := &youtubesource.YoutubeEmitterConfig{}

		switch ytFlow {
		case "api":
			ytEmitterConfig.UseApiKey = true
			ytEmitterConfig.ApiKey = ytApiKey
		case "oauth":
			ytEmitterConfig.UseOAuth = true
			ytEmitterConfig.ClientID = ytClientId
			ytEmitterConfig.ClientSecret = ytClientSecret
			ytEmitterConfig.Router = messageChannelConfig.Router.PathPrefix("/auth").Subrouter()
			ytEmitterConfig.RedirectBasedUrl = fmt.Sprintf("%s/auth", messageChannelConfig.BaseURL)
		}

		youtubeEmitter, err := youtubesource.NewEmitter(ytEmitterConfig)
		if err != nil {
			fmt.Printf("Error during creating a youtube emitter: %s\n", err.Error())
		} else {
			messageChannel.youtubeEmitter = youtubeEmitter
		}
	}

	msgC := make(chan service.MessageUpdate)

	if messageChannel.testEmitter != nil {
		go func() {
			for {
				testMsg := <-messageChannel.testEmitter.UpdateEmitter()
				fmt.Println("Message from test source!")
				msgC <- testMsg
			}
		}()
	}

	if messageChannel.discordEmitter != nil {
		go func() {
			for {
				discordMsg := <-messageChannel.discordEmitter.UpdateEmitter()
				fmt.Println("Message from discord!")
				msgC <- discordMsg
			}
		}()
	}

	if messageChannel.youtubeEmitter != nil {
		go func() {
			for {
				select {
				case ytMsg := <-messageChannel.youtubeEmitter.UpdateEmitter():
					fmt.Println("Message from youtube!")
					msgC <- ytMsg
				case err := <-messageChannel.youtubeEmitter.ErrorEmitter():
					fmt.Printf("Error from youtube:%s\n", err.Error())
					break
				}

			}
		}()
	}
	messageChannel.updateEmitter = msgC

	return &messageChannel
}