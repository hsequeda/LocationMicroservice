package main

import (
	"fmt"
	"github.com/graphql-go/graphql"
	"github.com/graphql-go/handler"
	"net/http"
)

var Db Storage

func init() {
	var err error
	Db, err = NewDb("postgres", "nightmare666", "location", SslModeDisable)
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

	http.Handle("/location", h)

	defer Db.Close()

	if err := http.ListenAndServe(":8080", nil); err != nil {
		fmt.Println(err)
	}
}
