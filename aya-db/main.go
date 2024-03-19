package main

import (
	"fmt"
	"github.com/joho/godotenv"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"log"
	"os"
	"path"
)

const (
	DB_PATH_ENV     = "DB_PATH"
	DEFAULT_DB_PATH = "./data"
	DB_NAME         = "aya.db"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	fmt.Println("Initializing database")
	dataLocation := os.Getenv(DB_PATH_ENV)
	if dataLocation == "" {
		fmt.Printf("Cannot find %s, using default value (%s) instead!\n", DB_PATH_ENV, DEFAULT_DB_PATH)
		dataLocation = DEFAULT_DB_PATH
	}

	err = os.MkdirAll(dataLocation, 0777)
	if err != nil {
		fmt.Printf("Error during creating %s directory: %s\n", dataLocation, err.Error())
		return
	}

	dataPath := path.Join(dataLocation, DB_NAME)

	_, err = os.Create(dataPath)
	if err != nil {
		fmt.Printf("Error during making %s: %s\n", path.Join(dataLocation, DB_NAME), err.Error())
		return
	}

	db, err := gorm.Open(sqlite.Open(dataPath), &gorm.Config{})
	if err != nil {
		fmt.Printf("Error when connect to db %s: %s\n", dataPath, err.Error())
		return
	}

	var session Session
	err = db.AutoMigrate(&session)
	if err != nil {
		fmt.Printf("Error when migrating the interface Session: %s\n", err.Error())
		return
	}
}
