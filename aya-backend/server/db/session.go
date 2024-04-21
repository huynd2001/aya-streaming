package db

import (
	models "aya-backend/db-models"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"time"
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

func (infoDB *InfoDB) GetResourcesInfo(registeredSessions map[string]bool, lastUpdated time.Time) map[string][]models.Resource {
	var notPopSessions []string
	var allSessions []string
	for sessionId, isPopulated := range registeredSessions {
		if !isPopulated {
			notPopSessions = append(notPopSessions, sessionId)
		}
		allSessions = append(allSessions, sessionId)
	}
	var sessions []models.GORMSession
	result := infoDB.db.Where("uuid IN ?", notPopSessions).Or("uuid IN ? AND updated_at != NULL AND updated_at >= ?", allSessions, lastUpdated).Find(&sessions)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return map[string][]models.Resource{}
	} else if result.Error != nil {
		fmt.Printf("Unknown error: %s\n", result.Error.Error())
		return map[string][]models.Resource{}
	}

	session2Resources := make(map[string][]models.Resource)
	for _, session := range sessions {
		sessionUUID := session.UUID.String()
		if !session.IsOn {
			session2Resources[sessionUUID] = []models.Resource{}
			continue
		}
		var resources []models.Resource
		err := json.Unmarshal([]byte(session.Resources), &resources)
		if err != nil {
			resources = []models.Resource{}
		}
		session2Resources[sessionUUID] = resources
	}
	return session2Resources
}

func NewInfoDB(db *gorm.DB) *InfoDB {
	return &InfoDB{db: db}
}
