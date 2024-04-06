package api

import (
	models "aya-backend/db-models"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gorilla/mux"
	"gorm.io/gorm"
	"net/http"
	"strings"
)

type DBApiServer struct {
	db *gorm.DB
}

type UserFilter struct {
	ID    *uint   `json:"id,omitempty"`
	Email *string `json:"email,omitempty"`
}

type SessionFilter struct {
	ID       uint    `json:"data,omitempty"`
	OwnerID  uint    `json:"owner_id,omitempty"`
	IsOn     *bool   `json:"is_on,omitempty"`
	IsDelete *bool   `json:"is_delete,omitempty"`
	Discord  *string `json:"discord,omitempty"`
	Twitch   *string `json:"twitch,omitempty"`
	Youtube  *string `json:"youtube,omitempty"`
}

type Content struct {
	Data any    `json:"data"`
	Err  string `json:"err,omitempty"`
}

func marshalReturnData(data any, errMsg string) string {
	returnData := Content{Data: data}
	if errMsg != "" {
		returnData.Err = errMsg
	}
	returnDataStr, err := json.Marshal(returnData)
	if err != nil {
		return "{}"
	} else {
		return string(returnDataStr)
	}
}

func NewApiServer(db *gorm.DB, r *mux.Router) *DBApiServer {

	dbApiServer := DBApiServer{db: db}

	session := r.PathPrefix("/session").Subrouter()
	dbApiServer.NewSessionApi(session)

	user := r.PathPrefix("/user").Subrouter()
	dbApiServer.NewUserApi(user)

	return &dbApiServer
}

func (dbApiServer *DBApiServer) NewSessionApi(r *mux.Router) {
	r.PathPrefix("/").
		Methods(http.MethodGet).
		HandlerFunc(func(writer http.ResponseWriter, req *http.Request) {
			var sessionFilter SessionFilter

			writer.Header().Set("Content-Type", "application/json")

			content := req.Header.Get("Content-Type")
			if content != "" {
				mediaType := strings.ToLower(strings.TrimSpace(strings.Split(content, ";")[0]))
				if mediaType != "application/json" {
					writer.WriteHeader(http.StatusUnsupportedMediaType)
					_, _ = writer.Write([]byte(marshalReturnData(nil, "Content-Type header is not application/json")))
					return
				}
			}

			err := json.NewDecoder(req.Body).Decode(&sessionFilter)
			if err != nil {
				writer.WriteHeader(http.StatusBadRequest)
				_, _ = writer.Write([]byte(marshalReturnData(nil, "cannot parse payload content")))

				return
			}

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
			var sessionFilter SessionFilter

			writer.Header().Set("Content-Type", "application/json")

			content := req.Header.Get("Content-Type")
			if content != "" {
				mediaType := strings.ToLower(strings.TrimSpace(strings.Split(content, ";")[0]))
				if mediaType != "application/json" {
					writer.WriteHeader(http.StatusUnsupportedMediaType)
					_, _ = writer.Write([]byte(marshalReturnData(nil, "Content-Type header is not application/json")))
					return
				}
			}

			err := json.NewDecoder(req.Body).Decode(&sessionFilter)
			if err != nil {
				writer.WriteHeader(http.StatusBadRequest)
				_, _ = writer.Write([]byte(marshalReturnData(nil, "cannot parse payload content")))
				return
			}

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
		Methods(http.MethodPatch).
		HandlerFunc(func(writer http.ResponseWriter, req *http.Request) {
			var sessionFilter SessionFilter

			writer.Header().Set("Content-Type", "application/json")

			content := req.Header.Get("Content-Type")
			if content != "" {
				mediaType := strings.ToLower(strings.TrimSpace(strings.Split(content, ";")[0]))
				if mediaType != "application/json" {
					writer.WriteHeader(http.StatusUnsupportedMediaType)
					_, _ = writer.Write([]byte(marshalReturnData(nil, "Content-Type header is not application/json")))
					return
				}
			}

			err := json.NewDecoder(req.Body).Decode(&sessionFilter)
			if err != nil {
				writer.WriteHeader(http.StatusBadRequest)
				_, _ = writer.Write([]byte(marshalReturnData(nil, "cannot parse payload content")))
				return
			}

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

func (dbApiServer *DBApiServer) NewUserApi(r *mux.Router) {

	r.PathPrefix("/").
		Methods(http.MethodGet).
		HandlerFunc(func(writer http.ResponseWriter, req *http.Request) {
			var uFilter UserFilter

			writer.Header().Set("Content-Type", "application/json")

			content := req.Header.Get("Content-Type")
			if content != "" {
				mediaType := strings.ToLower(strings.TrimSpace(strings.Split(content, ";")[0]))
				if mediaType != "application/json" {
					writer.WriteHeader(http.StatusUnsupportedMediaType)
					_, _ = writer.Write([]byte(marshalReturnData(nil, "Content-Type header is not application/json")))
					return
				}
			}

			err := json.NewDecoder(req.Body).Decode(&uFilter)
			if err != nil {
				writer.WriteHeader(http.StatusBadRequest)
				_, _ = writer.Write([]byte(marshalReturnData(nil, "cannot parse payload content")))
				return
			}

			user := models.GORMUser{}

			if uFilter.ID != nil {
				user.ID = *uFilter.ID
			}
			if uFilter.Email != nil {
				user.Email = *uFilter.Email
			}

			result := dbApiServer.db.First(&user)
			if errors.Is(result.Error, gorm.ErrRecordNotFound) {
				writer.WriteHeader(http.StatusBadRequest)
				_, _ = writer.Write([]byte(marshalReturnData(nil, "Cannot find the profile")))
				return
			} else if result.Error != nil {
				writer.WriteHeader(http.StatusInternalServerError)
				_, _ = writer.Write([]byte(marshalReturnData(nil, "Internal error")))
				return
			}

			writer.WriteHeader(http.StatusOK)
			_, _ = writer.Write([]byte(marshalReturnData(user, "")))
			return

		})

	r.PathPrefix("/").
		Methods(http.MethodPost).
		HandlerFunc(func(writer http.ResponseWriter, req *http.Request) {
			var uFilter UserFilter

			writer.Header().Set("Content-Type", "application/json")

			content := req.Header.Get("Content-Type")
			if content != "" {
				mediaType := strings.ToLower(strings.TrimSpace(strings.Split(content, ";")[0]))
				if mediaType != "application/json" {
					writer.WriteHeader(http.StatusUnsupportedMediaType)
					_, _ = writer.Write([]byte(marshalReturnData(nil, "Content-Type header is not application/json")))
					return
				}
			}

			err := json.NewDecoder(req.Body).Decode(&uFilter)
			if err != nil {
				writer.WriteHeader(http.StatusBadRequest)
				_, _ = writer.Write([]byte(marshalReturnData(nil, "Cannot parse requested data")))
				return
			}

			user := models.GORMUser{}

			if uFilter.Email != nil {
				user.Email = *uFilter.Email
			}

			result := dbApiServer.db.First(&user)
			if !errors.Is(result.Error, gorm.ErrRecordNotFound) && result.Error != nil {
				writer.WriteHeader(http.StatusInternalServerError)
				_, _ = writer.Write([]byte(marshalReturnData(nil, "Internal error")))
				return
			}

			if result.Error == nil {
				writer.WriteHeader(http.StatusBadRequest)
				_, _ = writer.Write([]byte(marshalReturnData(nil, "User already exists")))
				return
			}

			result = dbApiServer.db.Create(&user)
			if result.Error != nil {
				writer.WriteHeader(http.StatusInternalServerError)
				_, _ = writer.Write([]byte(marshalReturnData(nil, "Internal error")))
				return
			}

			writer.WriteHeader(http.StatusOK)
			_, _ = writer.Write([]byte(marshalReturnData(user, "")))
			return

		})
}
