package main

import (
	"errors"
	"fmt"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"os"
	"path/filepath"
)

func MakeDirFile(dbLocation string) error {
	if _, err := os.Stat(dbLocation); os.IsNotExist(err) {
		fmt.Printf("Create a new file %s", dbLocation)
		err := os.MkdirAll(filepath.Dir(dbLocation), os.ModePerm)
		if err != nil {
			return err
		}
		_, err = os.Create(dbLocation)
		if err != nil {
			return err
		}
	} else {
		fmt.Printf("File %s already exists", dbLocation)
	}
	return nil
}

func DbMigration(dataLocation string, models ...any) error {
	db, err := gorm.Open(sqlite.Open(dataLocation), &gorm.Config{})
	if err != nil {
		return err
	}
	var errList []error
	for _, model := range models {
		err = db.AutoMigrate(model)
		if err != nil {
			errList = append(errList, err)
		}
	}
	return errors.Join(errList...)
}
