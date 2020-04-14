package server

import (
	"context"
	"fmt"
	"github.com/functionalfoundry/graphqlws"
	"github.com/gorilla/mux"
	"github.com/graphql-go/graphql"
	"github.com/graphql-go/handler"
	"locationMicroService/libs/core"
	"locationMicroService/libs/data"
	"log"
	"net/http"
	"os"
)

var Db core.Storage

const (
	AuthorizationHeader = "Authorization"

	// Environment vars
	ENDPOINT          = "ENDPOINT"
	SERVER_ADDRESS    = "SERVER_ADDRESS"
	REFRESH_TOKEN_EXP = "REFRESH_TOKEN_EXP"
	TEMP_TOKEN_EXP    = "TEMP_TOKEN_EXP"
	SECRET            = "SECRET"

	// Roles
	USER_ROLE  = "User"
	ADMIN_ROLE = "Admin"

	// Token types
	REF_TOKEN_TYPE  = "RefToken"
	TEMP_TOKEN_TYPE = "TempToken"
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
