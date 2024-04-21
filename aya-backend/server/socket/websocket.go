package socket

import (
	"aya-backend/server/db"
	"aya-backend/server/hubs"
	. "aya-backend/server/service"
	. "aya-backend/server/service/composed"
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	ws "github.com/gorilla/websocket"
	"net/http"
	"net/url"
	"os"
	"slices"
	"sync"
	"unicode/utf8"
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
	infoDB           *db.InfoDB

	ChanMap map[string]*WSConnectionMap
}

func (server *WSServer) registerSessionForMessages(sessionId string) {
	resources := server.infoDB.GetResourcesOfSession(sessionId)
	for _, resource := range resources {
		(*server.msgHub).AddSession(sessionId, resource)
	}
	(*server.resourceRegister).Register(resources)
}

func (server *WSServer) deregisterSessionForMessages(sessionId string) {
	resources := server.infoDB.GetResourcesOfSession(sessionId)
	(*server.msgHub).RemoveSession(sessionId)
	(*server.resourceRegister).Deregister(resources)
}

func handleWSConn(wsServer *WSServer, w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)
	id := vars["id"]
	if id == "" {
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
	if wsServer.ChanMap[id] == nil {
		wsServer.ChanMap[id] = &WSConnectionMap{
			MessageConnChan: make(map[int]chan MessageUpdate),
			CountId:         0,
		}
	}
	wsServer.ChanMap[id].CountId += 1

	wsConnectionId := wsServer.ChanMap[id].CountId

	wsServer.ChanMap[id].MessageConnChan[wsConnectionId] = msgChannel
	wsServer.registerSessionForMessages(id)
	wsServer.mutex.Unlock()

	defer func(c *ws.Conn) {
		_ = c.Close()
	}(c)

	defaultCloseHandler := c.CloseHandler()

	c.SetCloseHandler(func(code int, text string) error {
		fmt.Println("close websocket")
		wsServer.mutex.Lock()
		if wsServer.ChanMap[id] != nil {
			delete(wsServer.ChanMap[id].MessageConnChan, wsConnectionId)
			if len(wsServer.ChanMap[id].MessageConnChan) == 0 {
				wsServer.deregisterSessionForMessages(id)
			}
		}
		wsServer.mutex.Unlock()
		err := defaultCloseHandler(code, text)
		return err
	})

	fmt.Printf("Session %s is connected\n", id)

	for {
		newMessage := <-msgChannel
		newMessageStr, err := json.Marshal(newMessage)
		if err != nil {
			fmt.Printf("Error found while marshal msg:\n%s\n", err.Error())
			continue
		}

		err = c.WriteMessage(ws.TextMessage, newMessageStr)
		if err != nil {
			fmt.Printf("Error counter while send msg:\n%s\n", err.Error())
			break
		}
	}
}

func equalASCIIFold(s, t string) bool {
	for s != "" && t != "" {
		sr, size := utf8.DecodeRuneInString(s)
		s = s[size:]
		tr, size := utf8.DecodeRuneInString(t)
		t = t[size:]
		if sr == tr {
			continue
		}
		if 'A' <= sr && sr <= 'Z' {
			sr = sr + 'a' - 'A'
		}
		if 'A' <= tr && tr <= 'Z' {
			tr = tr + 'a' - 'A'
		}
		if sr != tr {
			return false
		}
	}
	return s == t
}

func NewWSServer(
	s *mux.Router,
	msgHub *hubs.MessageHub,
	resourceRegister *MessageEmitter,
	infoDB *db.InfoDB,
) (*WSServer, error) {

	websiteOrigin := os.Getenv(WEBSITE_HOST_ORIGIN_ENV)
	if websiteOrigin != "" {
		acceptableOrigin = append(acceptableOrigin, websiteOrigin)
	}

	upg := ws.Upgrader{}

	upg.CheckOrigin = func(r *http.Request) bool {
		origin := r.Header["Origin"]

		// this means that the ws call was from the same origin as the host
		if len(origin) == 0 {
			return true
		}

		u, err := url.Parse(origin[0])
		if err != nil {
			return false
		}

		fmt.Println(u.Host)

		return slices.ContainsFunc(acceptableOrigin, func(acceptableHost string) bool {
			return equalASCIIFold(u.Host, acceptableHost)
		})
	}

	wsServer := WSServer{
		upg:              &upg,
		msgHub:           msgHub,
		infoDB:           infoDB,
		resourceRegister: resourceRegister,
		ChanMap:          make(map[string]*WSConnectionMap),
	}

	s.HandleFunc("/{id}", func(w http.ResponseWriter, r *http.Request) {
		handleWSConn(&wsServer, w, r)
	})

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
