package api

import (
	models "aya-backend/db-models"
	"context"
	"errors"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/mux"
	"gorm.io/gorm"
	"net/http"
	"slices"
)

type SessionFilter struct {
	ID       *uint   `json:"data,omitempty"`
	UserID   *uint   `json:"user_id,omitempty"`
	IsOn     *bool   `json:"is_on,omitempty"`
	IsDelete *bool   `json:"is_delete,omitempty"`
	Discord  *string `json:"discord,omitempty"`
	Youtube  *string `json:"youtube,omitempty"`
}

func extractSessionFilter(sessionFilter *SessionFilter) (*models.GORMSession, []string) {
	sessionQuery := models.GORMSession{}

	var args []string
	if sessionFilter.ID != nil {
		sessionQuery.ID = *sessionFilter.ID
		args = append(args, "id")
	}

	if sessionFilter.UserID != nil {
		sessionQuery.UserID = *sessionFilter.UserID
		args = append(args, "user_id")
	}

	if sessionFilter.IsOn != nil {
		sessionQuery.IsOn = *sessionFilter.IsOn
		args = append(args, "is_on")
	}

	if sessionFilter.IsDelete != nil {
		sessionQuery.IsDelete = *sessionFilter.IsDelete
		args = append(args, "is_delete")
	}

	if sessionFilter.Discord != nil {
		sessionQuery.Discord = *sessionFilter.Discord
		args = append(args, "discord")
	}

	if sessionFilter.Youtube != nil {
		sessionQuery.Youtube = *sessionFilter.Youtube
		args = append(args, "youtube")
	}

	return &sessionQuery, args

}

func authSessionOwnerMiddleware(db *gorm.DB) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(writer http.ResponseWriter, req *http.Request) {

			newReqWithContext := req

			sessionFilter := req.Context().Value(CONTEXT_KEY_REQ_FILTER).(*SessionFilter)
			jwtClaim := req.Context().Value(CONTEXT_KEY_JWT_CLAIM).(jwt.MapClaims)

			sessionQuery, args := extractSessionFilter(sessionFilter)

			if !slices.Contains(args, "user_id") {
				writer.WriteHeader(http.StatusForbidden)
				_, _ = writer.Write([]byte(marshalReturnData(nil, "Query filter does not contains required fields")))
				return
			}

			userQuery := models.GORMUser{
				ID: sessionQuery.UserID,
			}

			user := models.GORMUser{}

			userQueryResult := db.Where(&userQuery, "id").First(&user)

			if userQueryResult.Error != nil {
				writer.WriteHeader(http.StatusBadRequest)
				_, _ = writer.Write([]byte(marshalReturnData(nil, "User not found")))
				return
			}

			jwtClaimEmail := jwtClaim["email"].(string)
			userQueryEmail := user.Email

			if jwtClaimEmail != userQueryEmail {
				writer.WriteHeader(http.StatusForbidden)
				_, _ = writer.Write([]byte(marshalReturnData(nil, "User Not Authorized")))
				return
			}

			if !slices.Contains(args, "id") {
				// get the Session content
				getSessionFilter := models.GORMSession{
					ID: sessionQuery.ID,
				}

				session := models.GORMSession{}

				sessionQueryResult := db.
					Where(&getSessionFilter, "id").
					First(&session)

				if errors.Is(sessionQueryResult.Error, gorm.ErrRecordNotFound) {
					writer.WriteHeader(http.StatusBadRequest)
					_, _ = writer.Write([]byte(marshalReturnData(nil, "session id does not exists")))
					return
				}

				if sessionQueryResult.Error != nil {
					writer.WriteHeader(http.StatusInternalServerError)
					_, _ = writer.Write([]byte(marshalReturnData(nil, "Internal Server Error")))
					return
				}

				sessionUserEmail := session.User.Email
				if jwtClaimEmail != sessionUserEmail {
					writer.WriteHeader(http.StatusForbidden)
					_, _ = writer.Write([]byte(marshalReturnData(nil, "User Not Authorize")))
					return
				}

				newReqWithContext = newReqWithContext.WithContext(context.WithValue(newReqWithContext.Context(), CONTEXT_KEY_SESSION, &session))

			}

			newReqWithContext = newReqWithContext.WithContext(context.WithValue(newReqWithContext.Context(), CONTEXT_KEY_USER, &user))

			next.ServeHTTP(writer, newReqWithContext)
		})
	}
}

func (dbApiServer *DBApiServer) NewSessionApi(r *mux.Router) {

	r.Use(inputParsingMiddleware(&SessionFilter{}))

	r.Use(authSessionOwnerMiddleware(dbApiServer.db))

	r.PathPrefix("/").
		Methods(http.MethodGet).
		HandlerFunc(func(writer http.ResponseWriter, req *http.Request) {
			sessionFilter := req.Context().Value(CONTEXT_KEY_REQ_FILTER).(*SessionFilter)

			sessionQuery, args := extractSessionFilter(sessionFilter)

			var sessions []models.GORMSession

			result := dbApiServer.db.
				Where(&sessionQuery, args).
				Find(&sessions)

			if result.Error != nil {
				fmt.Println(result.Error.Error())
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
			sessionFilter := req.Context().Value(CONTEXT_KEY_REQ_FILTER).(*SessionFilter)
			user := req.Context().Value(CONTEXT_KEY_USER).(*models.GORMUser)

			newSession := models.GORMSession{
				UserID:   *sessionFilter.UserID,
				IsOn:     false,
				IsDelete: false,
				Discord:  *sessionFilter.Discord,
				Youtube:  *sessionFilter.Youtube,
				User:     *user,
			}

			result := dbApiServer.db.Create(&newSession)
			if result.Error != nil {
				fmt.Println(result.Error.Error())
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
			sessionFilter := req.Context().Value(CONTEXT_KEY_REQ_FILTER).(*SessionFilter)

			if req.Context().Value(CONTEXT_KEY_SESSION) == nil {
				writer.WriteHeader(http.StatusBadRequest)
				_, _ = writer.Write([]byte(marshalReturnData(nil, "Missing required fields")))
				return
			}

			session := req.Context().Value(CONTEXT_KEY_SESSION).(*models.GORMSession)

			updateFilter := &SessionFilter{
				IsOn:    sessionFilter.IsOn,
				Discord: sessionFilter.Discord,
				Youtube: sessionFilter.Youtube,
			}

			updateSession, args := extractSessionFilter(updateFilter)

			sessionResult := dbApiServer.db.
				Model(&session).
				Select(args).
				Updates(&updateSession)

			if sessionResult.Error != nil {
				fmt.Println(sessionResult.Error.Error())
				writer.WriteHeader(http.StatusInternalServerError)
				_, _ = writer.Write([]byte(marshalReturnData(nil, "Internal Server Error")))
				return
			}

			writer.WriteHeader(http.StatusOK)
			_, _ = writer.Write([]byte(marshalReturnData(session, "")))
			return

		})

	r.PathPrefix("/").
		Methods(http.MethodDelete).
		HandlerFunc(func(writer http.ResponseWriter, req *http.Request) {

			if req.Context().Value(CONTEXT_KEY_SESSION) == nil {
				writer.WriteHeader(http.StatusBadRequest)
				_, _ = writer.Write([]byte(marshalReturnData(nil, "Missing required fields")))
				return
			}

			session := req.Context().Value(CONTEXT_KEY_SESSION).(*models.GORMSession)

			updateSession := &models.GORMSession{
				IsDelete: true,
			}

			sessionResult := dbApiServer.db.
				Model(&session).
				Updates(&updateSession)

			if sessionResult.Error != nil {
				fmt.Println(sessionResult.Error.Error())
				writer.WriteHeader(http.StatusInternalServerError)
				_, _ = writer.Write([]byte(marshalReturnData(nil, "Internal Server Error")))
				return
			}

			writer.WriteHeader(http.StatusOK)
			_, _ = writer.Write([]byte(marshalReturnData(session, "")))
			return
		})
}
