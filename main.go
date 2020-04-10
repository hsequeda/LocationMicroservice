package main

import (
	"context"
	"fmt"
	"github.com/graphql-go/graphql"
	"github.com/graphql-go/handler"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"
)

var Db Storage

const (
	DB_USER        = "DB_USER"
	DB_PASS        = "DB_PASS"
	DB_NAME        = "DB_NAME"
	DB_HOST        = "DB_HOST"
	DB_SSLMODE     = "DB_SSL_MODE"
	ENDPOINT       = "ENDPOINT"
	SERVER_ADDRESS = "SERVER_ADDRESS"
)

func init() {
	var err error

	log.SetFlags(log.Lshortfile)
	Db, err = NewDb(os.Getenv(DB_USER),
		os.Getenv(DB_PASS),
		os.Getenv(DB_HOST),
		os.Getenv(DB_NAME),
		os.Getenv(DB_SSLMODE))
	if err != nil {
		fmt.Println(err)
	}
}

func main() {
	defer Db.Close()

	schema, _ := graphql.NewSchema(graphql.SchemaConfig{
		Query:    queryType,
		Mutation: mutationType,
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
