package main

import (
	"fmt"
	"github.com/graphql-go/graphql"
	"github.com/graphql-go/handler"
	"github.com/uber/h3-go"
	"net/http"
)


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

	if err := http.ListenAndServe(":8080", nil);
		err != nil {
		panic("Implement me!")
	}
}

func ExampleFromGeo() {
	geo := h3.GeoCoord{
		Latitude:  37.775938728915946,
		Longitude: -122.41795063018799,
	}

	geo2 := h3.GeoCoord{
		Latitude:  37.875938728915946,
		Longitude: -122.41795063018799,
	}
	resolution := 3

	h3P1 := h3.FromGeo(geo, resolution)
	h3P2 := h3.FromGeo(geo2, resolution)
	fmt.Printf("%#x\n", h3.FromGeo(geo, resolution))
	fmt.Printf("%#x\n", h3.FromGeo(geo2, resolution))
	fmt.Printf("%#v\n", h3.AreNeighbors(h3P1, h3P2))
	fmt.Printf("%#v\n", h3.ToParent(0x832830fffffffff, 3))
	fmt.Printf("%#v\n", h3.Resolution(0x832830fffffffff))
	fmt.Printf("%#v\n", h3.BaseCell(0x832830fffffffff))
	fmt.Printf("%#b\n", 0x832830fffffffff)

	fmt.Printf("%#v\n", h3.ToGeo(h3.FromGeo(geo, resolution)))
	rang, err := h3.HexRange(h3.FromGeo(geo, resolution), 1)
	if err != nil {
		fmt.Printf("%#v\n", err)
	}

	fmt.Printf("%#v\n", h3.KRingDistances(h3.FromGeo(geo, resolution), 1))

	fmt.Printf("%#v\n", rang)

	// Output:
	// 0x8928308280fffff
}

