package server

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"locationMicroService/libs/actors"
	"locationMicroService/libs/util"
	"net/http"
	"strconv"
	"time"
)

func endpointRegisterUser(w http.ResponseWriter, r *http.Request) {
	headerAuth, ok := r.Context().Value("token").(string)
	if !ok {
		http.Error(w, errInvalidToken.Error(), http.StatusNetworkAuthenticationRequired)
		return
	}

	adminToken, err := getTokenFromHeader(headerAuth)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNetworkAuthenticationRequired)
		return
	}

	claimsMap, err := verifyToken(adminToken)
	if err != nil {
		http.Error(w, errInvalidToken.Error(), http.StatusNetworkAuthenticationRequired)
		return
	}

	adminId, err := getAdminDataFromClaims(claimsMap)
	if err != nil {
		http.Error(w, errInvalidToken.Error(), http.StatusNetworkAuthenticationRequired)
		return
	}

	var userParams struct {
		Latitude  float64 `json:"lat"`
		Longitude float64 `json:"long"`
		Category  string  `json:"category"`
	}
	if err = json.NewDecoder(r.Body).Decode(&userParams); err != nil {
		http.Error(w, "couldn't get request body", http.StatusBadRequest)
		return
	}

	refreshToken := "mockToken"
	newUser := actors.NewUser(refreshToken, userParams.Latitude, userParams.Longitude, userParams.Category, adminId)
	userId, err := Db.AddUser(newUser)
	if err != nil {
		http.Error(w, "error writing in the database", http.StatusInternalServerError)
		return
	}

	refreshToken, err = updateRefreshToken(userId)
	if err != nil {
		http.Error(w, fmt.Sprintf("error generating refreshToken: %s", err.Error()), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err = json.NewEncoder(w).Encode(refreshToken); err != nil {
		http.Error(w, "error encoding refreshToken to json", http.StatusInternalServerError)
		return
	}
}

func endpointLoginAdmin(w http.ResponseWriter, r *http.Request) {
	userName, password, ok := r.BasicAuth()
	if !ok {
		http.Error(w, "not authorize!", http.StatusUnauthorized)
		return
	}
	admin, err := Db.GetAdmin(userName)
	if err != nil {
		http.Error(w, "couldn't get the admin data from the database", http.StatusInternalServerError)
		return
	}
	if err := util.VerifyPassword(password, admin.PassHash); err != nil {
		http.Error(w, "invalid credentials", http.StatusUnauthorized)
		return
	}

	token, err := createToken(jwt.MapClaims{
		"id":   strconv.Itoa(admin.Id),
		"role": ADMIN_ROLE,
		"iat":  time.Now().Unix(),
		"exp":  time.Now().Add(config.tempTokenExp).Unix(),
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err = json.NewEncoder(w).Encode(token); err != nil {
		http.Error(w, "error returning token", http.StatusInternalServerError)
		return
	}
}

func endpointGetRefreshTokenFromClient(w http.ResponseWriter, r *http.Request) {
	tokenStr, err := getTokenFromHeader(r.Context().Value("token").(string))
	if err != nil {
		http.Error(w, errInvalidToken.Error(), http.StatusNetworkAuthenticationRequired)
		return
	}
	claimsMap, err := verifyToken(tokenStr)
	if err != nil {
		http.Error(w, errInvalidToken.Error(), http.StatusNetworkAuthenticationRequired)
		return
	}

	_, err = getAdminDataFromClaims(claimsMap)
	if err != nil {
		http.Error(w, errInvalidToken.Error(), http.StatusNetworkAuthenticationRequired)
		return
	}

	var userId struct {
		Id int `json:"id"`
	}
	if err = json.NewDecoder(r.Body).Decode(&userId); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	user, err := Db.GetUser(userId.Id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	userClaimsMap, err := verifyToken(user.RefreshToken)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if int64(userClaimsMap["exp"].(float64)) >= time.Now().Unix() {
		user.RefreshToken, err = updateRefreshToken(userId.Id)
		if err != nil {
			http.Error(w, "error generating token", http.StatusInternalServerError)
			return
		}
	}

	w.Header().Set("Content-Type", "application/json")
	if err = json.NewEncoder(w).Encode(user.RefreshToken); err != nil {
		http.Error(w, "error returning token", http.StatusInternalServerError)
		return
	}
}

func endpointChangeAdminPassword(w http.ResponseWriter, r *http.Request) {
	tokenStr, err := getTokenFromHeader(r.Context().Value("token").(string))
	if err != nil {
		http.Error(w, errInvalidToken.Error(), http.StatusNetworkAuthenticationRequired)
		return
	}
	claimsMap, err := verifyToken(tokenStr)
	if err != nil {
		http.Error(w, errInvalidToken.Error(), http.StatusNetworkAuthenticationRequired)
		return
	}

	adminId, err := getAdminDataFromClaims(claimsMap)
	if err != nil {
		http.Error(w, errInvalidToken.Error(), http.StatusNetworkAuthenticationRequired)
		return
	}

	var changePasswordParams struct {
		LastPassword string `json:"last_password"`
		NewPassword  string `json:"new_password"`
	}

	if err = json.NewDecoder(r.Body).Decode(&changePasswordParams); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	admin, err := Db.GetAdminById(adminId)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "not found admin with that id", http.StatusBadRequest)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	if err := util.VerifyPassword(changePasswordParams.LastPassword, admin.PassHash); err != nil {
		http.Error(w, "password not match", http.StatusBadRequest)
		return
	}

	newPasswordHash, err := util.GeneratePasswordHash(changePasswordParams.NewPassword)
	if err != nil {
		http.Error(w, "error generating new Password", http.StatusInternalServerError)
		return
	}
	if err := Db.UpdateAdminPassHash(adminId, newPasswordHash); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func endpointDeleteUser(w http.ResponseWriter, r *http.Request) {
	tokenStr, err := getTokenFromHeader(r.Context().Value("token").(string))
	if err != nil {
		http.Error(w, errInvalidToken.Error(), http.StatusNetworkAuthenticationRequired)
		return
	}
	claimsMap, err := verifyToken(tokenStr)
	if err != nil {
		http.Error(w, errInvalidToken.Error(), http.StatusNetworkAuthenticationRequired)
		return
	}

	_, err = getAdminDataFromClaims(claimsMap)
	if err != nil {
		http.Error(w, errInvalidToken.Error(), http.StatusNetworkAuthenticationRequired)
		return
	}

	var userId struct {
		Id int `json:"id"`
	}
	if err = json.NewDecoder(r.Body).Decode(&userId); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if _, err = Db.DeleteUser(userId.Id); err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "not found user with that id", http.StatusBadRequest)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}
}
