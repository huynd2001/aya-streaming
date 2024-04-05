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
	Email string
}

type Content struct {
	data any
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
		Methods("GET").
		HandlerFunc(func(writer http.ResponseWriter, req *http.Request) {

		})
}

func (dbApiServer *DBApiServer) NewUserApi(r *mux.Router) {

	r.PathPrefix("/").
		Methods("GET").
		HandlerFunc(func(writer http.ResponseWriter, req *http.Request) {
			var uFilter UserFilter
			err := json.NewDecoder(req.Body).Decode(&uFilter)
			if err != nil {
				http.Error(writer, fmt.Sprintf("error: filter format is not corrent"), http.StatusBadRequest)
				return
			}

			user := models.User{
				Email: uFilter.Email,
			}

			result := dbApiServer.db.FirstOrCreate(&user)
			if errors.Is(result.Error, gorm.ErrRecordNotFound) {
				http.Error(writer, "Cannot find the profile", http.StatusBadRequest)
				return
			} else if result.Error != nil {
				http.Error(writer, "Internal error", http.StatusInternalServerError)
				return
			}

			content := req.Header.Get("Content-Type")
			if content != "" {
				mediaType := strings.ToLower(strings.TrimSpace(strings.Split(content, ";")[0]))
				if mediaType != "application/json" {
					http.Error(writer, "Content-Type header is not application/json", http.StatusUnsupportedMediaType)
					return
				}
			}

			returnContent := Content{data: user}

			returnJson, err := json.Marshal(returnContent)
			if err != nil {
				fmt.Printf("Error: cannot parse the user info:\n%s\n", err.Error())
				returnJson = []byte("{}")
			}

			writer.Header().Set("Content-Type", "application/json")
			_, err = writer.Write(returnJson)
			if err != nil {
				fmt.Printf("Error during response: %s\n", err.Error())
			}

		})
}
