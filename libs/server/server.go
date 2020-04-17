package server

import (
	"context"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/graphql-go/graphql"
	"github.com/graphql-go/handler"
	"github.com/sirupsen/logrus"
	"github.com/stdevHsequeda/graphqlws"
	"locationMicroService/libs/actors"
	"locationMicroService/libs/core"
	"locationMicroService/libs/data"
	"log"
	"net/http"
	"os"
	"sync"
	"time"
)

// Db is the instance of the database
var Db core.Storage

// updateCh is the channel for update the subscriptions
var updateCh chan actors.User

// mut is a mutex to prevent the race conditions with updateCh
var mut sync.Mutex

const (
	AuthorizationHeader = "Authorization"

	// Environment vars
	//      Database envVars
	ENDPOINT          = "ENDPOINT"
	SERVER_ADDRESS    = "SERVER_ADDRESS"
	REFRESH_TOKEN_EXP = "REFRESH_TOKEN_EXP"
	TEMP_TOKEN_EXP    = "TEMP_TOKEN_EXP"
	SECRET            = "SECRET"

	//      Token types envVars
	REF_TOKEN_TYPE  = "RefToken"
	TEMP_TOKEN_TYPE = "TempToken"

	//      TLS cert envVars
	TLS_CERT_PATH = "TLS_CERT_PATH"
	TLS_KEY_PATH  = "TLS_KEY_PATH"

	// Roles
	USER_ROLE  = "User"
	ADMIN_ROLE = "Admin"
)

func init() {
	logrus.SetLevel(logrus.DebugLevel)
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
	updateCh = make(chan actors.User, 100)
}

func Start() {
	defer func() {
		Db.Close()
		mut.Lock()
		close(updateCh)
		mut.Unlock()
	}()

	schema, _ := graphql.NewSchema(graphql.SchemaConfig{
		Query:        QueryType,
		Mutation:     MutationType,
		Subscription: SubscriptionType,
	})

	subscriptionManager := graphqlws.NewSubscriptionManager(&schema)

	wsHandler := graphqlws.NewHandler(graphqlws.HandlerConfig{
		SubscriptionManager: subscriptionManager,
		Authenticate: func(tokenHeader string) (i interface{}, err error) {
			log.Print(tokenHeader)
			token, err := getTokenFromHeader(tokenHeader)
			if err != nil {
				return nil, errAuthHeaderInvalid
			}
			claimsMap, err := verifyToken(token)
			if err != nil {
				return nil, err
			}

			_, err = getTempTokenDataFromUserClaims(claimsMap)
			return nil, err
		},
	})

	h := handler.New(&handler.Config{
		Schema:     &schema,
		Pretty:     true,
		GraphiQL:   false,
		Playground: true,
	})

	r := mux.NewRouter()
	auth := r.PathPrefix("/admin").Subrouter()
	auth.Use(headerAuthorization, disableCORS)
	auth.HandleFunc("/login", endpointLoginAdmin).Methods("POST")
	auth.HandleFunc("/registerUser", endpointRegisterUser).Methods("POST")
	auth.HandleFunc("/getRefreshToken", endpointGetRefreshTokenFromClient).Methods("POST")
	auth.HandleFunc("/changePassword", endpointChangeAdminPassword).Methods("POST")
	auth.HandleFunc("/deleteUser", endpointDeleteUser).Methods("POST")
	r.Handle(os.Getenv(ENDPOINT), headerAuthorization(disableCORS(h)))
	r.Handle("/subscriptions", disableCORS(wsHandler))

	server := http.Server{
		Addr:           os.Getenv(SERVER_ADDRESS),
		ReadTimeout:    15 * time.Second,
		WriteTimeout:   15 * time.Second,
		MaxHeaderBytes: 1 << 20,
		Handler:        r,
	}

	ctx, cancel := context.WithCancel(context.Background())
	go handleSubscriptions(subscriptionManager, &schema)
	go shutdown(&server, cancel)

	if config.keyPath != "" && config.certPath != "" {
		log.Print("Server Started in Https")
		if err := server.ListenAndServeTLS(config.certPath, config.keyPath); err != nil {
			log.Print(err)
		}
	} else {
		log.Print("Server Started in Http")
		if err := server.ListenAndServe(); err != nil {
			log.Print(err)
		}
	}

	<-ctx.Done()
	log.Print("Server are closed")
}

// handle subscriptions awaits updates to user locations and updates subscriptions.
func handleSubscriptions(manager graphqlws.SubscriptionManager, schema *graphql.Schema) {
	defer func() {
		for connection := range manager.Subscriptions() {
			manager.RemoveSubscriptions(connection)
		}
	}()

	for user := range updateCh {
		subscriptions := manager.Subscriptions()
		for conn, _ := range subscriptions {
			for _, subscription := range subscriptions[conn] {
				ctx := context.WithValue(context.Background(), "user", user)
				params := graphql.Params{
					Schema:         *schema, // The GraphQL schema
					RequestString:  subscription.Query,
					VariableValues: subscription.Variables,
					OperationName:  subscription.OperationName,
					Context:        ctx,
				}

				result := graphql.Do(params)
				if result.Data == nil && result.Errors == nil {
					continue
				}

				// Send query results back to the subscriber at any point
				dataRenamed := graphqlws.DataMessagePayload{
					// Data can be anything (interface{})
					Data: result.Data,
					// Errors is optional ([]error)
					Errors: graphqlws.ErrorsFromGraphQLErrors(result.Errors),
				}
				subscription.SendData(&dataRenamed)
			}
		}
	}
}
