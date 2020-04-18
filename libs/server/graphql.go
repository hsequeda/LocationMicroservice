package server

import (
	"github.com/graphql-go/graphql"
	"locationMicroService/libs/actors"
)

var geoCordType = graphql.NewObject(
	graphql.ObjectConfig{
		Name: "geo_cord",
		Fields: graphql.Fields{
			"latitude": &graphql.Field{
				Type: graphql.NewNonNull(graphql.Float),
			},
			"longitude": &graphql.Field{
				Type: graphql.NewNonNull(graphql.Float),
			},
		},
		Description: "Geographical point",
	})

var categoryEnum = graphql.NewEnum(graphql.EnumConfig{
	Name: "category",
	Values: graphql.EnumValueConfigMap{
		actors.Client: &graphql.EnumValueConfig{
			Value: actors.Client,
		},
		actors.ServiceProvider: &graphql.EnumValueConfig{
			Value: actors.ServiceProvider,
		},
		actors.Generic: &graphql.EnumValueConfig{
			Value: actors.Generic,
		},
	},
})

var userType = graphql.NewObject(graphql.ObjectConfig{
	Name: "user",

	Fields: graphql.Fields{
		"id": &graphql.Field{
			Type:        graphql.ID,
			Description: "User unique id",
		},
		"geo_cord": &graphql.Field{
			Type:        geoCordType,
			Description: "Geographical position of the current user",
		},
		"category": &graphql.Field{
			Type:        categoryEnum,
			Description: "Category of User Ex(CLIENT,SERVICE_PROVIDER)",
		},
	},
})

var QueryType = graphql.NewObject(graphql.ObjectConfig{
	Name: "Query",
	Fields: graphql.Fields{

		// Endpoint: /location?query={user(id: int ){id, geo_cord, category}}
		"user": &graphql.Field{
			Type:        userType,
			Description: "get User by id.",
			Args: graphql.FieldConfigArgument{
				"id": &graphql.ArgumentConfig{
					Type:        graphql.NewNonNull(graphql.Int),
					Description: "User ID",
				},
			},
			Resolve: GetUser,
		},

		// Endpoint: /location?query={currentUser(){id, geo_cord, category}}
		"currentUser": &graphql.Field{
			Type:        userType,
			Description: "get the current user.",
			Resolve:     GetCurrentUser,
		},

		// Endpoint: /location?query={allUsers(category: category){id, geo_cord, category}}
		"allUsers": &graphql.Field{
			Type: graphql.NewList(userType),
			Description: "Get all users by category.\n" +
				" Ex:(\"CLIENT\",\"SERVICE_PROVIDER\",\"GENERIC\").\n" +
				" With \"GENERIC\" returns all users.",
			Args: graphql.FieldConfigArgument{
				"category": &graphql.ArgumentConfig{
					Type:        categoryEnum,
					Description: "Category of User Ex(CLIENT, SERVICE_PROVIDER)",
				},
			},
			Resolve: GetAllUsers,
		},

		// Endpoint: /location?query={getCloseUsers(origin_lat: float, origin_long: float, resolution: int, category: category){id, geo_cord, category }}
		"getCloseUsers": &graphql.Field{
			Type:        graphql.NewList(userType),
			Description: "Get list of user close to a position.",
			Args: graphql.FieldConfigArgument{
				"origin_lat": &graphql.ArgumentConfig{
					Type:        graphql.NewNonNull(graphql.Float),
					Description: "Latitude of the origin point",
				},
				"origin_long": &graphql.ArgumentConfig{
					Type:        graphql.NewNonNull(graphql.Float),
					Description: "Longitude of the origin point",
				},
				"resolution": &graphql.ArgumentConfig{
					Type: graphql.NewNonNull(graphql.Int),
					Description: "Scale of the plane in H3 resolution(0-15.\n" +
						"More Info: https://h3geo.org/#/documentation/core-library/resolution-table",
				},
				"category": &graphql.ArgumentConfig{
					Type:        categoryEnum,
					Description: "Category of User Ex(CLIENT, SERVICE_PROVIDER)",
				},
			},
			Resolve: GetCloseUsers,
		},
	},
})

var MutationType = graphql.NewObject(graphql.ObjectConfig{
	Name: "Mutation",
	Fields: graphql.Fields{

		// Endpoint: /location?query=mutation+_{updateGeoCord(new_lat: float, new_long: float){id, geo_cord, category}}
		"updateGeoCord": &graphql.Field{
			Type:        userType,
			Description: "Update the coordinates of a User",
			Args: graphql.FieldConfigArgument{
				"new_lat": &graphql.ArgumentConfig{
					Type:        graphql.NewNonNull(graphql.Float),
					Description: "New latitude",
				},
				"new_long": &graphql.ArgumentConfig{
					Type:        graphql.NewNonNull(graphql.Float),
					Description: "New longitude",
				},
			},
			Resolve: UpdateUser,
		},

		// Endpoint: /location?query=mutation+_{getUserTempToken(refresh_token: string){id, geo_cord, category }}
		"getUserTempToken": &graphql.Field{
			Type: graphql.String,
			Description: "getUserTempToken  returns a temporary token with which to access the user's queries," +
				" mutations and subscriptions.",

			Args: graphql.FieldConfigArgument{
				"refresh_token": &graphql.ArgumentConfig{
					Type: graphql.NewNonNull(graphql.String),
					Description: "A refreshToken is a token with a longer duration that is temporaryToken," +
						" it works to obtain the temporaryToken, when the user's refreshToken expires, it has to" +
						" log in to the system again.",
				},
			},
			Resolve: GetUserTempToken,
		},
	},
})

var SubscriptionType = graphql.NewObject(graphql.ObjectConfig{
	Name: "Subscription",
	Fields: graphql.Fields{
		"getUserPos": &graphql.Field{
			Type: geoCordType,
			Args: graphql.FieldConfigArgument{
				"id": &graphql.ArgumentConfig{
					Type:        graphql.NewNonNull(graphql.Int),
					Description: "User ID",
				},
			},
			Resolve:     GetUserPos,
			Description: "subscribe to an User to get their position when it is updated",
		},
	},
})
