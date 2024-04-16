package youtube_source

import (
	"aya-backend/server/auth"
	"aya-backend/server/service"
	"context"
	"fmt"
	"github.com/gorilla/mux"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/option"
	yt "google.golang.org/api/youtube/v3"
)

type YoutubeInfo struct {
	YoutubeChannelId string
}

type YoutubeEmitterConfig struct {
	UseApiKey    bool
	UseOAuth     bool
	ApiKey       string
	ClientID     string
	ClientSecret string

	Router           *mux.Router
	RedirectBasedUrl string
}

type YoutubeEmitter struct {
	service.ChatEmitter
	updateEmitter chan service.MessageUpdate
	errorEmitter  chan error
	register      *youtubeRegister
}

func (youtubeEmitter *YoutubeEmitter) RegisterChannel(channelId string) {
	youtubeEmitter.register.registerChannel(channelId, youtubeEmitter.updateEmitter)
}

func (youtubeEmitter *YoutubeEmitter) DeregisterChannel(channelId string) {
	youtubeEmitter.register.deregisterChannel(channelId)
}

func (youtubeEmitter *YoutubeEmitter) UpdateEmitter() chan service.MessageUpdate {
	return youtubeEmitter.updateEmitter
}

func (youtubeEmitter *YoutubeEmitter) CloseEmitter() error {
	youtubeEmitter.register.Stop()
	close(youtubeEmitter.updateEmitter)
	close(youtubeEmitter.errorEmitter)
	return nil
}

func (youtubeEmitter *YoutubeEmitter) ErrorEmitter() chan error {
	return youtubeEmitter.errorEmitter
}

func SetupAsync(config *YoutubeEmitterConfig, ytEmitter *YoutubeEmitter) *yt.Service {

	var ytService *yt.Service
	var err error
	ctx := context.Background()
	// waits until we finish our setup
	if config.UseOAuth {

		fmt.Println("Using OAuth")
		// resolves the token workflow
		workflow := auth.NewWorkflow()

		// Configure an OpenID Connect aware OAuth2 client.
		oauth2Config := oauth2.Config{
			ClientID:     config.ClientID,
			ClientSecret: config.ClientSecret,
			RedirectURL:  fmt.Sprintf("%s/youtube.callback", config.RedirectBasedUrl),

			// Discovery returns the OAuth2 endpoints.
			Endpoint: google.Endpoint,

			// "openid" is a required scope for OpenID Connect flows.
			Scopes: []string{yt.YoutubeScope},
		}

		workflow.SetUpRedirectAndCodeChallenge(
			config.Router.PathPrefix("/youtube.redirect").Subrouter(),
			config.Router.PathPrefix("/youtube.callback").Subrouter(),
		)
		workflow.SetupAuth(
			oauth2Config,
			fmt.Sprintf("%s/youtube.redirect", config.RedirectBasedUrl),
		)

		// Await for the tokenSource from the workflow channel
		tokenSource := <-workflow.TokenSourceCh()

		ytService, err = yt.NewService(ctx, option.WithTokenSource(tokenSource))
		if err != nil {
			ytEmitter.ErrorEmitter() <- err
			return nil
		}

	} else if config.UseApiKey {
		fmt.Println("Using API key")
		ytService, err = yt.NewService(ctx, option.WithAPIKey(config.ApiKey))
		if err != nil {
			ytEmitter.ErrorEmitter() <- err
			return nil
		}
	} else {
		ytEmitter.ErrorEmitter() <- fmt.Errorf("cannot setup youtube service")
		return nil
	}

	if ytService == nil {
		ytEmitter.ErrorEmitter() <- fmt.Errorf("cannot setup youtube service from the config")
		return nil
	}

	fmt.Printf("Youtube emitter setup complete!\n")
	return ytService
}

// NewEmitter create a new YouTube emitter. Note that this blocks until the oauth key is
// retrieved from the workflow.
func NewEmitter(config *YoutubeEmitterConfig) (*YoutubeEmitter, error) {

	messageUpdates := make(chan service.MessageUpdate)
	errorCh := make(chan error)

	youtubeEmitter := YoutubeEmitter{
		updateEmitter: messageUpdates,
		errorEmitter:  errorCh,
		register:      newYoutubeRegister(nil),
	}

	go func() {
		ytService := SetupAsync(config, &youtubeEmitter)
		youtubeEmitter.register.Start(ytService)
	}()

	return &youtubeEmitter, nil
}
