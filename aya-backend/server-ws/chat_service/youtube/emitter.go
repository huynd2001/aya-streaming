package youtube_source

import (
	"aya-backend/server-ws/auth"
	"aya-backend/server-ws/chat_service"
	"context"
	"fmt"
	"github.com/fatih/color"
	"github.com/gorilla/mux"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/option"
	yt "google.golang.org/api/youtube/v3"
	"sync"
)

type YoutubeEmitterConfig struct {
	ApiKey       string
	ClientID     string
	ClientSecret string

	AuthRouter           *mux.Router
	AuthRedirectBasedUrl string
}

type YoutubeEmitter struct {
	chat_service.ChatEmitter
	chat_service.ResourceRegister

	mutex sync.Mutex

	updateEmitter       chan chat_service.MessageUpdate
	errorEmitter        chan error
	register            *youtubeRegister
	resource2Subscriber map[string]map[string]bool
}

func (emitter *YoutubeEmitter) Register(subscriber string, resourceInfo any) {
	ytInfo, ok := resourceInfo.(YoutubeInfo)
	if !ok {
		return
	}
	emitter.mutex.Lock()
	defer emitter.mutex.Unlock()
	channelId := ytInfo.YoutubeChannelId
	if emitter.resource2Subscriber[channelId] == nil {
		emitter.resource2Subscriber[channelId] = make(map[string]bool)
		emitter.resource2Subscriber[channelId][subscriber] = true
		emitter.register.registerChannel(channelId)
	} else {
		emitter.resource2Subscriber[channelId][subscriber] = true
	}
}

func (emitter *YoutubeEmitter) Deregister(subscriber string, resourceInfo any) {

	ytInfo, ok := resourceInfo.(YoutubeInfo)
	if !ok {
		return
	}
	emitter.mutex.Lock()
	defer emitter.mutex.Unlock()
	channelId := ytInfo.YoutubeChannelId
	if emitter.resource2Subscriber[channelId] == nil {
		// ignore since there is no resource to deregister
		return
	}
	delete(emitter.resource2Subscriber[channelId], subscriber)
	if len(emitter.resource2Subscriber[channelId]) == 0 {
		delete(emitter.resource2Subscriber, channelId)
		emitter.register.deregisterChannel(channelId)
	}

}

func (emitter *YoutubeEmitter) UpdateEmitter() chan chat_service.MessageUpdate {
	return emitter.updateEmitter
}

func (emitter *YoutubeEmitter) CloseEmitter() error {
	emitter.register.Stop()
	close(emitter.updateEmitter)
	close(emitter.errorEmitter)
	return nil
}

func (emitter *YoutubeEmitter) ErrorEmitter() chan error {
	return emitter.errorEmitter
}

func getApiKeyYTService(ctx context.Context, config *YoutubeEmitterConfig) (*yt.Service, error) {
	ytService, err := yt.NewService(ctx, option.WithAPIKey(config.ApiKey))
	if err != nil {
		return nil, err
	}
	return ytService, nil
}

func getOauthYTService(ctx context.Context, config *YoutubeEmitterConfig) (*yt.Service, error) {
	workflow := auth.NewWorkflow()

	// Configure an OpenID Connect aware OAuth2 client.
	oauth2Config := oauth2.Config{
		ClientID:     config.ClientID,
		ClientSecret: config.ClientSecret,
		RedirectURL:  fmt.Sprintf("%s/youtube.callback", config.AuthRedirectBasedUrl),

		// Discovery returns the OAuth2 endpoints.
		Endpoint: google.Endpoint,

		// "openid" is a required scope for OpenID Connect flows.
		Scopes: []string{yt.YoutubeScope},
	}

	workflow.SetUpRedirectAndCodeChallenge(
		config.AuthRouter.PathPrefix("/youtube.redirect").Subrouter(),
		config.AuthRouter.PathPrefix("/youtube.callback").Subrouter(),
	)
	workflow.SetUpAuth(
		oauth2Config,
		fmt.Sprintf("%s/youtube.redirect", config.AuthRedirectBasedUrl),
	)

	// Await for the tokenSource from the workflow channel
	tokenSource := <-workflow.TokenSourceCh()

	ytService, err := yt.NewService(ctx, option.WithTokenSource(tokenSource))
	if err != nil {
		return nil, err
	}
	return ytService, nil
}

// NewEmitter create a new YouTube emitter. Note that this blocks until the oauth key is
// retrieved from the workflow.
func NewEmitter(config *YoutubeEmitterConfig) (*YoutubeEmitter, error) {

	messageUpdates := make(chan chat_service.MessageUpdate)
	errorCh := make(chan error)

	ctx := context.Background()

	apiYTService, err := getApiKeyYTService(ctx, config)
	if err != nil {
		return nil, err
	}

	youtubeEmitter := YoutubeEmitter{
		updateEmitter:       messageUpdates,
		errorEmitter:        errorCh,
		register:            newYoutubeRegister(apiYTService, messageUpdates),
		resource2Subscriber: make(map[string]map[string]bool),
	}

	go func() {
		ytService, err := getOauthYTService(ctx, config)
		if err != nil {
			errorCh <- err
			return
		}
		youtubeEmitter.register.SetYTService(ytService)
	}()

	color.Green("New Youtube Emitter created!\n")
	return &youtubeEmitter, nil
}
