package socket

import (
	. "aya-backend/server/service"
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	ws "github.com/gorilla/websocket"
	"log"
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
	mutex   sync.RWMutex
	upg     *ws.Upgrader
	ChanMap map[string]*WSConnectionMap
}

func handleWSConn(wsServer *WSServer, w http.ResponseWriter, r *http.Request) {

	id := r.PathValue("id")
	if id == "" {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("id is empty"))
		return
	}

	c, err := wsServer.upg.Upgrade(w, r, nil)
	if err != nil {
		log.Print("upgrade:", err)
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
	wsServer.mutex.Unlock()

	defer func(c *ws.Conn) {
		_ = c.Close()
	}(c)

	defaultCloseHandler := c.CloseHandler()

	c.SetCloseHandler(func(code int, text string) error {
		fmt.Println("close websocket")
		err := defaultCloseHandler(code, text)
		if wsServer.ChanMap[id] != nil {
			delete(wsServer.ChanMap, id)
		}
		return err
	})

	fmt.Println("Connection found!")

	for {
		newMessage := <-msgChannel
		newMessageStr, err := json.Marshal(newMessage)
		if err != nil {
			fmt.Printf("Error found while marshal msg: %s\n", err.Error())
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

func NewWSServer(s *mux.Router) (*WSServer, error) {

	websiteOrigin := os.Getenv(WEBSITE_HOST_ORIGIN_ENV)
	if websiteOrigin != "" {
		acceptableOrigin = append(acceptableOrigin, websiteOrigin)
	}

	upg := ws.Upgrader{}

	// Check origin
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
		upg:     &upg,
		ChanMap: make(map[string]*WSConnectionMap),
	}

	s.HandleFunc("/{id}", func(w http.ResponseWriter, r *http.Request) {
		handleWSConn(&wsServer, w, r)
	})

	fmt.Println("Web socket ready!")

	return &wsServer, nil
}

func (wsServer *WSServer) SendMessageToSession(sessionId string, msg MessageUpdate) {
	wsServer.mutex.RLock()
	defer wsServer.mutex.RUnlock()

	if wsServer.ChanMap[sessionId] == nil {
		fmt.Printf("Do nothing since the session \"%s\" does not exist\n", sessionId)
		return
	}

	for _, conn := range wsServer.ChanMap[sessionId].MessageConnChan {
		conn <- msg
	}
}

func (wsServer *WSServer) SendMessageToSessions(sessionIds []string, msg MessageUpdate) {
	for _, session := range sessionIds {
		wsServer.SendMessageToSession(session, msg)
	}
}
