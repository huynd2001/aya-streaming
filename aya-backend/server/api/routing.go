package api

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/MicahParks/keyfunc/v3"
	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/mux"
	"github.com/gorilla/schema"
	"gorm.io/gorm"
	"net/http"
	"os"
	"strings"
)

type DBApiServer struct {
	db *gorm.DB
}

type Content struct {
	Data any    `json:"data,omitempty"`
	Err  string `json:"err,omitempty"`
}

const (
	AUTH_JWKS_ENDPOINT_ENV = "AUTH_JWKS_ENDPOINT"
)

var (
	authJwksEndpoint string
)

type contextKey int

const (
	CONTEXT_KEY_JWT_CLAIM contextKey = iota
	CONTEXT_KEY_REQ_FILTER
	CONTEXT_KEY_USER
	CONTEXT_KEY_SESSION
)

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

// inputParsingMiddleware resolves the filter data
func inputParsingMiddleware(dataModel any) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(writer http.ResponseWriter, req *http.Request) {

			writer.Header().Set("Content-Type", "application/json")

			if req.Method != http.MethodGet {
				content := req.Header.Get("Content-Type")
				if content != "" {
					mediaType := strings.ToLower(strings.TrimSpace(strings.Split(content, ";")[0]))
					if mediaType != "application/json" {
						writer.WriteHeader(http.StatusUnsupportedMediaType)
						_, _ = writer.Write([]byte(marshalReturnData(nil, "Content-Type header is not application/json")))
						return
					}
				}

				err := json.NewDecoder(req.Body).Decode(dataModel)
				if err != nil {
					writer.WriteHeader(http.StatusBadRequest)
					_, _ = writer.Write([]byte(marshalReturnData(nil, "Cannot parse payload content")))
					return
				}

			} else {
				reqQuery := req.URL.Query()
				var decoder = schema.NewDecoder()
				err := decoder.Decode(dataModel, reqQuery)
				if err != nil {
					writer.WriteHeader(http.StatusBadRequest)
					_, _ = writer.Write([]byte(marshalReturnData(nil, "Cannot parse payload content")))
					return
				}
			}

			reqWithFilter := req.WithContext(context.WithValue(req.Context(), CONTEXT_KEY_REQ_FILTER, dataModel))
			next.ServeHTTP(writer, reqWithFilter)

		})
	}
}

func jwtAuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, req *http.Request) {

		writer.Header().Set("Content-Type", "application/json")

		bearerTokenStr := req.Header.Get("Authorization")
		jwtStr := strings.TrimPrefix(bearerTokenStr, "Bearer ")

		jwkFunc, err := keyfunc.NewDefaultCtx(context.Background(), []string{authJwksEndpoint})
		if err != nil {
			writer.WriteHeader(http.StatusUnauthorized)
			_, _ = writer.Write([]byte(marshalReturnData(nil, "Unauthorized bearer token!")))
			return
		}

		token, err := jwt.Parse(jwtStr, jwkFunc.Keyfunc)

		if err != nil {

			writer.WriteHeader(http.StatusUnauthorized)

			switch {
			case errors.Is(err, jwt.ErrSignatureInvalid):
				_, _ = writer.Write([]byte(marshalReturnData(nil, "Invalid Signature!")))
				return
			case errors.Is(err, jwt.ErrTokenExpired) || errors.Is(err, jwt.ErrTokenNotValidYet):
				_, _ = writer.Write([]byte(marshalReturnData(nil, "Token expired!")))
				return
			default:
				_, _ = writer.Write([]byte(marshalReturnData(nil, "Unauthorized bearer token!")))
				return
			}
		}

		if !token.Valid {
			writer.WriteHeader(http.StatusUnauthorized)
			_, _ = writer.Write([]byte(marshalReturnData(nil, "Unauthorized bearer token!")))
		}

		reqWithAuthorization := req.WithContext(context.WithValue(req.Context(), contextKey(CONTEXT_KEY_JWT_CLAIM), token.Claims.(jwt.MapClaims)))

		next.ServeHTTP(writer, reqWithAuthorization)
	})
}

func NewApiServer(db *gorm.DB, r *mux.Router) *DBApiServer {

	dbApiServer := DBApiServer{db: db}
	authJwksEndpoint = os.Getenv(AUTH_JWKS_ENDPOINT_ENV)

	r.Use(jwtAuthMiddleware)

	session := r.PathPrefix("/session").Subrouter()
	dbApiServer.NewSessionApi(session)

	user := r.PathPrefix("/user").Subrouter()
	dbApiServer.NewUserApi(user)

	return &dbApiServer
}
