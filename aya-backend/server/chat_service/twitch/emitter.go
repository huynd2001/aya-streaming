package twitch_source

import (
	"aya-backend/server/auth"
	"aya-backend/server/chat_service"
	"fmt"
	"github.com/fatih/color"
	"github.com/gempir/go-twitch-irc/v4"
	"github.com/gorilla/mux"
	"golang.org/x/oauth2"
	twitch2 "golang.org/x/oauth2/twitch"
	"sync"
	"time"
)

type TwitchEmitterConfig struct {
	ClientID     string
	ClientSecret string
	BotUserName  string

	AuthRouter           *mux.Router
	AuthRedirectBasedUrl string
}

type TwitchEmitter struct {
	chat_service.ChatEmitter
	chat_service.ResourceRegister

	mutex sync.Mutex

	updateEmitter chan chat_service.MessageUpdate
	errorEmitter  chan error

	stopSignalCh chan bool

	resource2Subscriber map[string]map[string]bool

	twitchClient *twitch.Client
}

func (emitter *TwitchEmitter) Register(subscriber string, resourceInfo any) {
	twitchInfo, ok := resourceInfo.(TwitchInfo)
	if !ok {
		return
	}
	emitter.mutex.Lock()
	defer emitter.mutex.Unlock()
	channelName := twitchInfo.TwitchChannelName
	if emitter.resource2Subscriber[channelName] == nil {
		emitter.resource2Subscriber[channelName] = make(map[string]bool)
		emitter.resource2Subscriber[channelName][subscriber] = true
		emitter.twitchClient.Join(channelName)
	} else {
		emitter.resource2Subscriber[channelName][subscriber] = true
	}

}

func (emitter *TwitchEmitter) Deregister(subscriber string, resourceInfo any) {
	twitchInfo, ok := resourceInfo.(TwitchInfo)
	if !ok {
		return
	}
	emitter.mutex.Lock()
	defer emitter.mutex.Unlock()
	channelName := twitchInfo.TwitchChannelName
	if emitter.resource2Subscriber[channelName] == nil {
		// ignore since there is no resource to deregister
		return
	}
	delete(emitter.resource2Subscriber[channelName], subscriber)
	if len(emitter.resource2Subscriber[channelName]) == 0 {
		delete(emitter.resource2Subscriber, channelName)
		emitter.twitchClient.Depart(channelName)
	}
}

func (emitter *TwitchEmitter) UpdateEmitter() chan chat_service.MessageUpdate {
	return emitter.updateEmitter
}

func (emitter *TwitchEmitter) CloseEmitter() error {
	emitter.mutex.Lock()
	defer emitter.mutex.Unlock()
	clientErr := emitter.twitchClient.Disconnect()
	return clientErr

}

func (emitter *TwitchEmitter) ErrorEmitter() chan error {
	return emitter.errorEmitter
}

func TwitchPrivateMessageHandler(parser *TwitchMessageParser, msgChan chan chat_service.MessageUpdate) func(message twitch.PrivateMessage) {
	return func(twitchMsg twitch.PrivateMessage) {
		msgChan <- parser.ParseMessage(twitchMsg)
	}
}

func (emitter *TwitchEmitter) setClient(newClient *twitch.Client) error {
	emitter.mutex.Lock()
	defer emitter.mutex.Unlock()
	err := emitter.twitchClient.Disconnect()
	if err != nil {
		color.Red("Error when disconnect the current twitch client\n")
		emitter.ErrorEmitter() <- err
	}
	parser := TwitchMessageParser{}
	newClient.OnPrivateMessage(TwitchPrivateMessageHandler(&parser, emitter.updateEmitter))

	emitter.twitchClient = newClient
	return newClient.Connect()
}

func NewEmitter(config TwitchEmitterConfig) (*TwitchEmitter, error) {
	emitter := TwitchEmitter{
		updateEmitter: make(chan chat_service.MessageUpdate),
		errorEmitter:  make(chan error),
		twitchClient:  twitch.NewAnonymousClient(),
	}

	clientUpdateCh := make(chan *twitch.Client)
	go func() {
		for {
			newClient := <-clientUpdateCh
			err := emitter.setClient(newClient)
			if err != nil {
				emitter.errorEmitter <- err
			}
		}
	}()

	go func() {
		workflow := auth.NewWorkflow()

		oauth2Config := oauth2.Config{
			ClientID:     config.ClientID,
			ClientSecret: config.ClientSecret,
			RedirectURL:  fmt.Sprintf("%s/twitch.callback", config.AuthRedirectBasedUrl),

			Endpoint: twitch2.Endpoint,

			Scopes: []string{"chat:edit", "chat:read"},
		}

		workflow.SetUpRedirectAndCodeChallenge(
			config.AuthRouter.PathPrefix("/twitch.redirect").Subrouter(),
			config.AuthRouter.PathPrefix("/twitch.callback").Subrouter(),
		)

		workflow.SetUpAuth(
			oauth2Config,
			fmt.Sprintf("%s/twitch.redirect", config.AuthRedirectBasedUrl),
		)

		tokenSource := <-workflow.TokenSourceCh()

		for {
			token, err := tokenSource.Token()
			if err != nil {
				color.Red("stop client retrieval process")
				emitter.errorEmitter <- fmt.Errorf("cannot get token from retrieved token source: %s", err.Error())
				return
			}
			// TODO: get the claim from the access OAUTH token from oidc code flow?
			client := twitch.NewClient(config.BotUserName, fmt.Sprintf("oauth2:%s", token.AccessToken))
			err = emitter.setClient(client)
			if err != nil {
				emitter.errorEmitter <- err
				return
			}
			expirationTime := token.Expiry
			refreshDuration := time.Until(expirationTime)
			select {
			case <-time.After(refreshDuration):
			case <-emitter.stopSignalCh:
				color.Red("stop client retrieval process")
				return
			}
		}
	}()

	color.Green("New Twitch Emitter created!\n")
	return &emitter, nil
}
