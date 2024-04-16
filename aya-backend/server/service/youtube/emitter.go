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
	"sync"
	"time"
)

type YoutubeSpecificInfo struct {
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

type channelsTable struct {
	mutexLock         sync.Mutex
	registeredChannel map[string]chan bool
	liveChatApiCaller *liveChatApiCaller
	ytService         *yt.Service
}

func (chanTable *channelsTable) removeChannel(channelId string) {
	chanTable.mutexLock.Lock()
	defer chanTable.mutexLock.Unlock()
	if chanTable.registeredChannel == nil {
		// Don't have to do anything
		fmt.Printf("channel %s have not been registered, doing nothing\n", channelId)
		return
	}
	chanTable.registeredChannel[channelId] <- true
	delete(chanTable.registeredChannel, channelId)
}

func (chanTable *channelsTable) registerChannel(channelId string, msgChan chan service.MessageUpdate) {
	// attempt to get the channel info, i.e. is there any live vid at the moment

	chanTable.mutexLock.Lock()
	defer chanTable.mutexLock.Unlock()

	if chanTable.registeredChannel[channelId] != nil {
		// Do not have to do anything, since it is already been registered
		fmt.Printf("channel %s have been registered, doing nothing\n", channelId)
		return
	}

	stopSignals := make(chan bool)
	chanTable.registeredChannel[channelId] = stopSignals
	errCh := make(chan error)
	stopDuringListening := make(chan bool)

	ytParser := YoutubeMessageParser{}

	setupChannel := func() {
		searchRes, err := chanTable.ytService.Search.
			List([]string{"id"}).
			ChannelId(channelId).
			EventType("live").
			Type("video").
			Do()
		if err != nil {
			errCh <- err
			return
		}

		if len(searchRes.Items) == 0 {
			errCh <- fmt.Errorf("no live videos found for channel %s", channelId)
			return
		}

		videoId := searchRes.Items[0].Id.VideoId

		videoRes, err :=
			chanTable.ytService.Videos.
				List([]string{"liveStreamingDetails"}).
				Id(videoId).
				Do()

		if err != nil {
			errCh <- err
			return
		}

		liveChatId := ""

		for _, item := range videoRes.Items {
			liveChatId = item.LiveStreamingDetails.ActiveLiveChatId
		}

		if liveChatId == "" {
			errCh <- fmt.Errorf("the live has ended")
			return
		}

		apiErrCh := make(chan error)
		responseCh := make(chan *yt.LiveChatMessageListResponse)
		var pageToken *string

		for {
			liveChatMessagesService := yt.NewLiveChatMessagesService(chanTable.ytService)
			liveChatServiceCall := liveChatMessagesService.List(liveChatId, []string{"snippet", "authorDetails"})
			if pageToken != nil {
				liveChatServiceCall = liveChatServiceCall.PageToken(*pageToken)
			}
			liveChatApiRequest := liveChatAPIRequest{
				requestCall: liveChatServiceCall,
				responseCh:  responseCh,
				errCh:       apiErrCh,
			}
			chanTable.liveChatApiCaller.Request(liveChatApiRequest)
			select {
			case <-stopSignals:
				close(stopSignals)
				stopDuringListening <- true
				return
			case err := <-apiErrCh:
				close(apiErrCh)
				fmt.Printf("Error during api calls: %v\n", err.Error())
				errCh <- err
				return
			case response := <-responseCh:
				pageToken = &response.NextPageToken
				for _, item := range response.Items {
					if item != nil && item.Snippet != nil {
						publishedTime, err := time.Parse(time.RFC3339, item.Snippet.PublishedAt)
						if err != nil {
							fmt.Println("Error when parsing time in chat. Opt for current Time")
							publishedTime = time.Now()
						}
						fmt.Println(publishedTime.Format(time.RFC822Z))
						msgChan <- service.MessageUpdate{
							UpdateTime: publishedTime,
							Update:     service.New,
							Message:    ytParser.ParseMessage(item),
							ExtraFields: YoutubeSpecificInfo{
								YoutubeChannelId: channelId,
							},
						}
					}
				}
			}
		}
	}

	go func() {
		go setupChannel()
		for {
			select {
			case err := <-errCh:
				// sleep for a duration before a cool reset
				sleepDuration := 1 * time.Minute
				fmt.Printf("Error during processing channel %s: %s\nReseting in %s\n", channelId, err.Error(), sleepDuration)
				time.Sleep(sleepDuration)
				go setupChannel()
			case _ = <-stopDuringListening:
				close(stopDuringListening)
				close(errCh)
				return
			}
		}
	}()

}

type YoutubeEmitter struct {
	service.ChatEmitter
	updateEmitter chan service.MessageUpdate
	errorEmitter  chan error
	chanTable     channelsTable
}

func (youtubeEmitter *YoutubeEmitter) RegisterChannel(channelId string) {
	youtubeEmitter.chanTable.registerChannel(channelId, youtubeEmitter.updateEmitter)
}

func (youtubeEmitter *YoutubeEmitter) RemoveChannel(channelId string) {
	youtubeEmitter.chanTable.removeChannel(channelId)
}

func (youtubeEmitter *YoutubeEmitter) UpdateEmitter() chan service.MessageUpdate {
	return youtubeEmitter.updateEmitter
}

func (youtubeEmitter *YoutubeEmitter) CloseEmitter() error {
	youtubeEmitter.chanTable.mutexLock.Lock()
	defer youtubeEmitter.chanTable.mutexLock.Unlock()
	for _, stopCh := range youtubeEmitter.chanTable.registeredChannel {
		stopCh <- true
	}
	youtubeEmitter.chanTable.liveChatApiCaller.Stop()
	close(youtubeEmitter.updateEmitter)
	close(youtubeEmitter.errorEmitter)
	return nil
}

func (youtubeEmitter *YoutubeEmitter) ErrorEmitter() chan error {
	return youtubeEmitter.errorEmitter
}

func SetupAsync(config *YoutubeEmitterConfig, ytEmitter *YoutubeEmitter) {

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
			return
		}

	} else if config.UseApiKey {
		fmt.Println("Using API key")
		ytService, err = yt.NewService(ctx, option.WithAPIKey(config.ApiKey))
		if err != nil {
			ytEmitter.ErrorEmitter() <- err
			return
		}
	} else {
		ytEmitter.ErrorEmitter() <- fmt.Errorf("cannot setup youtube service")
		return
	}

	if ytService == nil {
		ytEmitter.ErrorEmitter() <- fmt.Errorf("cannot setup youtube service from the config")
		return
	}
	ytEmitter.chanTable.ytService = ytService

	fmt.Printf("Youtube emitter setup complete!\n")
}

// NewEmitter create a new YouTube emitter. Note that this blocks until the oauth key is
// retrieved from the workflow.
func NewEmitter(config *YoutubeEmitterConfig) (*YoutubeEmitter, error) {

	messageUpdates := make(chan service.MessageUpdate)
	errorCh := make(chan error)

	youtubeEmitter := YoutubeEmitter{
		updateEmitter: messageUpdates,
		errorEmitter:  errorCh,
		chanTable: channelsTable{
			registeredChannel: make(map[string]chan bool),
			liveChatApiCaller: nil,
		},
	}

	go func() {
		youtubeEmitter.chanTable.mutexLock.Lock()
		SetupAsync(config, &youtubeEmitter)
		youtubeEmitter.chanTable.liveChatApiCaller = newApiCaller(youtubeEmitter.chanTable.ytService)
		youtubeEmitter.chanTable.liveChatApiCaller.Start()
		youtubeEmitter.chanTable.mutexLock.Unlock()
	}()

	return &youtubeEmitter, nil
}
