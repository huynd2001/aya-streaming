package api

import (
	models "aya-backend/db-models"
	"errors"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/mux"
	"gorm.io/gorm"
	"net/http"
	"slices"
)

type UserFilter struct {
	ID    *uint   `json:"id,omitempty" schema:"id"`
	Email *string `json:"email,omitempty" schema:"email"`
}

func extractUserFilter(userFilter *UserFilter) (*models.GORMUser, []string) {
	userQuery := models.GORMUser{}

	var args []string
	if userFilter.ID != nil {
		userQuery.ID = *userFilter.ID
		args = append(args, "id")
	}

	if userFilter.Email != nil {
		userQuery.Email = *userFilter.Email
		args = append(args, "email")
	}

	return &userQuery, args

}

func authUserOwnerMiddleware(db *gorm.DB) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(writer http.ResponseWriter, req *http.Request) {

			if req.Method == "OPTIONS" {
				next.ServeHTTP(writer, req)
				return
			}

			newReqWithContext := req

			userFilter := req.Context().Value(CONTEXT_KEY_REQ_FILTER).(*UserFilter)
			jwtClaim := req.Context().Value(CONTEXT_KEY_JWT_CLAIM).(jwt.MapClaims)

			userQuery, args := extractUserFilter(userFilter)

			if !slices.Contains(args, "email") {
				writer.Header().Set("Content-Type", "application/json")
				writer.WriteHeader(http.StatusForbidden)
				_, _ = writer.Write([]byte(marshalReturnData(nil, "Query filter does not contains required fields")))
				return
			}

			jwtClaimEmail := jwtClaim["email"].(string)

			if jwtClaimEmail != userQuery.Email {
				writer.Header().Set("Content-Type", "application/json")
				writer.WriteHeader(http.StatusUnauthorized)
				_, _ = writer.Write([]byte(marshalReturnData(nil, "Unauthorized Bearer Token")))
				return
			}

			next.ServeHTTP(writer, newReqWithContext)
		})
	}
}

func (dbApiServer *DBApiServer) NewUserApi(r *mux.Router) {

	r.Use(inputParsingMiddleware(func() any {
		return &UserFilter{}
	}))
	r.Use(authUserOwnerMiddleware(dbApiServer.db))

	r.PathPrefix("/").
		Methods(http.MethodOptions).
		HandlerFunc(func(writer http.ResponseWriter, req *http.Request) {
			writer.Header().Set("Allow", "OPTIONS, GET, POST")
			writer.WriteHeader(http.StatusNoContent)
			return
		})

	r.PathPrefix("/").
		Methods(http.MethodGet).
		HandlerFunc(func(writer http.ResponseWriter, req *http.Request) {
			userFilter := req.Context().Value(CONTEXT_KEY_REQ_FILTER).(*UserFilter)

			userQuery, args := extractUserFilter(userFilter)
			var user models.GORMUser

			result := dbApiServer.db.Where(&userQuery, args).First(&user)

			if errors.Is(result.Error, gorm.ErrRecordNotFound) {
				writer.Header().Set("Content-Type", "application/json")
				writer.WriteHeader(http.StatusBadRequest)
				_, _ = writer.Write([]byte(marshalReturnData(nil, "Cannot find the profile")))
				return
			}
			if result.Error != nil {
				fmt.Println(result.Error.Error())
				writer.Header().Set("Content-Type", "application/json")
				writer.WriteHeader(http.StatusInternalServerError)
				_, _ = writer.Write([]byte(marshalReturnData(nil, "Internal error")))
				return
			}

			writer.Header().Set("Content-Type", "application/json")
			writer.WriteHeader(http.StatusOK)
			_, _ = writer.Write([]byte(marshalReturnData(user, "")))
			return

		})

	r.PathPrefix("/").
		Methods(http.MethodPost).
		HandlerFunc(func(writer http.ResponseWriter, req *http.Request) {
			userFilter := req.Context().Value(CONTEXT_KEY_REQ_FILTER).(*UserFilter)

			userQuery := models.GORMUser{
				Email: *userFilter.Email,
			}
			var user models.GORMUser

			result := dbApiServer.db.
				Where(&userQuery, "email").
				First(&user)

			if result.Error == nil {
				writer.Header().Set("Content-Type", "application/json")
				writer.WriteHeader(http.StatusBadRequest)
				_, _ = writer.Write([]byte(marshalReturnData(nil, "User already exists")))
				return
			}

			if !errors.Is(result.Error, gorm.ErrRecordNotFound) {
				fmt.Println(result.Error.Error())
				writer.Header().Set("Content-Type", "application/json")
				writer.WriteHeader(http.StatusInternalServerError)
				_, _ = writer.Write([]byte(marshalReturnData(nil, "Internal error")))
				return
			}

			newUser := models.GORMUser{Email: *userFilter.Email}

			result = dbApiServer.db.Create(&newUser)
			if result.Error != nil {
				fmt.Println(result.Error.Error())
				writer.Header().Set("Content-Type", "application/json")
				writer.WriteHeader(http.StatusInternalServerError)
				_, _ = writer.Write([]byte(marshalReturnData(nil, "Internal error")))
				return
			}

			writer.Header().Set("Content-Type", "application/json")
			writer.WriteHeader(http.StatusOK)
			_, _ = writer.Write([]byte(marshalReturnData(newUser, "")))
			return

		})
}
