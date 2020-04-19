package server

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"locationMicroService/libs/actors"
	"locationMicroService/libs/util"
	"net/http"
	"strconv"
	"time"
)

var errInvalidCredentials = errors.New("invalid credentials")

func endpointRegisterUser(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	headerAuth, ok := r.Context().Value("token").(string)
	if !ok {
		http.Error(w, string(getHttpErr(errInvalidToken.Error(), http.StatusNetworkAuthenticationRequired)), http.StatusNetworkAuthenticationRequired)
		return
	}

	adminToken, err := getTokenFromHeader(headerAuth)
	if err != nil {
		http.Error(w, string(getHttpErr(err.Error(), http.StatusNetworkAuthenticationRequired)), http.StatusNetworkAuthenticationRequired)
		return
	}

	claimsMap, err := verifyToken(adminToken)
	if err != nil {
		http.Error(w, string(getHttpErr(errInvalidToken.Error(), http.StatusNetworkAuthenticationRequired)), http.StatusNetworkAuthenticationRequired)
		return
	}

	adminId, err := getAdminDataFromClaims(claimsMap)
	if err != nil {
		http.Error(w, string(getHttpErr(errInvalidToken.Error(), http.StatusNetworkAuthenticationRequired)), http.StatusNetworkAuthenticationRequired)
		return
	}

	var userParams struct {
		Latitude  float64 `json:"lat"`
		Longitude float64 `json:"long"`
		Category  string  `json:"category"`
	}
	if err = json.NewDecoder(r.Body).Decode(&userParams); err != nil {
		http.Error(w, string(getHttpErr("couldn't get request body", http.StatusBadRequest)), http.StatusBadRequest)
		return
	}

	if userParams.Category != actors.Client && userParams.Category != actors.ServiceProvider {
		http.Error(w, string(getHttpErr("invalid user category", http.StatusBadRequest)), http.StatusBadRequest)
		return
	}

	refreshToken := "mockToken"
	newUser := actors.NewUser(refreshToken, userParams.Latitude, userParams.Longitude, userParams.Category, adminId)
	userId, err := Db.AddUser(newUser)
	if err != nil {
		http.Error(w, string(getHttpErr("error writing in the database", http.StatusInternalServerError)), http.StatusInternalServerError)
		return
	}

	refreshToken, err = updateRefreshToken(userId)
	if err != nil {
		http.Error(w, string(getHttpErr(fmt.Sprintf("error generating refreshToken: %s", err.Error()), http.StatusInternalServerError)), http.StatusInternalServerError)
		return
	}

	if err = json.NewEncoder(w).Encode(struct {
		UserId       int    `json:"user_id"`
		RefreshToken string `json:"refresh_token"`
	}{UserId: userId, RefreshToken: refreshToken}); err != nil {
		http.Error(w, string(getHttpErr("error encoding refreshToken to json", http.StatusInternalServerError)), http.StatusInternalServerError)
		return
	}
}

func endpointLoginAdmin(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	userName, password, ok := r.BasicAuth()
	if !ok {
		http.Error(w, string(getHttpErr(errInvalidCredentials.Error(), http.StatusUnauthorized)), http.StatusUnauthorized)
		return
	}
	admin, err := Db.GetAdmin(userName)
	switch {
	case err == sql.ErrNoRows:
		http.Error(w, string(getHttpErr("not found admin with that username", http.StatusBadRequest)), http.StatusBadRequest)
		return
	case err != nil:
		http.Error(w, string(getHttpErr(err.Error(), http.StatusInternalServerError)), http.StatusInternalServerError)
		return
	}

	if err := util.VerifyPassword(password, admin.PassHash); err != nil {
		http.Error(w, string(getHttpErr(errInvalidCredentials.Error(), http.StatusUnauthorized)), http.StatusUnauthorized)
		return
	}

	token, err := createToken(jwt.MapClaims{
		"id":   strconv.Itoa(admin.Id),
		"role": ADMIN_ROLE,
		"iat":  time.Now().Unix(),
		"exp":  time.Now().Add(config.tempTokenExp).Unix(),
	})
	if err != nil {
		http.Error(w, string(getHttpErr(err.Error(), http.StatusInternalServerError)), http.StatusInternalServerError)
		return
	}

	if err = json.NewEncoder(w).Encode(struct {
		TempToken string `json:"temp_token"`
	}{TempToken: token}); err != nil {
		http.Error(w, string(getHttpErr("error returning token", http.StatusInternalServerError)), http.StatusInternalServerError)
		return
	}
}

func endpointGetRefreshTokenFromClient(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	tokenStr, err := getTokenFromHeader(r.Context().Value("token").(string))
	if err != nil {
		http.Error(w, string(getHttpErr(err.Error(), http.StatusNetworkAuthenticationRequired)), http.StatusNetworkAuthenticationRequired)
		return
	}
	claimsMap, err := verifyToken(tokenStr)
	if err != nil {
		http.Error(w, string(getHttpErr(errInvalidToken.Error(), http.StatusNetworkAuthenticationRequired)), http.StatusNetworkAuthenticationRequired)
		return
	}

	_, err = getAdminDataFromClaims(claimsMap)
	if err != nil {
		http.Error(w, string(getHttpErr(errInvalidToken.Error(), http.StatusNetworkAuthenticationRequired)), http.StatusNetworkAuthenticationRequired)
		return
	}

	var userId struct {
		Id int `json:"id"`
	}
	if err = json.NewDecoder(r.Body).Decode(&userId); err != nil {
		http.Error(w, string(getHttpErr(err.Error(), http.StatusBadRequest)), http.StatusBadRequest)
		return
	}

	user, err := Db.GetUser(userId.Id)
	switch {
	case err == sql.ErrNoRows:
		http.Error(w, string(getHttpErr("not found user with that id", http.StatusBadRequest)), http.StatusBadRequest)
		return
	case err != nil:
		http.Error(w, string(getHttpErr(err.Error(), http.StatusInternalServerError)), http.StatusInternalServerError)
		return
	}

	_, err = verifyToken(user.RefreshToken)
	if err != nil {
		user.RefreshToken, err = updateRefreshToken(userId.Id)
		if err != nil {
			http.Error(w, string(getHttpErr("error generating token", http.StatusInternalServerError)), http.StatusInternalServerError)
			return
		}
	}

	if err = json.NewEncoder(w).Encode(struct {
		RefreshToken string `json:"refresh_token"`
	}{RefreshToken: user.RefreshToken}); err != nil {
		http.Error(w, string(getHttpErr("error returning token", http.StatusInternalServerError)), http.StatusInternalServerError)
		return
	}
}

func endpointChangeAdminPassword(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	tokenStr, err := getTokenFromHeader(r.Context().Value("token").(string))
	if err != nil {
		http.Error(w, string(getHttpErr(errInvalidToken.Error(), http.StatusNetworkAuthenticationRequired)), http.StatusNetworkAuthenticationRequired)
		return
	}
	claimsMap, err := verifyToken(tokenStr)
	if err != nil {
		http.Error(w, string(getHttpErr(errInvalidToken.Error(), http.StatusNetworkAuthenticationRequired)), http.StatusNetworkAuthenticationRequired)
		return
	}

	adminId, err := getAdminDataFromClaims(claimsMap)
	if err != nil {
		http.Error(w, string(getHttpErr(errInvalidToken.Error(), http.StatusNetworkAuthenticationRequired)), http.StatusNetworkAuthenticationRequired)
		return
	}

	var changePasswordParams struct {
		LastPassword string `json:"last_password"`
		NewPassword  string `json:"new_password"`
	}

	if err = json.NewDecoder(r.Body).Decode(&changePasswordParams); err != nil {
		http.Error(w, string(getHttpErr(err.Error(), http.StatusBadRequest)), http.StatusBadRequest)
		return
	}

	admin, err := Db.GetAdminById(adminId)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, string(getHttpErr("not found admin with that id", http.StatusBadRequest)), http.StatusBadRequest)
		} else {
			http.Error(w, string(getHttpErr(err.Error(), http.StatusInternalServerError)), http.StatusInternalServerError)
		}
		return
	}

	if err := util.VerifyPassword(changePasswordParams.LastPassword, admin.PassHash); err != nil {
		http.Error(w, string(getHttpErr("password not match", http.StatusBadRequest)), http.StatusBadRequest)
		return
	}

	newPasswordHash, err := util.GeneratePasswordHash(changePasswordParams.NewPassword)
	if err != nil {
		http.Error(w, string(getHttpErr("error generating new Password", http.StatusInternalServerError)), http.StatusInternalServerError)
		return
	}
	if err = Db.UpdateAdminPassHash(adminId, newPasswordHash); err != nil {
		http.Error(w, string(getHttpErr(err.Error(), http.StatusInternalServerError)), http.StatusInternalServerError)
		return
	}
}

func endpointDeleteUser(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	tokenStr, err := getTokenFromHeader(r.Context().Value("token").(string))
	if err != nil {
		http.Error(w, string(getHttpErr(errInvalidToken.Error(), http.StatusNetworkAuthenticationRequired)), http.StatusNetworkAuthenticationRequired)
		return
	}
	claimsMap, err := verifyToken(tokenStr)
	if err != nil {
		http.Error(w, string(getHttpErr(errInvalidToken.Error(), http.StatusNetworkAuthenticationRequired)), http.StatusNetworkAuthenticationRequired)
		return
	}

	_, err = getAdminDataFromClaims(claimsMap)
	if err != nil {
		http.Error(w, string(getHttpErr(errInvalidToken.Error(), http.StatusNetworkAuthenticationRequired)), http.StatusNetworkAuthenticationRequired)
		return
	}

	var userId struct {
		Id int `json:"id"`
	}
	if err = json.NewDecoder(r.Body).Decode(&userId); err != nil {
		http.Error(w, string(getHttpErr(err.Error(), http.StatusBadRequest)), http.StatusBadRequest)
		return
	}

	if _, err = Db.DeleteUser(userId.Id); err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, string(getHttpErr("not found user with that id", http.StatusBadRequest)), http.StatusBadRequest)
		} else {
			http.Error(w, string(getHttpErr(err.Error(), http.StatusInternalServerError)), http.StatusInternalServerError)
		}
		return
	}
}
