package main

import (
	model "aya-backend/db-models"
	"errors"
	"fmt"
	"github.com/joho/godotenv"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"os"
	"path"
)

const (
	DB_PATH_ENV     = "DB_PATH"
	SQL_DB_PATH_ENV = "SQL_DB_PATH"
	DEFAULT_DB_PATH = "./data"
	DB_NAME         = "aya.db"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		fmt.Printf("Error: Cannot load .env file: %s\n", err.Error())
	}

	fmt.Println("Initializing database")
	var dataPath string
	sqlDb := os.Getenv(SQL_DB_PATH_ENV)
	if sqlDb == "" {
		fmt.Printf("Cannot find %s, looking to setting up local db from %s\n", SQL_DB_PATH_ENV, DB_PATH_ENV)

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

		dataPath = path.Join(dataLocation, DB_NAME)

		if _, err := os.Stat(dataPath); errors.Is(err, os.ErrNotExist) {
			fmt.Printf("%s not found! Start initializing db...\n", dataPath)
			_, err2 := os.Create(dataPath)
			if err2 != nil {
				fmt.Printf("Error during making %s: %s\n", path.Join(dataLocation, DB_NAME), err.Error())
				return
			}
		} else {
			fmt.Printf("%s found!\n", dataPath)
		}

	} else {
		dataPath = sqlDb
	}

	db, err := gorm.Open(sqlite.Open(dataPath), &gorm.Config{})
	if err != nil {
		fmt.Printf("Error when connect to db %s: %s\n", dataPath, err.Error())
		return
	}

	var session model.Session
	err = db.AutoMigrate(&session)
	if err != nil {
		fmt.Printf("Error when migrating the interface Session: %s\n", err.Error())
		return
	}

	fmt.Printf("Migrate database successfully! Have fun developing.")
}
