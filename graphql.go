package main

import (
	"errors"
	"github.com/graphql-go/graphql"
)

var geoCordType = graphql.NewObject(
	graphql.ObjectConfig{
		Name: "geo_cord",
		Fields: graphql.Fields{
			"latitude": &graphql.Field{
				Type: graphql.Float,
			},
			"longitude": &graphql.Field{
				Type: graphql.Float,
			},
		},
		Description: "Geographical point",
	})

var userType = graphql.NewObject(
	graphql.ObjectConfig{
		Name: "User",
		Fields: graphql.Fields{
			"id": &graphql.Field{
				Type:        graphql.ID,
				Description: "User unique id",
			},
			"geo_cord": &graphql.Field{
				Type:        geoCordType,
				Description: "Geographical position of the current user",
			},
			"name": &graphql.Field{
				Type:        graphql.String,
				Description: "User name",
			},
			"category": &graphql.Field{
				Type:        graphql.String,
				Description: "Category of User Ex(CLIENT,SERVICE_PROVIDER)",
			},
		},
	})

var queryType = graphql.NewObject(graphql.ObjectConfig{
	Name: "Query",
	Fields: graphql.Fields{
		// Endpoint: /location?query={user(id: int ){name, geo_cord, category}}
		"user": &graphql.Field{
			Type:        userType,
			Description: "get User by id.",
			Args: graphql.FieldConfigArgument{
				"id": &graphql.ArgumentConfig{
					Type:        graphql.Int,
					Description: "User ID",
				},
			},
			Resolve: func(p graphql.ResolveParams) (i interface{}, err error) {
				if id, ok := p.Args["id"].(int); ok {
					return GetUser(id)
				}
				return nil, errors.New("id argument could be missing")
			},
		},
		// Endpoint: /location? query={allUsers(category = "GENERIC": String ){id, name, geo_cord, category}}
		"allUsers": &graphql.Field{
			Type: graphql.NewList(userType),
			Description: "Get all users by category.\n" +
				" Ex:(\"CLIENT\",\"SERVICE_PROVIDER\",\"GENERIC\").\n" +
				"default_value = \"GENERIC\".\n" +
				" With \"GENERIC\" returns all users.",
			Args: graphql.FieldConfigArgument{
				"category": &graphql.ArgumentConfig{
					Type:         graphql.String,
					DefaultValue: Generic,
					Description:  "Category of User Ex(CLIENT,SERVICE_PROVIDER)",
				},
			},
			Resolve: func(p graphql.ResolveParams) (i interface{}, err error) {
				if val, ok := p.Args["category"]; ok {
					return GetAllUsers(Category(val.(string)))
				}
				return nil, errors.New("category argument could be missing")
			},
		},
		// Endpoint: /location?query={getCloseUsers(originLat:float, originalLong:float, resolution:int, category:String){id, name,geo_cord, category }}
		"getCloseUsers": &graphql.Field{
			Type:        graphql.NewList(userType),
			Description: "Get list of user close to a position.",
			Args: graphql.FieldConfigArgument{
				"originLat": &graphql.ArgumentConfig{
					Type:        graphql.Float,
					Description: "Latitude of the origin point",
				},
				"originLong": &graphql.ArgumentConfig{
					Type:        graphql.Float,
					Description: "Longitude of the origin point",
				},
				"resolution": &graphql.ArgumentConfig{
					Type:         graphql.Int,
					DefaultValue: 0,
					Description: "Scale of the plane in H3 resolution(0-15.\n" +
						"default_value: 0\n" +
						"More Info: https://h3geo.org/#/documentation/core-library/resolution-table",
				},
				"category": &graphql.ArgumentConfig{
					Type:         graphql.String,
					DefaultValue: Generic,
					Description:  "Category of User Ex(CLIENT,SERVICE_PROVIDER)",
				},
			},
			Resolve: func(p graphql.ResolveParams) (i interface{}, err error) {
				var lat, long float64
				var resolution int
				var category string
				var ok bool

				if lat, ok = p.Args["originLat"].(float64); !ok {
					return nil, errors.New("lat argument could be missing")
				}

				if long, ok = p.Args["originLong"].(float64); ok {
					return nil, errors.New("long argument could be missing")
				}

				if category, ok = p.Args["category"].(string); !ok {
					return nil, errors.New("category argument could be missing")
				}

				if resolution, ok = p.Args["resolution"].(int); !ok {
					return nil, errors.New("resolution argument could be missing")
				}

				return GetCloseUsers(lat, long, resolution, Category(category))
			},
		},
	},
})

var mutationType = graphql.NewObject(graphql.ObjectConfig{
	Name: "Mutation",
	Fields: graphql.Fields{
		// Endpoint: /location?query=mutation+_{addUser(name:String, lat:float, long:float, category: String){id, name, geo_cord, category}}
		"addUser": &graphql.Field{
			Type: userType,
			Args: graphql.FieldConfigArgument{
				"name": &graphql.ArgumentConfig{
					Type:        graphql.String,
					Description: "User name",
				},
				"lat": &graphql.ArgumentConfig{
					Type:        graphql.Float,
					Description: "Current latitude",
				},
				"long": &graphql.ArgumentConfig{
					Type:        graphql.Float,
					Description: "Current longitude",
				},
				"category": &graphql.ArgumentConfig{
					Type: graphql.String,
					Description:  "Category of User Ex(CLIENT,SERVICE_PROVIDER)",
				},
			},
			Resolve: func(p graphql.ResolveParams) (i interface{}, err error) {
				var lat, long float64
				var name, category string
				var ok bool
				if lat, ok = p.Args["originLat"].(float64); !ok {
					return nil, errors.New("lat argument could be missing")
				}

				if long, ok = p.Args["originLong"].(float64); ok {
					return nil, errors.New("long argument could be missing")
				}

				if category, ok = p.Args["category"].(string); !ok {
					return nil, errors.New("category argument could be missing")
				}

				if name, ok = p.Args["name"].(string); !ok {
					return nil, errors.New("category argument could be missing")
				}

				return AddUser(NewUser(name, lat, long, Category(category)))
			},
		},
		// Endpoint: /location?query=mutation+_{updateGeoCord(lat:float, long:float){id, name, geo_cord, category}}
		"updateGeoCord": &graphql.Field{
			Type: userType,
			Description:"Update the coordinates of a User",
			Args: graphql.FieldConfigArgument{
				"id": &graphql.ArgumentConfig{
					Type: graphql.Int,
					Description:"User ID",
				},
				"lat": &graphql.ArgumentConfig{
					Type: graphql.Float,
					Description:"New latitude",
				},
				"long": &graphql.ArgumentConfig{
					Type: graphql.Float,
					Description:"New longitude",
				},
			},
			Resolve: func(p graphql.ResolveParams) (i interface{}, err error) {
				var lat, long float64
				var ok bool
				var id int

				if id, ok = p.Args["id"].(int); !ok {
					return nil, errors.New("id argument could be missing")
				}

				if lat, ok = p.Args["originLat"].(float64); !ok {
					return nil, errors.New("lat argument could be missing")
				}

				if long, ok = p.Args["originLong"].(float64); ok {
					return nil, errors.New("long argument could be missing")
				}

				return UpdateUser(id, lat, long)
			},
		},
		// Endpoint: /location?query=mutation+_{deleteUser(id : int){id, name, geo_cord, category}}
		"deleteUser": &graphql.Field{
			Type: userType,
			Description:"Remove an user by its Id",
			Args: graphql.FieldConfigArgument{
				"id": &graphql.ArgumentConfig{
					Type: graphql.Int,
					Description:"User ID",
				},
			},
			Resolve: func(p graphql.ResolveParams) (i interface{}, err error) {
				if id, ok := p.Args["id"].(int); ok {
					return DeleteUser(id)
				}

				return nil, errors.New("id argument could be missing")
			},
		},
	},
})

