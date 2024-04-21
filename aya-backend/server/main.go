package main

import (
	"aya-backend/server/api"
	"aya-backend/server/db"
	"aya-backend/server/hubs"
	"aya-backend/server/service/composed"
	"aya-backend/server/socket"
	"errors"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path"
	"strings"
	"syscall"
)

const (
	SOURCES_ENV     = "SOURCES"
	DB_PATH_ENV     = "DB_PATH"
	SQL_DB_PATH_ENV = "SQL_DB_PATH"
	DEFAULT_DB_PATH = "data"
	DB_NAME         = "aya.db"

	REDIRECT_URL_ENV = "REDIRECT_URL"
)

func getDB() (*gorm.DB, error) {
	var dataLocation string
	sqlDbPath := os.Getenv(SQL_DB_PATH_ENV)
	if sqlDbPath == "" {
		fmt.Printf("%s environment variable not set\n", SQL_DB_PATH_ENV)
		fmt.Printf("Retrieving startup info from %s\n", DB_PATH_ENV)
		dbPath := os.Getenv(DB_PATH_ENV)

		if dbPath == "" {
			fmt.Printf("%s environment variable not set\n", DB_PATH_ENV)
			fmt.Printf("Assuming defualt db path (%s)\n", DEFAULT_DB_PATH)
		}

		dataLocation = path.Join(dbPath, DB_NAME)
	} else {
		dataLocation = sqlDbPath
	}

	gormDB, err := gorm.Open(sqlite.Open(dataLocation), &gorm.Config{})
	if err != nil {
		fmt.Printf("Failed to connect to database!\n")
		return nil, err
	}
	return gormDB, nil

}

func parseEmitterConfig(msgSettingStr string) *composed.MessageChannelConfig {

	config := composed.MessageChannelConfig{
		Test:    false,
		Discord: false,
		Youtube: false,
	}
	enabledSources := strings.Split(msgSettingStr, " ")
	for _, enabledSource := range enabledSources {
		switch enabledSource {
		case "test_source":
			config.Test = true
		case "discord":
			config.Discord = true
		case "youtube":
			config.Youtube = true
		}
	}

	return &config
}

func main() {

	server := &http.Server{
		Addr: ":8000",
	}

	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	fmt.Println("Hello, world!")
	r := mux.NewRouter()

	gormDB, err := getDB()
	if err != nil {
		fmt.Printf("Error during accessing the db: %s\n", err.Error())
		return
	}

	enabledSourceStr := os.Getenv(SOURCES_ENV)
	msgChanConfig := parseEmitterConfig(enabledSourceStr)

	msgChanConfig.BaseURL = os.Getenv(REDIRECT_URL_ENV)
	msgChanConfig.Router = r

	msgChanEmitter := composed.NewMessageEmitter(msgChanConfig)
	msgHub := hubs.NewMessageHub()
	infoDB := db.NewInfoDB(gormDB)

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)

	streamRouter := r.PathPrefix("/stream").Subrouter()

	wsServer, err := socket.NewWSServer(streamRouter, msgHub, msgChanEmitter, infoDB)
	if err != nil {
		fmt.Printf("Error during create the websocket server: %s\n", err.Error())
		return
	}

	apiRouter := r.PathPrefix("/api").Subrouter()

	api.NewApiServer(gormDB, apiRouter)

	http.Handle("/", r)
	fmt.Println("Server's up and running!")

	go func() {
		for {
			select {
			case msg := <-msgChanEmitter.UpdateEmitter():
				fmt.Printf("%#v\n", msg)
				sessionIds := msgHub.GetSessionId(msg.ExtraFields)
				wsServer.SendMessageToSessions(sessionIds, msg)
			case <-sc:
				fmt.Println("End Server!")
				if err := server.Close(); err != nil {
					fmt.Printf("Error when closing server: %s\n", err.Error())
				}
				if err := msgChanEmitter.CloseEmitter(); err != nil {
					fmt.Printf("%s\n", err.Error())
				}
			}
		}

	}()

	if err := server.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
		fmt.Printf("HTTP server error: %s\n", err.Error())
	}
}
