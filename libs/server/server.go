package server

import (
	"context"
	"fmt"
	"github.com/graphql-go/graphql"
	"github.com/graphql-go/handler"
	"locationMicroService/libs/core"
	"locationMicroService/libs/data"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"
)

var Db core.Storage

const (
	ENDPOINT       = "ENDPOINT"
	SERVER_ADDRESS = "SERVER_ADDRESS"
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
		Query:    QueryType,
		Mutation: MutationType,
	})

	h := handler.New(&handler.Config{
		Schema:   &schema,
		Pretty:   true,
		GraphiQL: true,
	})
	http.Handle(os.Getenv(ENDPOINT), h)

	server := http.Server{
		Addr:           os.Getenv(SERVER_ADDRESS),
		ReadTimeout:    1 * time.Second,
		WriteTimeout:   1 * time.Second,
		MaxHeaderBytes: 1,
	}

	ctx, cancel := context.WithCancel(context.Background())

	go shutdown(&server, cancel)

	log.Print("Starting Server")
	if err := server.ListenAndServe(); err != nil {
		log.Print(err)
	}

	<-ctx.Done()
	log.Print("Server are closed")
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
