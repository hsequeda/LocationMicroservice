package server

import (
	"errors"
	"github.com/dgrijalva/jwt-go"
	"log"
	"strings"
)

var secret = "Maria"         // Implement me!
var refreshTokenExp = "360h" // Implement me!
var tempTokenExp = "15m"     // Implement me!

var errInvalidToken = errors.New(" invalid token!")
var errInvalidRefreshToken = errors.New(" invalid refresh token!")
var errAuthHeaderInvalid = errors.New(" header Authorization invalid!")

func createToken(claims jwt.MapClaims) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

func verifyToken(tokenStr string) (jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (i interface{}, err error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return false, errInvalidToken
		}
		return []byte(secret), nil
	})
	if err != nil {
		log.Printf("invalid token error: %s", err)
		return nil, errInvalidToken
	}

	if token.Valid == true {
		return token.Claims.(jwt.MapClaims), nil
	} else {
		log.Print(errInvalidToken)
		return nil, errInvalidToken
	}
}

func getTokenFromHeader(rawToken string) (string, error) {
	if rawToken == "" {
		return "", errAuthHeaderInvalid
	}
	rawTokenSplit := strings.Split(rawToken, " ")
	if len(rawTokenSplit) != 2 {
		return "", errAuthHeaderInvalid
	}
	if rawTokenSplit[0] != "Bearer" {
		return "", errAuthHeaderInvalid
	}

	return rawTokenSplit[1], nil
}
