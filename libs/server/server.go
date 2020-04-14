package server

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"github.com/functionalfoundry/graphqlws"
	"github.com/gorilla/mux"
	"github.com/graphql-go/graphql"
	"github.com/graphql-go/handler"
	"locationMicroService/libs/actors"
	"locationMicroService/libs/core"
	"locationMicroService/libs/data"
	"locationMicroService/libs/util"
	"log"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"time"
)

var Db core.Storage

const (
	AuthorizationHeader = "Authorization"
	ENDPOINT            = "ENDPOINT"
	SERVER_ADDRESS      = "SERVER_ADDRESS"
)

func init() {
	var err error
	log.SetFlags(log.Lshortfile | log.LstdFlags)

	Db, err = data.NewDb(os.Getenv(data.DB_USER),
		os.Getenv(data.DB_PASS),
		os.Getenv(data.DB_HOST),
		os.Getenv(data.DB_NAME),
		os.Getenv(data.DB_SSLMODE))
	if err != nil {
		fmt.Println(err)
	}
}

func Start() {
	defer Db.Close()

	schema, _ := graphql.NewSchema(graphql.SchemaConfig{
		Query:        QueryType,
		Mutation:     MutationType,
		Subscription: SubscriptionType,
	})

	subscriptionManager := graphqlws.NewSubscriptionManager(&schema)

	wsHandler := graphqlws.NewHandler(graphqlws.HandlerConfig{
		SubscriptionManager: subscriptionManager,
	})

	h := handler.New(&handler.Config{
		Schema:   &schema,
		Pretty:   true,
		GraphiQL: true,
	})
	r := mux.NewRouter()
	auth := r.PathPrefix("/auth").Subrouter()
	auth.Use(headerAuthorization, disableCORS)
	auth.HandleFunc("/login", endpointLoginAdmin).Methods("POST")
	auth.HandleFunc("/registerUser", endpointRegisterUser).Methods("POST")
	auth.HandleFunc("/getRefreshToken", endpointGetRefreshTokenFromClient).Methods("POST")
	r.Handle(os.Getenv(ENDPOINT), headerAuthorization(disableCORS(h)))
	r.Handle("/subscriptions", wsHandler)

	server := http.Server{
		Addr: os.Getenv(SERVER_ADDRESS),
		// ReadTimeout:  1 * time.Second,
		// WriteTimeout: 1 * time.Second,
		Handler: r,
	}

	ctx, cancel := context.WithCancel(context.Background())

	go handleSubscriptions(subscriptionManager)
	go shutdown(&server, cancel)

	log.Print("Server Started")
	if err := server.ListenAndServe(); err != nil {
		log.Print(err)
	}

	<-ctx.Done()
	log.Print("Server are closed")
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
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET_USER, OPTIONS, PUT, DELETE")
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

func handleSubscriptions(manager graphqlws.SubscriptionManager) {
	println(" implement me")
}

func endpointRegisterUser(w http.ResponseWriter, r *http.Request) {
	headerAuth, ok := r.Context().Value("token").(string)
	if !ok {
		http.Error(w, errInvalidToken.Error(), http.StatusNetworkAuthenticationRequired)
		return
	}

	adminToken, err := getTokenFromHeader(headerAuth)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNetworkAuthenticationRequired)
		return
	}

	claimsMap, err := verifyToken(adminToken)
	if err != nil {
		http.Error(w, errInvalidToken.Error(), http.StatusNetworkAuthenticationRequired)
		return
	}

	adminId, err := getAdminDataFromClaims(claimsMap)
	if err != nil {
		http.Error(w, errInvalidToken.Error(), http.StatusNetworkAuthenticationRequired)
		return
	}

	var userParams map[string]interface{}
	if err = json.NewDecoder(r.Body).Decode(&userParams); err != nil {
		http.Error(w, "couldn't get request body", http.StatusBadRequest)
		return
	}

	var latitude float64
	if latitude, ok = userParams["lat"].(float64); !ok {
		http.Error(w, "lat argument could be missing", http.StatusBadRequest)
		return
	}

	var longitude float64
	if longitude, ok = userParams["long"].(float64); !ok {
		http.Error(w, "long argument could be missing", http.StatusBadRequest)
		return
	}

	var category string
	if category, ok = userParams["category"].(string); !ok {
		http.Error(w, "category argument could be missing", http.StatusBadRequest)
		return
	}

	refreshToken := "mockToken"
	newUser := actors.NewUser(refreshToken, latitude, longitude, category, adminId)
	userId, err := Db.AddUser(newUser)
	if err != nil {
		http.Error(w, "error writing in the database", http.StatusInternalServerError)
		return
	}

	refreshToken, err = updateRefreshToken(userId)
	if err != nil {
		http.Error(w, fmt.Sprintf("error generating refreshToken: %s", err.Error()), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err = json.NewEncoder(w).Encode(refreshToken); err != nil {
		http.Error(w, "error returning headerAuth", http.StatusInternalServerError)
		return
	}
}

func endpointLoginAdmin(w http.ResponseWriter, r *http.Request) {
	userName, password, ok := r.BasicAuth()
	if !ok {
		http.Error(w, "not authorize!", http.StatusUnauthorized)
		return
	}
	admin, err := Db.GetAdmin(userName)
	if err != nil {
		http.Error(w, "couldn't get the admin data from the database", http.StatusInternalServerError)
		return
	}
	if err := util.VerifyPassword(password, admin.PassHash); err != nil {
		http.Error(w, "invalid credentials", http.StatusUnauthorized)
		return
	}

	exp, err := time.ParseDuration(tempTokenExp)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	token, err := createToken(jwt.MapClaims{
		"id":   strconv.Itoa(admin.Id),
		"role": "Admin",
		"iat":  time.Now().Unix(),
		"exp":  time.Now().Add(exp).Unix(),
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err = json.NewEncoder(w).Encode(token); err != nil {
		http.Error(w, "error returning token", http.StatusInternalServerError)
		return
	}
}

func endpointGetRefreshTokenFromClient(w http.ResponseWriter, r *http.Request) {
	tokenStr, err := getTokenFromHeader(r.Context().Value("token").(string))
	if err != nil {
		http.Error(w, errInvalidToken.Error(), http.StatusNetworkAuthenticationRequired)
		return
	}
	claimsMap, err := verifyToken(tokenStr)
	if err != nil {
		http.Error(w, errInvalidToken.Error(), http.StatusNetworkAuthenticationRequired)
		return
	}

	_, err = getAdminDataFromClaims(claimsMap)
	if err != nil {
		http.Error(w, errInvalidToken.Error(), http.StatusNetworkAuthenticationRequired)
		return
	}

	var userId struct {
		Id int `json:"id"`
	}
	if err = json.NewDecoder(r.Body).Decode(&userId); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	user, err := Db.GetUser(userId.Id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	userClaimsMap, err := verifyToken(user.RefreshToken)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if int64(userClaimsMap["exp"].(float64)) >= time.Now().Unix() {
		exp, err := time.ParseDuration(refreshTokenExp)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		user.RefreshToken, err = createToken(jwt.MapClaims{
			"seed":  rand.New(rand.NewSource(time.Now().Unix())),
			"iat":   time.Now().Unix(),
			"exp":   time.Now().Add(exp).Unix(),
			"seed2": user.GeoCord.Latitude + user.GeoCord.Longitude,
		})
		if err != nil {
			http.Error(w, "error generating token", http.StatusInternalServerError)
			return
		}
	}

	w.Header().Set("Content-Type", "application/json")
	if err = json.NewEncoder(w).Encode(user.RefreshToken); err != nil {
		http.Error(w, "error returning token", http.StatusInternalServerError)
		return
	}
}

func updateRefreshToken(userId int) (refreshToken string, err error) {

	exp, err := time.ParseDuration(refreshTokenExp)
	if err != nil {
		return
	}

	refreshToken, err = createToken(jwt.MapClaims{
		"id":   strconv.Itoa(userId),
		"role": "User",
		"type": "RefToken",
		"iat":  time.Now().Unix(),
		"exp":  time.Now().Add(exp).Unix(),
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

	if role != "User" {
		return -1, errInvalidToken
	}

	tokenType, ok := refTokenClaims["type"].(string)
	if !ok {
		return -1, errInvalidToken
	}

	if tokenType != "RefToken" {
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

	if role != "User" {
		return -1, errInvalidToken
	}

	tokenType, ok := tempTokenClaims["type"].(string)
	if !ok {
		return -1, errInvalidToken
	}

	if tokenType != "TempToken" {
		return -1, errInvalidToken
	}

	return
}
