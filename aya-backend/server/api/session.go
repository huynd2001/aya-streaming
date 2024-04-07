package api

import (
	models "aya-backend/db-models"
	"errors"
	"fmt"
	"github.com/gorilla/mux"
	"gorm.io/gorm"
	"net/http"
)

type SessionFilter struct {
	ID       uint    `json:"data,omitempty"`
	OwnerID  uint    `json:"owner_id,omitempty"`
	IsOn     *bool   `json:"is_on,omitempty"`
	IsDelete *bool   `json:"is_delete,omitempty"`
	Discord  *string `json:"discord,omitempty"`
	Twitch   *string `json:"twitch,omitempty"`
	Youtube  *string `json:"youtube,omitempty"`
}

func (dbApiServer *DBApiServer) NewSessionApi(r *mux.Router) {

	r.Use(getContentParsingHandler(&SessionFilter{}))

	r.PathPrefix("/").
		Methods(http.MethodGet).
		HandlerFunc(func(writer http.ResponseWriter, req *http.Request) {
			sessionFilter := req.Context().Value(contextKey("filter")).(*SessionFilter)

			sessionQuery := models.GORMSession{
				ID:       sessionFilter.ID,
				UserID:   sessionFilter.OwnerID,
				IsDelete: false,
			}

			var sessions []models.GORMSession

			result := dbApiServer.db.
				Where(&sessionQuery, "id", "owner_id", "is_delete").
				Find(&sessions)

			if result.Error != nil {
				writer.WriteHeader(http.StatusInternalServerError)
				_, _ = writer.Write([]byte(marshalReturnData(nil, "Internal Server Error")))
				return
			}

			writer.WriteHeader(http.StatusOK)
			_, _ = writer.Write([]byte(marshalReturnData(sessions, "")))
			return

		})

	r.PathPrefix("/").
		Methods(http.MethodPost).
		HandlerFunc(func(writer http.ResponseWriter, req *http.Request) {
			sessionFilter := req.Context().Value(contextKey("filter")).(*SessionFilter)

			user := models.GORMUser{
				ID: sessionFilter.OwnerID,
			}

			userResult := dbApiServer.db.First(&user)
			if !errors.Is(userResult.Error, gorm.ErrRecordNotFound) && userResult.Error != nil {
				writer.WriteHeader(http.StatusInternalServerError)
				_, _ = writer.Write([]byte(marshalReturnData(nil, "Internal Server Error")))
				return
			}

			if errors.Is(userResult.Error, gorm.ErrRecordNotFound) {
				writer.WriteHeader(http.StatusBadRequest)
				_, _ = writer.Write([]byte(marshalReturnData(nil, "Request does not contains proper user Id!")))
				return
			}

			sessionQuery := models.GORMSession{
				ID:       sessionFilter.ID,
				IsDelete: false,
			}

			var session models.GORMSession

			result := dbApiServer.db.
				Where(&sessionQuery, "id", "is_delete").
				First(&session)

			if !errors.Is(result.Error, gorm.ErrRecordNotFound) && result.Error != nil {
				writer.WriteHeader(http.StatusInternalServerError)
				_, _ = writer.Write([]byte(marshalReturnData(nil, "Internal Server Error")))
				return
			}

			if result.Error == nil {
				writer.WriteHeader(http.StatusBadRequest)
				_, _ = writer.Write([]byte(marshalReturnData(nil, "The item already exists! The operation would override the item!")))
				return
			}

			newSession := models.GORMSession{
				ID:       sessionFilter.ID,
				UserID:   sessionFilter.OwnerID,
				IsOn:     false,
				IsDelete: false,
				Discord:  *sessionFilter.Discord,
				Youtube:  *sessionFilter.Youtube,
				User:     user,
			}

			result = dbApiServer.db.Create(&newSession)
			if result.Error != nil {
				writer.WriteHeader(http.StatusInternalServerError)
				_, _ = writer.Write([]byte(marshalReturnData(nil, "Internal Server Error")))
				return
			}

			writer.WriteHeader(http.StatusCreated)
			_, _ = writer.Write([]byte(marshalReturnData(newSession, "")))
			return

		})

	r.PathPrefix("/").
		Methods(http.MethodPut).
		HandlerFunc(func(writer http.ResponseWriter, req *http.Request) {
			sessionFilter := req.Context().Value(contextKey("filter")).(*SessionFilter)

			sessionQuery := models.GORMSession{
				ID:       sessionFilter.ID,
				IsDelete: false,
			}

			var session models.GORMSession

			result := dbApiServer.db.
				Where(sessionQuery, "id", "is_delete").
				First(&session)

			if !errors.Is(result.Error, gorm.ErrRecordNotFound) && result.Error != nil {
				writer.WriteHeader(http.StatusInternalServerError)
				_, _ = writer.Write([]byte(marshalReturnData(nil, "Internal Server Error")))
				return
			}

			if errors.Is(result.Error, gorm.ErrRecordNotFound) {
				writer.WriteHeader(http.StatusBadRequest)
				_, _ = writer.Write([]byte(marshalReturnData(nil, "Cannot find the requested item!")))
				return
			}

			if sessionFilter.IsOn != nil {
				session.IsOn = *sessionFilter.IsOn
			}

			if sessionFilter.IsDelete != nil {
				session.IsDelete = *sessionFilter.IsDelete
			}

			if sessionFilter.Discord != nil {
				session.Discord = *sessionFilter.Discord
			}

			if sessionFilter.Youtube != nil {
				session.Youtube = *sessionFilter.Youtube
			}

			result = dbApiServer.db.Save(&session)

			if result.Error != nil {
				fmt.Println(result.Error.Error())
				writer.WriteHeader(http.StatusInternalServerError)
				_, _ = writer.Write([]byte(marshalReturnData(nil, "Internal Server Error")))
				return
			}

			writer.WriteHeader(http.StatusOK)
			_, _ = writer.Write([]byte(marshalReturnData(session, "")))
			return

		})
}
