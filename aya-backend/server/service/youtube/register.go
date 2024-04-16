package youtube_source

import (
	"aya-backend/server/service"
	"fmt"
	yt "google.golang.org/api/youtube/v3"
	"sync"
	"time"
)

type youtubeRegister struct {
	mutex             sync.Mutex
	channelKillSignal map[string]chan bool
	apiCaller         *liveChatApiCaller
	ytService         *yt.Service
}

func newYoutubeRegister(ytService *yt.Service) *youtubeRegister {
	youtubeReg := youtubeRegister{
		channelKillSignal: make(map[string]chan bool),
		apiCaller:         newApiCaller(ytService),
		ytService:         ytService,
	}
	if ytService == nil {
		// Wait until ytService is here before any registration can start
		youtubeReg.mutex.Lock()
	}
	return &youtubeReg
}

func (register *youtubeRegister) removeChannel(channelId string) {
	register.mutex.Lock()
	defer register.mutex.Unlock()
	if register.channelKillSignal[channelId] == nil {
		// Don't have to do anything
		fmt.Printf("channel %s have not been registered, doing nothing\n", channelId)
		return
	}
	register.channelKillSignal[channelId] <- true
	delete(register.channelKillSignal, channelId)
}

func (register *youtubeRegister) registerChannel(channelId string, msgChan chan service.MessageUpdate) {
	// attempt to get the channel info, i.e. is there any live vid at the moment

	register.mutex.Lock()
	defer register.mutex.Unlock()

	if register.channelKillSignal[channelId] != nil {
		// Do not have to do anything, since it is already been registered
		fmt.Printf("channel %s have been registered, doing nothing\n", channelId)
		return
	}

	stopSignals := make(chan bool)
	register.channelKillSignal[channelId] = stopSignals
	errCh := make(chan error)
	stopDuringListening := make(chan bool)

	ytParser := YoutubeMessageParser{}

	setupChannel := func() {
		searchRes, err := register.ytService.Search.
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
			register.ytService.Videos.
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
			liveChatMessagesService := yt.NewLiveChatMessagesService(register.ytService)
			liveChatServiceCall := liveChatMessagesService.List(liveChatId, []string{"snippet", "authorDetails"})
			if pageToken != nil {
				liveChatServiceCall = liveChatServiceCall.PageToken(*pageToken)
			}
			liveChatApiRequest := liveChatAPIRequest{
				requestCall: liveChatServiceCall,
				responseCh:  responseCh,
				errCh:       apiErrCh,
			}
			register.apiCaller.Request(liveChatApiRequest)
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
							ExtraFields: YoutubeInfo{
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

				select {
				case <-time.After(sleepDuration):
					go setupChannel()
				case <-stopSignals:
					register.Stop()
					return
				}
			case <-stopDuringListening:
				close(stopDuringListening)
				close(errCh)
				return
			}
		}
	}()
}

func (register *youtubeRegister) Start(ytService *yt.Service) {
	register.ytService = ytService
	register.apiCaller.Start(ytService)
	register.mutex.Unlock()
}

func (register *youtubeRegister) Stop() {
	register.mutex.Lock()
	defer register.mutex.Unlock()
	for _, killSig := range register.channelKillSignal {
		killSig <- true
	}
	register.apiCaller.Stop()
}
