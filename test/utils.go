package test

import (
	"bytes"
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"
)

const userBaseURL = "http://localhost:8081"
const boxBaseURL = "http://localhost:8080/boxes"

func RegisterAndLogin(t *testing.T, email string, password string) string {
	var token string

	t.Run("setup - register and login", func(t *testing.T) {
		// Register
		body, _ := json.Marshal(map[string]string{
			"email":    email,
			"password": password,
		})
		resp, err := http.Post(userBaseURL+"/users/register", "application/json", bytes.NewReader(body))
		assert.NoError(t, err)
		assert.Equal(t, http.StatusCreated, resp.StatusCode)

		// Login
		body, _ = json.Marshal(map[string]string{
			"email":    email,
			"password": password,
		})
		resp, err = http.Post(userBaseURL+"/users/login", "application/json", bytes.NewReader(body))
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var res map[string]interface{}
		_ = json.NewDecoder(resp.Body).Decode(&res)
		token = res["token"].(string)
		assert.NotEmpty(t, token)
	})
	return token
}

func RegisterAndLoginEmployee(t *testing.T, email string, password string) string {
	var token string

	t.Run("setup - register and login", func(t *testing.T) {
		// Register
		body, _ := json.Marshal(map[string]string{
			"email":    email,
			"password": password,
		})
		resp, err := http.Post(userBaseURL+"/users/create-employee", "application/json", bytes.NewReader(body))
		assert.NoError(t, err)
		assert.Equal(t, http.StatusCreated, resp.StatusCode)

		// Login
		body, _ = json.Marshal(map[string]string{
			"email":    email,
			"password": password,
		})
		resp, err = http.Post(userBaseURL+"/users/login", "application/json", bytes.NewReader(body))
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var res map[string]interface{}
		_ = json.NewDecoder(resp.Body).Decode(&res)
		token = res["token"].(string)
		assert.NotEmpty(t, token)
	})
	return token
}

func CreateSortPackedBox(t *testing.T, token string) string {
	boxReq := map[string]string{
		"packing_mode": "sort",
		"item_name":    "Shoes",
		"item_note":    "To be sorted",
	}
	body, _ := json.Marshal(boxReq)
	req, _ := http.NewRequest(http.MethodPost, boxBaseURL, bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	var res map[string]interface{}
	_ = json.NewDecoder(resp.Body).Decode(&res)
	boxID := res["id"].(string)
	assert.NotEmpty(t, boxID)
	return boxID
}

func AddItemToBox(t *testing.T, token string, boxID string, item map[string]interface{}) map[string]string {
	body, _ := json.Marshal(item)
	req, _ := http.NewRequest(http.MethodPost, boxBaseURL+"/"+boxID+"/items", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	var res map[string]string
	_ = json.NewDecoder(resp.Body).Decode(&res)
	assert.NotEmpty(t, res["id"])
	return res
}
