package youtube_source

import (
	yt "google.golang.org/api/youtube/v3"
	"time"
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
	return &liveChatApiCaller{
		apiStopCallSig: make(chan bool),
		requestCall:    make(chan liveChatAPIRequest),
		ytService:      ytService,
	}
}

func (apiCaller *liveChatApiCaller) Start(ytService *yt.Service) {
	apiCaller.ytService = ytService

	go func() {
		nextApiCall := time.Now()
		intervalWait := make(chan bool)

		for {
			go func() {
				sleepDuration := nextApiCall.Sub(time.Now())
				if sleepDuration > 0 {
					time.Sleep(sleepDuration)
				}
				intervalWait <- true
			}()

			select {
			case <-apiCaller.apiStopCallSig:
				close(apiCaller.apiStopCallSig)
				close(apiCaller.requestCall)
				return
			case <-intervalWait:
				select {
				case <-apiCaller.apiStopCallSig:
					close(apiCaller.apiStopCallSig)
					close(apiCaller.requestCall)
					return
				case liveChatCall := <-apiCaller.requestCall:

					response, err := liveChatCall.requestCall.Do()
					if err != nil {
						liveChatCall.errCh <- err
					} else {
						nextApiCall = time.Now().Add(time.Duration(response.PollingIntervalMillis) * time.Millisecond)
						liveChatCall.responseCh <- response
					}
				}
			}

		}

	}()
}

func (apiCaller *liveChatApiCaller) Stop() {
	apiCaller.apiStopCallSig <- true
}

func (apiCaller *liveChatApiCaller) Request(req liveChatAPIRequest) {
	apiCaller.requestCall <- req
}
