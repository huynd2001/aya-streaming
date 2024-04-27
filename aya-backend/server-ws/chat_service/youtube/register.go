package youtube_source

import (
	"aya-backend/server-ws/chat_service"
	"fmt"
	"sync"
	"time"

	"github.com/fatih/color"
	yt "google.golang.org/api/youtube/v3"
)

const (
	TIME_UNTIL_RETRY = 30 * time.Second
)

type youtubeRegister struct {
	mutex             sync.Mutex
	channelKillSignal map[string]chan bool
	apiCaller         *liveChatApiCaller
	ytService         *yt.Service
	msgChan           chan chat_service.MessageUpdate
}

func newYoutubeRegister(ytService *yt.Service, msgChan chan chat_service.MessageUpdate) *youtubeRegister {
	youtubeReg := youtubeRegister{
		channelKillSignal: make(map[string]chan bool),
		apiCaller:         newApiCaller(ytService),
		ytService:         ytService,
		msgChan:           msgChan,
	}
	return &youtubeReg
}

func getLiveChatIdFromChannelId(ytService *yt.Service, channelId string) (string, error) {
	searchRes, err := ytService.Search.
		List([]string{"id"}).
		ChannelId(channelId).
		EventType("live").
		Type("video").
		Do()
	if err != nil {
		return "", err
	}

	if len(searchRes.Items) == 0 {
		return "", fmt.Errorf("no live videos found for channel %s", channelId)
	}

	videoId := searchRes.Items[0].Id.VideoId

	videoRes, err :=
		ytService.Videos.
			List([]string{"liveStreamingDetails"}).
			Id(videoId).
			Do()

	if err != nil {
		return "", err
	}

	var liveChatId string

	for _, item := range videoRes.Items {
		liveChatId = item.LiveStreamingDetails.ActiveLiveChatId
	}

	if liveChatId == "" {
		return "", fmt.Errorf("cannot find live videos on channel %s", channelId)
	}

	color.Green("Got the live video for channel %s", channelId)
	return liveChatId, nil
}

func listenForChatMessages(
	ytService *yt.Service,
	apiCaller *liveChatApiCaller,
	liveChatId string,
	channelId string,
	stopSignal chan bool,
	parser *YoutubeMessageParser,
) chan chat_service.MessageUpdate {
	var err error
	msgChan := make(chan chat_service.MessageUpdate)

	go func() {
		<-stopSignal
		err = fmt.Errorf(color.RedString("Stop Signal received. Kill the live stream reading"))
	}()

	go func() {
		var pageToken *string
		for err == nil {
			liveChatMessagesService := yt.NewLiveChatMessagesService(ytService)
			liveChatServiceCall := liveChatMessagesService.List(liveChatId, []string{"snippet", "authorDetails"})
			if pageToken != nil {
				liveChatServiceCall = liveChatServiceCall.PageToken(*pageToken)
			}
			color.Green("Calling liveChatApi")
			responseCh, apiErrCh := apiCaller.Request(liveChatServiceCall)
			color.Yellow("api call dispatch")
			select {
			case apiErr := <-apiErrCh:
				err = apiErr
			case response := <-responseCh:
				color.Green("response received!")
				for _, item := range response.Items {
					if item != nil && item.Snippet != nil {
						publishedTime, parseErr := time.Parse(time.RFC3339, item.Snippet.PublishedAt)
						if parseErr != nil {
							publishedTime = time.Now()
						}
						fmt.Printf("%#v\n", item)
						msgChan <- chat_service.MessageUpdate{
							UpdateTime: publishedTime,
							Update:     chat_service.New,
							Message:    parser.ParseMessage(item),
							ExtraFields: YoutubeInfo{
								YoutubeChannelId: channelId,
							},
						}
					}
				}
				pageToken = &response.NextPageToken
			}
		}
		color.Red("Error during listening to live channel %s: %s", channelId, err.Error())
		close(msgChan)
	}()

	return msgChan
}

func (register *youtubeRegister) registerChannel(channelId string) {
	// attempt to get the channel info, i.e. is there any live vid at the moment

	register.mutex.Lock()
	defer register.mutex.Unlock()

	if register.ytService == nil {
		fmt.Printf("ytService not set up, skipping registering %s\n", channelId)
		return
	}

	if register.channelKillSignal[channelId] != nil {
		// Do not have to do anything, since it is already been registered
		fmt.Printf("channel %s have been registered, doing nothing\n", channelId)
		return
	}

	stopSignals := make(chan bool)
	errCh := make(chan error)
	stopDuringListening := make(chan bool)
	register.channelKillSignal[channelId] = stopSignals

	ytParser := YoutubeMessageParser{}

	setupChannel := func() chan chat_service.MessageUpdate {

		liveChatId, err := getLiveChatIdFromChannelId(register.ytService, channelId)
		if err != nil {
			errCh <- err
			return nil
		}

		return listenForChatMessages(register.ytService, register.apiCaller, liveChatId, channelId, stopDuringListening, &ytParser)
	}

	go func() {

		for {
			color.Green("Start listening for messages from channel %s", channelId)
			go func() {
				liveChatMsg := setupChannel()
				color.Cyan("start listening from livechat")
				if liveChatMsg == nil {
					return
				} else {
					ok := true
					var ytMsg chat_service.MessageUpdate
					for ok {
						ytMsg, ok = <-liveChatMsg
						if ok {
							register.msgChan <- ytMsg
						}

					}
				}
			}()
			select {
			case err := <-errCh:
				// sleep for a duration before a cool reset
				color.Red("Error during processing channel %s: %s\n", channelId, err.Error())
				color.Yellow("Resetting in %s\n", TIME_UNTIL_RETRY)
				select {
				case <-time.After(TIME_UNTIL_RETRY):
				case <-stopSignals:
					stopDuringListening <- true
					color.Red("Stop when listening for channel %s. Return", channelId)
					return
				}
			case <-stopSignals:
				stopDuringListening <- true
				color.Red("Stop when listening for channel %s. Return", channelId)
				return
			}
		}
	}()

	fmt.Printf("Finish register channel %s\n", channelId)
}

func (register *youtubeRegister) SetYTService(ytService *yt.Service) {
	register.ytService = ytService
	register.apiCaller.SetYTService(ytService)
}

func (register *youtubeRegister) deregisterChannel(channelId string) {
	register.mutex.Lock()
	defer register.mutex.Unlock()
	if register.channelKillSignal[channelId] == nil {
		// Don't have to do anything
		fmt.Printf("channel %s have not been registered, doing nothing\n", channelId)
		return
	}
	register.channelKillSignal[channelId] <- true
	close(register.channelKillSignal[channelId])
	delete(register.channelKillSignal, channelId)
	fmt.Printf("channel %s has been deregistered\n", channelId)
}

func (register *youtubeRegister) Stop() {
	register.mutex.Lock()
	defer register.mutex.Unlock()
	for channelId, killSig := range register.channelKillSignal {
		killSig <- true
		color.Red("Kill Signal sent to channel %s", channelId)
		close(killSig)
		delete(register.channelKillSignal, channelId)
	}

	register.apiCaller.Stop()
}
