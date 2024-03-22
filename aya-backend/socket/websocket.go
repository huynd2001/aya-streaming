package socket

import (
	. "aya-backend/service"
	"encoding/json"
	"fmt"
	ws "github.com/gorilla/websocket"
	"log"
	"net/http"
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
		delete(wsServer.ChanMap, id)
	}
	wsServer.ChanMap[id] = &msgChannel

	defer func(c *ws.Conn) {
		_ = c.Close()
		delete(wsServer.ChanMap, id)
	}(c)

	for {
		newMessage := <-msgChannel
		newMessageStr, err := json.Marshal(newMessage)
		if err != nil {
			fmt.Printf("Error found: %s\n", err.Error())
		}

		err = c.WriteMessage(ws.TextMessage, newMessageStr)
		if err != nil {
			fmt.Println("write:", err)
			break
		}
	}
}

func NewWSServer() (*WSServer, error) {

	upg := ws.Upgrader{}

	wsServer := WSServer{
		upg:     &upg,
		ChanMap: make(map[string]*chan MessageUpdate),
	}

	http.HandleFunc("/test/{id}", func(w http.ResponseWriter, r *http.Request) {
		handleWSConn(&wsServer, w, r)
	})

	fmt.Println("Web socket ready!")

	return &wsServer, nil
}
