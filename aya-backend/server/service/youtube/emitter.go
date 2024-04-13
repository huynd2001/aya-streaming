package youtube_source

import (
	"aya-backend/server/service"
	"context"
	"fmt"
	"google.golang.org/api/option"
	yt "google.golang.org/api/youtube/v3"
	"os"
	"time"
)

type YoutubeEmitterConfig struct {
	// TODO: Handling OAuth2.0 code flow
	ApiKey string
}

type YoutubeEmitter struct {
	service.ChatEmitter
	youtubeService *yt.Service
	updateEmitter  *chan service.MessageUpdate
}

func (youtubeEmitter *YoutubeEmitter) UpdateEmitter() *chan service.MessageUpdate {
	return youtubeEmitter.updateEmitter
}

func (youtubeEmitter *YoutubeEmitter) CloseEmitter() error {
	close(*youtubeEmitter.updateEmitter)
	return nil
}

func NewEmitter(config *YoutubeEmitterConfig) (*YoutubeEmitter, error) {

	messageUpdates := make(chan service.MessageUpdate)

	ctx := context.Background()

	ytService, err := yt.NewService(ctx, option.WithAPIKey(config.ApiKey))
	if err != nil {
		return nil, err
	}

	// TODO: work with database to retrieve the Youtube URL

	channelId := os.Getenv("TEST_YT_CHANNEL_ID")
	if channelId == "" {
		return nil, fmt.Errorf("env variable TEST_UT_CHANNEL_ID not found")
	}

	searchRes, err := ytService.Search.
		List([]string{"id"}).
		ChannelId(channelId).
		EventType("live").
		Type("video").
		Do()
	if err != nil {
		return nil, err
	}

	if len(searchRes.Items) == 0 {
		return nil, fmt.Errorf("no live videos found for channel %s", channelId)
	}

	videoId := searchRes.Items[0].Id.VideoId

	videoService := yt.NewVideosService(ytService)

	videoRes, err :=
		videoService.
			List([]string{"liveStreamingDetails"}).
			Id(videoId).
			Do()

	if err != nil {
		return nil, err
	}

	liveChatId := ""

	for _, item := range videoRes.Items {
		liveChatId = item.LiveStreamingDetails.ActiveLiveChatId
	}
	if liveChatId == "" {
		return nil, fmt.Errorf("the live has ended")
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
			fmt.Printf("%s\n", err.Error())
		}
	}()

	fmt.Printf("New Youtube Emitter created!\n")

	return &YoutubeEmitter{
		youtubeService: ytService,
		updateEmitter:  &messageUpdates,
	}, nil
}
