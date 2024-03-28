package socket

import (
	. "aya-backend/service"
	"encoding/json"
	"fmt"
	ws "github.com/gorilla/websocket"
	"log"
	"net/http"
	"net/url"
	"os"
	"slices"
	"unicode/utf8"
)

const (
	WEBSITE_HOST_ORIGIN_ENV = "WEBSITE_HOST_ORIGIN"
)

var (
	acceptableOrigin []string
)

type WSServer struct {
	upg     *ws.Upgrader
	ChanMap map[string]*chan MessageUpdate
}

func handleWSConn(wsServer *WSServer, w http.ResponseWriter, r *http.Request) {

	id := r.PathValue("id")

	c, err := wsServer.upg.Upgrade(w, r, nil)
	if err != nil {
		log.Print("upgrade:", err)
		return
	}

	msgChannel := make(chan MessageUpdate)
	if wsServer.ChanMap[id] != nil {
		msgChan := wsServer.ChanMap[id]
		delete(wsServer.ChanMap, id)
		close(*msgChan)
	}
	wsServer.ChanMap[id] = &msgChannel

	defer func(c *ws.Conn) {
		_ = c.Close()
	}(c)

	defaultCloseHandler := c.CloseHandler()

	c.SetCloseHandler(func(code int, text string) error {
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

func NewWSServer() (*WSServer, error) {

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
		ChanMap: make(map[string]*chan MessageUpdate),
	}

	http.HandleFunc("/stream/{id}", func(w http.ResponseWriter, r *http.Request) {
		handleWSConn(&wsServer, w, r)
	})

	fmt.Println("Web socket ready!")

	return &wsServer, nil
}
