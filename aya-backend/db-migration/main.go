package main

import (
	models "aya-backend/db-models"
	"fmt"
	"github.com/joho/godotenv"
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

		dataPath = path.Join(dataLocation, DB_NAME)

		err := MakeDirFile(dataPath)
		if err != nil {
			fmt.Printf("Error: Cannot create directory: %s\n", err.Error())
			return
		}

	} else {
		dataPath = sqlDb
	}

	err = DbMigration(dataPath, &models.GORMSession{}, &models.GORMUser{})
	if err != nil {
		fmt.Printf("Error: %s\n", err.Error())
		return
	}

	fmt.Printf("Migrate database successfully! Have fun developing.\n")
}
