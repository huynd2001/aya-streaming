package main

import (
	"aya-backend/server-api/api"
	"errors"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"net/http"
	"os"
	"os/signal"
	"path"
	"syscall"
)

const (
	DB_PATH_ENV     = "DB_PATH"
	SQL_DB_PATH_ENV = "SQL_DB_PATH"
	DEFAULT_DB_PATH = "data"
	DB_NAME         = "aya.db"
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

func main() {

	server := &http.Server{
		Addr: ":6000",
	}

	err := godotenv.Load()
	if err != nil {
		fmt.Println("Error loading .env file")
	}

	r := mux.NewRouter()

	gormDB, err := getDB()
	if err != nil {
		fmt.Printf("Error during accessing the db: %s\n", err.Error())
		return
	}

	apiRouter := r.PathPrefix("/api").Subrouter()

	api.NewApiServer(gormDB, apiRouter)

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)

	http.Handle("/", r)
	fmt.Println("Server's up and running!")

	go func() {
		<-sc
		fmt.Println("End Server!")
		if err := server.Close(); err != nil {
			fmt.Printf("Error when closing api server: %s\n", err.Error())
		}

	}()

	if err := server.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
		fmt.Printf("HTTP server error: %s\n", err.Error())
	}

}
