package test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/stretchr/testify/require"
	"net/http"
	"testing"
)

/*
 Is recommended run these test with the initial database without changes, is not
recommended using the database with which you will integrate your application.
 For run these test is necessary set different environment vars, is recommended using the
test.sh script. To see the list of the environment vars necessaries consult the test.sh script
 or the README.md file.
*/

var client = http.DefaultClient
var serverAddress = "http://localhost:8080"

func Test_LoginAdmin(t *testing.T) {

	var test_Cases = []struct {
		name          string
		adminUser     string
		adminPass     string
		withBasicAuth bool
		errCode       int
		errMsg        string
	}{
		{name: "OK", adminUser: "root", adminPass: "12345678", withBasicAuth: true, errCode: 200, errMsg: ""},

		{name: "withoutBasicAuth", adminUser: "root", withBasicAuth: false, adminPass: "12345678", errCode: 401,
			errMsg: "invalid credentials"},

		{name: "wrongAdmin", adminUser: "wrongUser", adminPass: "wrongPass", withBasicAuth: true, errCode: 500,
			errMsg: "invalid credentials"},

		{name: "wrongPassword", adminUser: "wrongUser", adminPass: "wrongPass", withBasicAuth: true, errCode: 500,
			errMsg: "invalid credentials"},
	}
	for _, tc := range test_Cases {
		t.Run(fmt.Sprintf("login Admin: %s", tc.name), func(t *testing.T) {
			resp, err := client.Do(generateLoginAdminReq(t, tc.adminUser, tc.adminPass, tc.withBasicAuth))
			require.NoError(t, err, tc.name)
			require.Equal(t, tc.errCode, resp.StatusCode, tc.name)

			var respJSON map[string]interface{}
			err = json.NewDecoder(resp.Body).Decode(&respJSON)
			require.NoError(t, err, tc.name)

			if resp.StatusCode == 200 {
				require.NotEmpty(t, respJSON["temp_token"])
			} else {
				require.Equal(t, tc.errCode, int(respJSON["http_status"].(float64)), tc.name)
				require.Equal(t, tc.errMsg, respJSON["error"], tc.name)
			}
		})
	}
}

func Test_ChangeAdminPassword(t *testing.T) {
	var test_Cases = []struct {
		name        string
		adminPass   string
		lastPass    string
		newPassword string
		token       string
		errCode     int
		errMsg      string
	}{
		{name: "invalidToken", adminPass: "12345678", token: "Bearer Wrong", lastPass: "12345678", newPassword: "11111111111",
			errCode: 511, errMsg: "invalid token"},

		{name: "invalidPass", adminPass: "12345678", lastPass: "wrongPass", newPassword: "111111111111", errCode: 400,
			errMsg: "password not match"},

		{name: "OK", adminPass: "12345678", lastPass: "12345678", newPassword: "87654321", errCode: 200, errMsg: ""},

		{name: "reset", adminPass: "87654321", lastPass: "87654321", newPassword: "12345678", errCode: 200, errMsg: ""},
	}

	for _, tc := range test_Cases {
		t.Run(fmt.Sprintf("ChangePassword: %s", tc.name), func(t *testing.T) {
			resp, err := client.Do(generateLoginAdminReq(t, "root", tc.adminPass, true))
			require.NoError(t, err, tc.name)

			var respJSON map[string]interface{}
			err = json.NewDecoder(resp.Body).Decode(&respJSON)
			require.NoError(t, err, tc.name)

			if tc.token == "" {
				var ok bool
				tc.token, ok = respJSON["temp_token"].(string)
				require.True(t, ok, tc.name)
			}

			resp, err = client.Do(generateChangeAdminPassReq(t, tc.token, tc.newPassword, tc.lastPass))
			require.NoError(t, err, tc.name)

			if resp.StatusCode != 200 {
				err = json.NewDecoder(resp.Body).Decode(&respJSON)
				require.NoError(t, err, tc.name)
				require.Equal(t, tc.errCode, int(respJSON["http_status"].(float64)), tc.name)
				require.Equal(t, tc.errMsg, respJSON["error"], tc.name)
			}
		})
	}
}

func Test_RegisterUser(t *testing.T) {
	var test_Cases = []struct {
		name         string
		userLat      float64
		userLong     float64
		userCategory string
		token        string
		errCode      int
		errMsg       string
	}{
		{name: "Ok", userLat: 1, userLong: 2, userCategory: "CLIENT", errCode: 200},
		{name: "WrongCategory", userLat: 1, userLong: 2, userCategory: "Wrong", errCode: 400, errMsg: "invalid user category"},
		{name: "WrongToken", userLat: 1, token: "Wrong", userLong: 2, userCategory: "SERVICE_PROVIDER", errCode: 511, errMsg: "invalid token"},
	}

	for _, tc := range test_Cases {
		t.Run(fmt.Sprintf("RegisterUser: %s", tc.name), func(t *testing.T) {
			resp, err := client.Do(generateLoginAdminReq(t, "root", "12345678", true))
			require.NoError(t, err, tc.name)

			var respJSON map[string]interface{}
			err = json.NewDecoder(resp.Body).Decode(&respJSON)
			require.NoError(t, err, tc.name)

			if tc.token == "" {
				var ok bool
				tc.token, ok = respJSON["temp_token"].(string)
				require.True(t, ok, tc.name)
			}

			resp, err = client.Do(generateRegisterUserReq(t, tc.token, tc.userLat, tc.userLong, tc.userCategory))
			require.NoError(t, err, tc.name)

			err = json.NewDecoder(resp.Body).Decode(&respJSON)
			require.NoError(t, err, tc.name)

			if resp.StatusCode == 200 {
				require.NotEmpty(t, respJSON["refresh_token"], tc.name)
				require.NotEmpty(t, respJSON["user_id"], tc.name)
				floatId, ok := respJSON["user_id"].(float64)
				require.True(t, ok, tc.name)
				require.NotEmpty(t, floatId, tc.name)
				resp, err = client.Do(generateDeleteUser(t, tc.token, int(floatId)))
				require.NoError(t, err, tc.name)
				require.Equal(t, 200, resp.StatusCode, tc.name)
			} else {
				require.Equal(t, tc.errCode, int(respJSON["http_status"].(float64)), tc.name)
				require.Equal(t, tc.errMsg, respJSON["error"], tc.name)
			}
		})
	}
}

func Test_GetRefreshTokenFromClient(t *testing.T) {
	resp, err := client.Do(generateLoginAdminReq(t, "root", "12345678", true))
	require.NoError(t, err, "GetRefreshTokenFromClient")

	var respJSON map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&respJSON)
	require.NoError(t, err, "GetRefreshTokenFromClient")

	token, ok := respJSON["temp_token"].(string)
	require.True(t, ok, "GetRefreshTokenFromClient")

	resp, err = client.Do(generateRegisterUserReq(t, token, 1, 2, "CLIENT"))
	require.NoError(t, err, "GetRefreshTokenFromClient")

	err = json.NewDecoder(resp.Body).Decode(&respJSON)
	require.NoError(t, err, "GetRefreshTokenFromClient")

	require.NotEmpty(t, respJSON["user_id"], "GetRefreshTokenFromClient")
	floatId, ok := respJSON["user_id"].(float64)
	require.True(t, ok, "GetRefreshTokenFromClient")

	var test_Cases = []struct {
		name    string
		token   string
		userId  int
		errCode int
		errMsg  string
	}{
		{name: "Ok", userId: int(floatId), errCode: 200},
		{name: "WrongToken", token: "Wrong", errCode: 511, errMsg: "invalid token"},
		{name: "WrongId", userId: int(floatId), errCode: 400, errMsg: "not found user with that id"},
	}

	for _, tc := range test_Cases {
		t.Run(fmt.Sprintf("GetRefreshTokenFromClient: %s", tc.name), func(t *testing.T) {
			resp, err := client.Do(generateLoginAdminReq(t, "root", "12345678", true))
			require.NoError(t, err, tc.name)

			var respJSON map[string]interface{}
			err = json.NewDecoder(resp.Body).Decode(&respJSON)
			require.NoError(t, err, tc.name)

			if tc.token == "" {
				tc.token, ok = respJSON["temp_token"].(string)
				require.True(t, ok, tc.name)
			}

			resp, err = client.Do(generateGetRefreshTokenReq(t, tc.token, tc.userId))
			require.NoError(t, err, tc.name)

			err = json.NewDecoder(resp.Body).Decode(&respJSON)
			require.NoError(t, err, tc.name)

			if resp.StatusCode == 200 {
				require.NotEmpty(t, respJSON["refresh_token"], tc.name)
				resp, err = client.Do(generateDeleteUser(t, tc.token, tc.userId))
				require.NoError(t, err, tc.name)
				require.Equal(t, 200, resp.StatusCode, tc.name)
			} else {
				require.Equal(t, tc.errCode, int(respJSON["http_status"].(float64)), tc.name)
				require.Equal(t, tc.errMsg, respJSON["error"], tc.name)
			}
		})
	}

}

func generateLoginAdminReq(t *testing.T, user, password string, withBasicAuth bool) *http.Request {
	req, err := http.NewRequest("POST", fmt.Sprintf("%s/admin/login", serverAddress), nil)
	require.NoError(t, err)
	if withBasicAuth {
		req.SetBasicAuth(user, password)
	}
	return req
}

func generateChangeAdminPassReq(t *testing.T, tempToken, newPassword, lastPassword string) *http.Request {
	b, err := json.Marshal(struct {
		LastPassword string `json:"last_password"`
		NewPassword  string `json:"new_password"`
	}{LastPassword: lastPassword, NewPassword: newPassword})
	require.NoError(t, err)

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/admin/changePassword", serverAddress), bytes.NewReader(b))
	require.NoError(t, err)

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", tempToken))
	return req
}

func generateRegisterUserReq(t *testing.T, tempToken string, userLat, userLong float64, userCategory string) *http.Request {
	b, err := json.Marshal(struct {
		Latitude  float64 `json:"lat"`
		Longitude float64 `json:"long"`
		Category  string  `json:"category"`
	}{Latitude: userLat, Longitude: userLong, Category: userCategory})
	require.NoError(t, err)

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/admin/registerUser", serverAddress), bytes.NewReader(b))
	require.NoError(t, err)

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", tempToken))
	req.Header.Set("Content-Type", "application/json")
	return req
}

func generateDeleteUser(t *testing.T, tempToken string, id int) *http.Request {
	b, err := json.Marshal(struct {
		Id int `json:"id"`
	}{Id: id})
	require.NoError(t, err)

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/admin/deleteUser", serverAddress), bytes.NewReader(b))
	require.NoError(t, err)

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", tempToken))
	return req
}

func generateGetRefreshTokenReq(t *testing.T, tempToken string, id int) *http.Request {
	b, err := json.Marshal(struct {
		Id int `json:"id"`
	}{Id: id})
	require.NoError(t, err)

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/admin/getRefreshToken", serverAddress), bytes.NewReader(b))
	require.NoError(t, err)

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", tempToken))
	return req
}
