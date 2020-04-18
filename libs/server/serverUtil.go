package server

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"time"
)

type serverConfig struct {
	graphQlAPIEndpoint string
	serverAddress      string
	// Token
	refreshTokenExp time.Duration
	tempTokenExp    time.Duration
	secret          string
	// TLS
	certPath string
	keyPath  string
}

var config serverConfig
var once sync.Once

func init() {
	once.Do(func() {
		config.serverAddress = os.Getenv(SERVER_ADDRESS)
		if config.serverAddress == "" {
			panic(fmt.Sprintf("%s is empty", SERVER_ADDRESS))
		}

		config.graphQlAPIEndpoint = os.Getenv(ENDPOINT)
		if config.graphQlAPIEndpoint == "" {
			panic(fmt.Sprintf("%s is empty", ENDPOINT))
		}

		config.certPath = os.Getenv(TLS_CERT_PATH)
		config.keyPath = os.Getenv(TLS_KEY_PATH)

		config.secret = os.Getenv(SECRET)
		if config.secret == "" {
			panic(fmt.Sprintf("%s is empty", SECRET))
		}

		duration, err := time.ParseDuration(os.Getenv(REFRESH_TOKEN_EXP))
		if err != nil {
			panic(err)
		}
		config.refreshTokenExp = duration

		duration, err = time.ParseDuration(os.Getenv(TEMP_TOKEN_EXP))
		if err != nil {
			panic(err)
		}

		config.tempTokenExp = duration
	})
}

func updateRefreshToken(userId int) (refreshToken string, err error) {
	refreshToken, err = createToken(jwt.MapClaims{
		"id":   strconv.Itoa(userId),
		"role": USER_ROLE,
		"type": REF_TOKEN_TYPE,
		"iat":  time.Now().Unix(),
		"exp":  time.Now().Add(config.refreshTokenExp).Unix(),
	})
	if err != nil {
		return
	}

	err = Db.UpdateRefreshToken(userId, refreshToken)
	return
}

func getAdminDataFromClaims(adminClaims jwt.MapClaims) (int, error) {
	idStr, ok := adminClaims["id"].(string)
	if !ok {
		return -1, errInvalidToken
	}

	id, err := strconv.Atoi(idStr)
	if err != nil {
		return -1, errInvalidToken
	}
	role, ok := adminClaims["role"].(string)
	if !ok {
		return -1, errInvalidToken
	}

	if role != "Admin" {
		return -1, errInvalidToken
	}

	return id, nil
}

func getRefreshTokenDataFromUserClaims(refTokenClaims jwt.MapClaims) (id int, err error) {
	idStr, ok := refTokenClaims["id"].(string)
	if !ok {
		return -1, errInvalidToken
	}

	id, err = strconv.Atoi(idStr)
	if err != nil {
		return -1, errInvalidToken
	}

	role, ok := refTokenClaims["role"].(string)
	if !ok {
		return -1, errInvalidToken
	}

	if role != USER_ROLE {
		return -1, errInvalidToken
	}

	tokenType, ok := refTokenClaims["type"].(string)
	if !ok {
		return -1, errInvalidToken
	}

	if tokenType != REF_TOKEN_TYPE {
		return -1, errInvalidToken
	}

	return
}

func getTempTokenDataFromUserClaims(tempTokenClaims jwt.MapClaims) (id int, err error) {
	idStr, ok := tempTokenClaims["id"].(string)
	if !ok {
		return -1, errInvalidToken
	}

	id, err = strconv.Atoi(idStr)
	if err != nil {
		return -1, errInvalidToken
	}

	role, ok := tempTokenClaims["role"].(string)
	if !ok {
		return -1, errInvalidToken
	}

	if role != USER_ROLE {
		return -1, errInvalidToken
	}

	tokenType, ok := tempTokenClaims["type"].(string)
	if !ok {
		return -1, errInvalidToken
	}

	if tokenType != TEMP_TOKEN_TYPE {
		return -1, errInvalidToken
	}

	return
}

func headerAuthorization(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var ctx = context.WithValue(r.Context(), "token", r.Header.Get(AuthorizationHeader))
		h.ServeHTTP(w, r.WithContext(ctx))
	})
}

func disableCORS(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		w.Header().Set("Access-Control-Allow-Headers", "Accept, Authorization, Content-Type, "+
			"Content-Length, Accept-Encoding")
		if r.Method == "OPTIONS" {
			w.Header().Set("Access-Control-Max-Age", "86400")
			w.WriteHeader(http.StatusOK)
		}
		h.ServeHTTP(w, r)
	})
}

func shutdown(s *http.Server, cancelFunc context.CancelFunc) {
	sigint := make(chan os.Signal, 1)
	signal.Notify(sigint, os.Kill, os.Interrupt)
	<-sigint
	log.Print("Server Shutdown....")
	if err := s.Shutdown(context.Background()); err != nil {
		log.Fatal(err)
	}
	cancelFunc()
}

func getHttpErr(srtErr string, httpStatus int) json.RawMessage {
	b, _ := json.Marshal(struct {
		Error      string `json:"err"`
		HttpStatus int    `json:"http_status"`
	}{
		Error: srtErr, HttpStatus: httpStatus,
	})
	return b
}
