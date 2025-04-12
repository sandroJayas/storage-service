package test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

const boxBaseURL = "http://localhost:8080/boxes"
const baseURL = "http://localhost:8081"

func TestCreateBox(t *testing.T) {
	timestamp := time.Now().Format("150405")

	// Create test user via auth service (or use fixed one)
	email := "box+" + timestamp + "@test.com"
	password := "testpass123"

	var token string

	t.Run("setup - register user", func(t *testing.T) {
		body, _ := json.Marshal(map[string]string{
			"email":    email,
			"password": password,
		})
		resp, err := http.Post(baseURL+"/users/register", "application/json", bytes.NewReader(body))
		assert.NoError(t, err)
		assert.Equal(t, http.StatusCreated, resp.StatusCode)
	})

	t.Run("setup - login to get token", func(t *testing.T) {
		body, _ := json.Marshal(map[string]string{
			"email":    email,
			"password": password,
		})
		resp, err := http.Post(baseURL+"/users/login", "application/json", bytes.NewReader(body))
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var res map[string]interface{}
		_ = json.NewDecoder(resp.Body).Decode(&res)
		token = res["token"].(string)
		assert.NotEmpty(t, token)
	})

	t.Run("create box - valid request", func(t *testing.T) {
		boxReq := map[string]string{
			"packing_mode": "self",
			"item_name":    "Shoes",
			"item_note":    "Running and dress shoes",
		}
		body, _ := json.Marshal(boxReq)
		req, _ := http.NewRequest(http.MethodPost, boxBaseURL, bytes.NewReader(body))
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", "application/json")

		client := http.Client{}
		resp, err := client.Do(req)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusCreated, resp.StatusCode)

		var res map[string]interface{}
		_ = json.NewDecoder(resp.Body).Decode(&res)
		assert.NotEmpty(t, res["id"])
		assert.IsType(t, "", res["id"])
	})

	t.Run("create box - missing packing mode", func(t *testing.T) {
		boxReq := map[string]string{
			"item_name": "Shoes",
			"item_note": "Missing packing mode",
		}
		body, _ := json.Marshal(boxReq)
		req, _ := http.NewRequest(http.MethodPost, boxBaseURL, bytes.NewReader(body))
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", "application/json")

		client := http.Client{}
		resp, err := client.Do(req)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})
}
