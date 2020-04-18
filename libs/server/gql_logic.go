package server

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"github.com/graphql-go/graphql"
	"locationMicroService/libs/actors"
	"strconv"
	"time"
)

// ---- Queries ---->

// GetCloseUsers
func GetCloseUsers(params graphql.ResolveParams) (interface{}, error) {
	tokenStr, err := getTokenFromHeader(params.Context.Value("token").(string))
	if err != nil {
		return nil, err
	}
	claimsMap, err := verifyToken(tokenStr)
	if err != nil {
		return nil, err
	}

	_, err = getTempTokenDataFromUserClaims(claimsMap)
	if err != nil {
		return nil, err
	}

	var lat, long float64
	var resolution int
	var category string
	var ok bool

	if lat, ok = params.Args["origin_lat"].(float64); !ok {
		return nil, errors.New("lat argument could be missing")
	}

	if long, ok = params.Args["origin_long"].(float64); !ok {
		return nil, errors.New("long argument could be missing")
	}

	if category, ok = params.Args["category"].(string); !ok {
		return nil, errors.New("category argument could be missing")
	}

	if resolution, ok = params.Args["resolution"].(int); !ok {
		return nil, errors.New("resolution argument could be missing")
	}

	auxUser := actors.NewUser("", lat, long, actors.Generic, -1)
	if resolution < 0 || resolution > 15 {
		return nil, errors.New("resolution must be a value between 0 and 15. ")
	}
	return Db.GetCloseUsers(resolution, auxUser.H3Positions[resolution], category)
}

// GetAllUsers
func GetAllUsers(params graphql.ResolveParams) (interface{}, error) {
	tokenStr, err := getTokenFromHeader(params.Context.Value("token").(string))
	if err != nil {
		return nil, err
	}
	claimsMap, err := verifyToken(tokenStr)
	if err != nil {
		return nil, err
	}

	_, err = getTempTokenDataFromUserClaims(claimsMap)
	if err != nil {
		return nil, err
	}

	if category, ok := params.Args["category"].(string); ok {
		return Db.ListUsers(category)
	}
	return nil, errors.New("category argument could be missing")
}

func GetUser(params graphql.ResolveParams) (interface{}, error) {
	tokenStr, err := getTokenFromHeader(params.Context.Value("token").(string))
	if err != nil {
		return nil, err
	}
	claimsMap, err := verifyToken(tokenStr)
	if err != nil {
		return nil, err
	}

	_, err = getTempTokenDataFromUserClaims(claimsMap)
	if err != nil {
		return nil, err
	}

	if id, ok := params.Args["id"].(int); !ok {
		return nil, errors.New("id argument could be missing")
	} else {

		if user, err := Db.GetUser(id); err == sql.ErrNoRows {
			return nil, errors.New("not found User with inserted id")
		} else {

			return user, nil
		}
	}
}

func GetCurrentUser(params graphql.ResolveParams) (interface{}, error) {
	tokenStr, err := getTokenFromHeader(params.Context.Value("token").(string))
	if err != nil {
		return nil, err
	}
	claimsMap, err := verifyToken(tokenStr)
	if err != nil {
		return nil, err
	}

	id, err := getTempTokenDataFromUserClaims(claimsMap)
	if err != nil {
		return nil, err
	}

	if user, err := Db.GetUser(id); err == sql.ErrNoRows {
		return nil, errors.New("not found User with inserted id")
	} else {

		return user, nil
	}
}

// ---- Mutations ---->

// UpdateUser
func UpdateUser(params graphql.ResolveParams) (interface{}, error) {
	tokenStr, err := getTokenFromHeader(params.Context.Value("token").(string))
	if err != nil {
		return nil, err
	}
	claimsMap, err := verifyToken(tokenStr)
	if err != nil {
		return nil, err
	}

	id, err := getTempTokenDataFromUserClaims(claimsMap)
	if err != nil {
		return nil, err
	}

	var ok bool
	var lat, long float64
	if lat, ok = params.Args["new_lat"].(float64); !ok {
		return nil, errors.New("lat argument could be missing")
	}
	if long, ok = params.Args["new_long"].(float64); !ok {
		return nil, errors.New("long argument could be missing")
	}

	auxUser := actors.NewUser("", lat, long, actors.Generic, -1)

	user, err := Db.UpdateUser(id, lat, long, auxUser.H3Positions)
	switch {
	case err == sql.ErrNoRows:
		return nil, errors.New("not found User with inserted id")
	case err != nil:
		return nil, err
	default:
		go func() {
			mut.Lock()
			updateCh <- *user
			mut.Unlock()
		}()
		return user, nil
	}
}

// GetUserTempToken
func GetUserTempToken(params graphql.ResolveParams) (interface{}, error) {
	refreshToken, ok := params.Args["refresh_token"].(string)
	if !ok {
		return nil, errors.New("refreshToken argument could be missing")
	}

	tokenClaims, err := verifyToken(refreshToken)
	if err != nil {
		return nil, err
	}

	id, err := getRefreshTokenDataFromUserClaims(tokenClaims)
	if err != nil {
		return nil, errInvalidRefreshToken
	}

	user, err := Db.GetUser(id)
	if err != nil {
		return nil, errInvalidRefreshToken
	}

	if user.RefreshToken != refreshToken {
		return nil, errInvalidRefreshToken
	}

	return createToken(jwt.MapClaims{
		"id":       strconv.Itoa(id),
		"category": user.Category,
		"type":     TEMP_TOKEN_TYPE,
		"role":     USER_ROLE,
		"iat":      time.Now().Unix(),
		"exp":      time.Now().Add(config.tempTokenExp).Unix(),
	})
}

// ---- Subscriptions ---->

func GetUserPos(params graphql.ResolveParams) (interface{}, error) {
	user, ok := params.Context.Value("user").(actors.User)
	if ok {
		if id, ok := params.Args["id"].(int); !ok {
			return nil, errors.New("id argument could be missing")
		} else {
			if _, err := Db.GetUser(id); err != nil {
				return nil, fmt.Errorf("not found user with id: %d", id)
			}

			if id == user.Id {
				return user.GeoCord, nil
			}
		}
		return nil, nil
	}
	return nil, errors.New("invalid request")
}
