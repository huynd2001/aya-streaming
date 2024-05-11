package socket

import (
	. "aya-backend/server-ws/chat_service"
	. "aya-backend/server-ws/chat_service/composed"
	"aya-backend/server-ws/hubs"
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	ws "github.com/gorilla/websocket"
	"net/http"
	"os"
	"sync"
)

const (
	WEBSITE_HOST_ORIGIN_ENV = "WEBSITE_HOST_ORIGIN"
)

var (
	acceptableOrigin []string
)

type WSConnectionMap struct {
	MessageConnChan map[int]chan MessageUpdate
	CountId         int
}

type WSServer struct {
	mutex sync.RWMutex
	upg   *ws.Upgrader

	msgHub           *hubs.MessageHub
	resourceRegister *MessageEmitter

	ChanMap map[string]*WSConnectionMap
}

func (server *WSServer) registerSessionForMessages(sessionId string) {
	server.msgHub.AddSession(sessionId)
}

func (server *WSServer) deregisterSessionForMessages(sessionId string) {
	server.msgHub.RemoveSession(sessionId)
}

func wsHandler(wsServer *WSServer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		sessionUUID := vars["id"]
		if sessionUUID == "" {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte("id is empty"))
			return
		}

		c, err := wsServer.upg.Upgrade(w, r, nil)
		if err != nil {
			fmt.Printf("upgrade: %s\n", err.Error())
			return
		}

		wsServer.mutex.Lock()
		msgChannel := make(chan MessageUpdate)
		if wsServer.ChanMap[sessionUUID] == nil {
			wsServer.ChanMap[sessionUUID] = &WSConnectionMap{
				MessageConnChan: make(map[int]chan MessageUpdate),
				CountId:         0,
			}
		}
		wsServer.ChanMap[sessionUUID].CountId += 1

		wsConnectionId := wsServer.ChanMap[sessionUUID].CountId

		wsServer.ChanMap[sessionUUID].MessageConnChan[wsConnectionId] = msgChannel
		wsServer.registerSessionForMessages(sessionUUID)
		wsServer.mutex.Unlock()

		fmt.Printf("Session %s is connected\n", sessionUUID)

		errChannel := make(chan error)

		go func() {
			for {
				_, msg, err := c.ReadMessage()
				fmt.Printf("Message from connection: %s\n", string(msg))
				if err != nil {
					errChannel <- err
					return
				}
			}
		}()

		var connectErr error

		for connectErr == nil {
			select {
			case newMessage := <-msgChannel:
				newMessageStr, err := json.Marshal(newMessage)
				if err != nil {
					fmt.Printf("Error found while marshal msg:\n%s\n", err.Error())
					continue
				}
				err = c.WriteMessage(ws.TextMessage, newMessageStr)
				if err != nil {
					fmt.Printf("Error counter while send msg:\n%s\n", err.Error())
					connectErr = err
				}
			case err := <-errChannel:
				if err != nil {
					fmt.Printf("Error from connection:\n%s\n", err.Error())
					connectErr = err
				}
			}
		}

		fmt.Printf("End sending message to %s, conn#%d, start cleaning up\n", sessionUUID, wsConnectionId)
		_ = c.Close()
		fmt.Printf("close websocket to %s\n", sessionUUID)
		wsServer.mutex.Lock()
		if wsServer.ChanMap[sessionUUID] != nil {
			delete(wsServer.ChanMap[sessionUUID].MessageConnChan, wsConnectionId)
			if len(wsServer.ChanMap[sessionUUID].MessageConnChan) == 0 {
				wsServer.deregisterSessionForMessages(sessionUUID)
			}
		}
		wsServer.mutex.Unlock()
		fmt.Printf("Session %s is disconnected\n", sessionUUID)
	}
}

func NewWSServer(
	s *mux.Router,
	msgHub *hubs.MessageHub,
	resourceRegister *MessageEmitter,
) (*WSServer, error) {

	websiteOrigin := os.Getenv(WEBSITE_HOST_ORIGIN_ENV)
	if websiteOrigin != "" {
		acceptableOrigin = append(acceptableOrigin, websiteOrigin)
	}

	upg := ws.Upgrader{}

	upg.CheckOrigin = func(r *http.Request) bool {
		return true
	}

	wsServer := WSServer{
		upg:              &upg,
		msgHub:           msgHub,
		resourceRegister: resourceRegister,
		ChanMap:          make(map[string]*WSConnectionMap),
	}

	s.HandleFunc("/{id}", wsHandler(&wsServer))

	fmt.Println("Web socket ready!")

	return &wsServer, nil
}

func (server *WSServer) SendMessageToSession(sessionId string, msg MessageUpdate) {
	server.mutex.RLock()
	defer server.mutex.RUnlock()

	if server.ChanMap[sessionId] == nil {
		fmt.Printf("Do nothing since the session \"%s\" does not exist\n", sessionId)
		return
	}

	for _, conn := range server.ChanMap[sessionId].MessageConnChan {
		conn <- msg
	}
}

func (server *WSServer) SendMessageToSessions(sessionIds []string, msg MessageUpdate) {
	for _, session := range sessionIds {
		server.SendMessageToSession(session, msg)
	}
}
