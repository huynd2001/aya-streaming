package youtube_source

import (
	"time"

	"github.com/fatih/color"
	yt "google.golang.org/api/youtube/v3"
)

type liveChatAPIRequest struct {
	requestCall *yt.LiveChatMessagesListCall
	responseCh  chan *yt.LiveChatMessageListResponse
	errCh       chan error
}

type liveChatApiCaller struct {
	apiStopCallSig chan bool
	requestCall    chan liveChatAPIRequest
	ytService      *yt.Service
}

func newApiCaller(ytService *yt.Service) *liveChatApiCaller {

	apiCaller := liveChatApiCaller{
		apiStopCallSig: make(chan bool),
		requestCall:    make(chan liveChatAPIRequest),
		ytService:      ytService,
	}

	go func() {
		nextApiCall := time.Now()

		for {
			sleepDuration := time.Until(nextApiCall)

			select {
			case <-apiCaller.apiStopCallSig:
				color.Red("Kill api call")
				close(apiCaller.requestCall)
				return
			case <-time.After(sleepDuration):
				select {
				case <-apiCaller.apiStopCallSig:
					color.Red("Kill api call")
					close(apiCaller.requestCall)
					return
				case liveChatCall := <-apiCaller.requestCall:
					color.Yellow("Got an api call")
					response, err := liveChatCall.requestCall.Do()
					if err != nil {
						liveChatCall.errCh <- err
					} else {
						nextApiCall = time.Now().Add(time.Duration(response.PollingIntervalMillis) * time.Millisecond)
						liveChatCall.responseCh <- response
					}
					close(liveChatCall.errCh)
					close(liveChatCall.responseCh)
				}
			}

		}

	}()

	return &apiCaller
}

func (apiCaller *liveChatApiCaller) SetYTService(ytService *yt.Service) {
	apiCaller.ytService = ytService
}

func (apiCaller *liveChatApiCaller) Stop() {
	apiCaller.apiStopCallSig <- true
}

func (apiCaller *liveChatApiCaller) Request(reqCall *yt.LiveChatMessagesListCall) (chan *yt.LiveChatMessageListResponse, chan error) {
	responseCh := make(chan *yt.LiveChatMessageListResponse)
	errCh := make(chan error)
	req := liveChatAPIRequest{
		requestCall: reqCall,
		responseCh:  responseCh,
		errCh:       errCh,
	}
	apiCaller.requestCall <- req
	return responseCh, errCh
}
