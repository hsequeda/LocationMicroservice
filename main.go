package main

import (
	"fmt"
	"github.com/graphql-go/graphql"
	"github.com/graphql-go/handler"
	"log"
	"net/http"
	"os"
)

var Db Storage

const (
	DB_USER        = "rimaydb"
	DB_PASS        = "Wipaydb8##"
	DB_NAME        = "geodbv1"
	DB_HOST        = "localhost"
	DB_SSLMODE     = "require"
	ENDPOINT       = "location"
	SERVER_ADDRESS = "ranti.store:5432"
)

func init() {
	var err error
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
	schema, _ := graphql.NewSchema(graphql.SchemaConfig{
		Query:    queryType,
		Mutation: mutationType,
	})

	h := handler.New(&handler.Config{
		Schema: &schema,
		Pretty: true,
	})

	http.Handle(os.Getenv(ENDPOINT), h)

	defer Db.Close()

	log.Print("Starting Server")
	if err := http.ListenAndServe(os.Getenv(SERVER_ADDRESS), nil); err != nil {
		log.Fatal(err)
	}
}
