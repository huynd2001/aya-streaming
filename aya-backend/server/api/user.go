package api

import (
	models "aya-backend/db-models"
	"encoding/json"
	"errors"
	"github.com/gorilla/mux"
	"gorm.io/gorm"
	"net/http"
)

type UserFilter struct {
	ID    *uint   `json:"id,omitempty"`
	Email *string `json:"email,omitempty"`
}

func (dbApiServer *DBApiServer) NewUserApi(r *mux.Router) {

	r.Use(getContentParsingHandler(&UserFilter{}))

	r.PathPrefix("/").
		Methods(http.MethodGet).
		HandlerFunc(func(writer http.ResponseWriter, req *http.Request) {
			uFilter := req.Context().Value(contextKey("filter")).(*UserFilter)

			// Skip error handling in this step since it has been done in middleware
			_ = json.NewDecoder(req.Body).Decode(&uFilter)

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
			uFilter := req.Context().Value(contextKey("filter")).(*UserFilter)

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
