package db

import (
	models "aya-backend/db-models"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type InfoDB struct {
	db *gorm.DB
}

func (infoDB *InfoDB) GetResourcesOfSession(sessionId string) []models.Resource {
	sessionUUID, err := uuid.Parse(sessionId)
	if err != nil {
		return []models.Resource{}
	}

	session := models.GORMSession{
		UUID: sessionUUID,
		IsOn: true,
	}

	result := infoDB.db.
		Where(&session, "uuid", "is_on").
		First(&session)

	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return []models.Resource{}
	}
	if result.Error != nil {
		fmt.Printf("Unknown error: %s\n", result.Error.Error())
		return []models.Resource{}
	}
	var resources []models.Resource
	err = json.Unmarshal([]byte(session.Resources), &resources)
	if err != nil {
		return []models.Resource{}
	}
	return resources
}

func NewInfoDB(db *gorm.DB) *InfoDB {
	return &InfoDB{db: db}
}
