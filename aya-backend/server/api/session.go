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
	ID        *uint   `json:"id,omitempty" schema:"id"`
	UserID    *uint   `json:"user_id,omitempty" schema:"user_id"`
	IsOn      *bool   `json:"is_on,omitempty" schema:"is_on"`
	Resources *string `json:"resources,omitempty" schema:"resources"`
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

	if sessionFilter.Resources != nil {
		sessionQuery.Resources = *sessionFilter.Resources
		args = append(args, "resources")
	}

	return &sessionQuery, args

}

func authSessionOwnerMiddleware(db *gorm.DB) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(writer http.ResponseWriter, req *http.Request) {

			if req.Method == "OPTIONS" {
				next.ServeHTTP(writer, req)
				return
			}

			newReqWithContext := req

			sessionFilter, ok := req.Context().Value(CONTEXT_KEY_REQ_FILTER).(*SessionFilter)
			if !ok {
				writer.Header().Set("Content-Type", "application/json")
				writer.WriteHeader(http.StatusBadRequest)
				_, _ = writer.Write([]byte(marshalReturnData(nil, "session filter is required")))
				return
			}

			jwtClaim, ok := req.Context().Value(CONTEXT_KEY_JWT_CLAIM).(jwt.MapClaims)
			if !ok {
				writer.Header().Set("Content-Type", "application/json")
				writer.WriteHeader(http.StatusBadRequest)
				_, _ = writer.Write([]byte(marshalReturnData(nil, "jwt claim is required")))
				return
			}

			sessionQuery, args := extractSessionFilter(sessionFilter)

			if !slices.Contains(args, "user_id") {
				writer.Header().Set("Content-Type", "application/json")
				writer.WriteHeader(http.StatusForbidden)
				_, _ = writer.Write([]byte(marshalReturnData(nil, "Query filter does not contains required fields")))
				return
			}

			user := models.GORMUser{
				ID: sessionQuery.UserID,
			}

			userQueryResult := db.First(&user)

			if userQueryResult.Error != nil {
				writer.Header().Set("Content-Type", "application/json")
				writer.WriteHeader(http.StatusBadRequest)
				_, _ = writer.Write([]byte(marshalReturnData(nil, "User not found")))
				return
			}

			jwtClaimEmail := jwtClaim["email"].(string)
			userQueryEmail := user.Email

			if jwtClaimEmail != userQueryEmail {
				writer.Header().Set("Content-Type", "application/json")
				writer.WriteHeader(http.StatusForbidden)
				_, _ = writer.Write([]byte(marshalReturnData(nil, "User Not Authorized")))
				return
			}

			if slices.Contains(args, "id") {
				// get the Session content

				session := models.GORMSession{
					ID: sessionQuery.ID,
				}

				sessionQueryResult := db.
					First(&session)

				if errors.Is(sessionQueryResult.Error, gorm.ErrRecordNotFound) {
					writer.Header().Set("Content-Type", "application/json")
					writer.WriteHeader(http.StatusBadRequest)
					_, _ = writer.Write([]byte(marshalReturnData(nil, "session id does not exists")))
					return
				}

				if sessionQueryResult.Error != nil {
					fmt.Println(sessionQueryResult.Error.Error())
					writer.Header().Set("Content-Type", "application/json")
					writer.WriteHeader(http.StatusInternalServerError)
					_, _ = writer.Write([]byte(marshalReturnData(nil, "Internal Server Error")))
					return
				}

				sessionUserID := session.UserID
				if user.ID != sessionUserID {
					writer.Header().Set("Content-Type", "application/json")
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

	r.Use(inputParsingMiddleware(func() any {
		return &SessionFilter{}
	}))
	r.Use(authSessionOwnerMiddleware(dbApiServer.db))

	r.PathPrefix("/").
		Methods(http.MethodOptions).
		HandlerFunc(func(writer http.ResponseWriter, req *http.Request) {
			writer.Header().Set("Allow", "OPTIONS, GET, POST, PUT, DELETE")
			writer.WriteHeader(http.StatusNoContent)
		})

	r.PathPrefix("/").
		Methods(http.MethodGet).
		HandlerFunc(func(writer http.ResponseWriter, req *http.Request) {
			sessionFilter, ok := req.Context().Value(CONTEXT_KEY_REQ_FILTER).(*SessionFilter)
			if !ok {
				writer.Header().Set("Content-Type", "application/json")
				writer.WriteHeader(http.StatusBadRequest)
				_, _ = writer.Write([]byte(marshalReturnData(nil, "session filter is required")))
				return
			}

			sessionQuery, args := extractSessionFilter(sessionFilter)

			var sessions []models.GORMSession

			result := dbApiServer.db.
				Where(&sessionQuery, args).
				Find(&sessions)

			if result.Error != nil {
				fmt.Println(result.Error.Error())
				writer.Header().Set("Content-Type", "application/json")
				writer.WriteHeader(http.StatusInternalServerError)
				_, _ = writer.Write([]byte(marshalReturnData(nil, "Internal Server Error")))
				return
			}

			writer.Header().Set("Content-Type", "application/json")
			writer.WriteHeader(http.StatusOK)
			_, _ = writer.Write([]byte(marshalReturnData(sessions, "")))

		})

	r.PathPrefix("/").
		Methods(http.MethodPost).
		HandlerFunc(func(writer http.ResponseWriter, req *http.Request) {
			sessionFilter, ok := req.Context().Value(CONTEXT_KEY_REQ_FILTER).(*SessionFilter)
			if !ok {
				writer.Header().Set("Content-Type", "application/json")
				writer.WriteHeader(http.StatusBadRequest)
				_, _ = writer.Write([]byte(marshalReturnData(nil, "session filter is required")))
				return
			}

			user, ok := req.Context().Value(CONTEXT_KEY_USER).(*models.GORMUser)
			if !ok {
				writer.Header().Set("Content-Type", "application/json")
				writer.WriteHeader(http.StatusBadRequest)
				_, _ = writer.Write([]byte(marshalReturnData(nil, "user not found!")))
				return
			}

			newSession := models.GORMSession{
				UserID:    *sessionFilter.UserID,
				IsOn:      false,
				Resources: *sessionFilter.Resources,
				User:      *user,
			}

			result := dbApiServer.db.Create(&newSession)
			if result.Error != nil {
				fmt.Println(result.Error.Error())
				writer.Header().Set("Content-Type", "application/json")
				writer.WriteHeader(http.StatusInternalServerError)
				_, _ = writer.Write([]byte(marshalReturnData(nil, "Internal Server Error")))
				return
			}

			writer.Header().Set("Content-Type", "application/json")
			writer.WriteHeader(http.StatusOK)
			_, _ = writer.Write([]byte(marshalReturnData(newSession, "")))

		})

	r.PathPrefix("/").
		Methods(http.MethodPut).
		HandlerFunc(func(writer http.ResponseWriter, req *http.Request) {
			sessionFilter, ok := req.Context().Value(CONTEXT_KEY_REQ_FILTER).(*SessionFilter)
			if !ok {
				writer.Header().Set("Content-Type", "application/json")
				writer.WriteHeader(http.StatusBadRequest)
				_, _ = writer.Write([]byte(marshalReturnData(nil, "session filter is required")))
				return
			}

			session, ok := req.Context().Value(CONTEXT_KEY_SESSION).(*models.GORMSession)

			if !ok {
				writer.Header().Set("Content-Type", "application/json")
				writer.WriteHeader(http.StatusBadRequest)
				_, _ = writer.Write([]byte(marshalReturnData(nil, "session id is required")))
				return
			}

			updateFilter := &SessionFilter{
				IsOn:      sessionFilter.IsOn,
				Resources: sessionFilter.Resources,
			}

			updateSession, args := extractSessionFilter(updateFilter)

			sessionResult := dbApiServer.db.
				Model(&session).
				Select(args).
				Updates(&updateSession)

			if sessionResult.Error != nil {
				fmt.Println(sessionResult.Error.Error())
				writer.Header().Set("Content-Type", "application/json")
				writer.WriteHeader(http.StatusInternalServerError)
				_, _ = writer.Write([]byte(marshalReturnData(nil, "Internal Server Error")))
				return
			}

			writer.Header().Set("Content-Type", "application/json")
			writer.WriteHeader(http.StatusOK)
			_, _ = writer.Write([]byte(marshalReturnData(session, "")))

		})

	r.PathPrefix("/").
		Methods(http.MethodDelete).
		HandlerFunc(func(writer http.ResponseWriter, req *http.Request) {

			session, ok := req.Context().Value(CONTEXT_KEY_SESSION).(*models.GORMSession)

			if !ok {
				writer.Header().Set("Content-Type", "application/json")
				writer.WriteHeader(http.StatusBadRequest)
				_, _ = writer.Write([]byte(marshalReturnData(nil, "Missing required fields")))
				return
			}

			sessionResult := dbApiServer.db.
				Delete(&session)

			if sessionResult.Error != nil {
				fmt.Println(sessionResult.Error.Error())
				writer.Header().Set("Content-Type", "application/json")
				writer.WriteHeader(http.StatusInternalServerError)
				_, _ = writer.Write([]byte(marshalReturnData(nil, "Internal Server Error")))
				return
			}

			writer.Header().Set("Content-Type", "application/json")
			writer.WriteHeader(http.StatusOK)
			_, _ = writer.Write([]byte(marshalReturnData(session, "")))

		})

	fmt.Println("Finished setting up /session")
}
