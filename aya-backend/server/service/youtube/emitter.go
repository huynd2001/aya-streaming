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
	"os"
	"time"
)

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
	updateEmitter *chan service.MessageUpdate
	errorEmitter  chan error
}

func (youtubeEmitter *YoutubeEmitter) UpdateEmitter() *chan service.MessageUpdate {
	return youtubeEmitter.updateEmitter
}

func (youtubeEmitter *YoutubeEmitter) CloseEmitter() error {
	close(*youtubeEmitter.updateEmitter)
	close(youtubeEmitter.errorEmitter)
	return nil
}

func (youtubeEmitter *YoutubeEmitter) ErrorEmitter() chan error {
	return youtubeEmitter.errorEmitter
}

// NewEmitter create a new YouTube emitter. Note that this blocks until the oauth key is
// retrieved from the workflow.
// TODO: make the error return a channel instead to listen to error handling
func NewEmitter(config *YoutubeEmitterConfig) (*YoutubeEmitter, error) {

	messageUpdates := make(chan service.MessageUpdate)
	errorCh := make(chan error)

	var ytService *yt.Service
	var err error
	ctx := context.Background()

	go func() {
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
				errorCh <- err
				return
			}

		} else if config.UseApiKey {
			fmt.Println("Using API key")
			ytService, err = yt.NewService(ctx, option.WithAPIKey(config.ApiKey))
			if err != nil {
				errorCh <- err
				return
			}
		} else {
			errorCh <- fmt.Errorf("cannot setup youtube service")
			return
		}

		// TODO: work with database to retrieve the Youtube URL
		channelId := os.Getenv("TEST_YT_CHANNEL_ID")
		if channelId == "" {
			errorCh <- fmt.Errorf("env variable TEST_UT_CHANNEL_ID not found")
			return
		}

		if ytService == nil {
			errorCh <- fmt.Errorf("cannot setup youtube service from the config")
			return
		}

		searchRes, err := ytService.Search.
			List([]string{"id"}).
			ChannelId(channelId).
			EventType("live").
			Type("video").
			Do()
		if err != nil {
			errorCh <- err
			return
		}

		if len(searchRes.Items) == 0 {
			errorCh <- fmt.Errorf("no live videos found for channel %s", channelId)
			return
		}

		videoId := searchRes.Items[0].Id.VideoId

		videoService := yt.NewVideosService(ytService)

		videoRes, err :=
			videoService.
				List([]string{"liveStreamingDetails"}).
				Id(videoId).
				Do()

		if err != nil {
			errorCh <- err
		}

		liveChatId := ""

		for _, item := range videoRes.Items {
			liveChatId = item.LiveStreamingDetails.ActiveLiveChatId
		}
		if liveChatId == "" {
			errorCh <- fmt.Errorf("the live has ended")
		}

		// repeated polling from the livestream until an error occurred.
		go func() {

			ytParser := YoutubeMessageParser{}

			liveChatMessagesService := yt.NewLiveChatMessagesService(ytService)
			liveChatServiceCall := liveChatMessagesService.List(liveChatId, []string{"snippet", "authorDetails"})

			err := liveChatServiceCall.Pages(context.Background(), func(response *yt.LiveChatMessageListResponse) error {
				waitUntilTimeStamp := time.Now().Add(time.Duration(response.PollingIntervalMillis) * time.Millisecond)
				for _, item := range response.Items {
					if item != nil && item.Snippet != nil {
						publishedTime, err := time.Parse(time.RFC3339, item.Snippet.PublishedAt)
						if err != nil {
							fmt.Println("sup")
							publishedTime = time.Now()
						}
						messageUpdates <- service.MessageUpdate{
							UpdateTime: publishedTime,
							Update:     service.New,
							Message:    ytParser.ParseMessage(item),
						}
					}
				}
				waitDuration := waitUntilTimeStamp.Sub(time.Now())
				if waitDuration > 0 {
					time.Sleep(waitDuration)
				}

				return nil
			})

			if err != nil {
				fmt.Println("End livestream with an error:")
				fmt.Printf("%s\n", err.Error())
			}
		}()

		fmt.Printf("New Youtube Emitter created!\n")
	}()

	return &YoutubeEmitter{
		updateEmitter: &messageUpdates,
		errorEmitter:  errorCh,
	}, nil
}
