package test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/sandroJayas/storage-service/test"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestDeleteBox(t *testing.T) {
	timestamp := time.Now().Format("150405")
	email := "delete+" + timestamp + "@test.com"
	password := "deletePass123!"
	var boxID string

	token := test.RegisterAndLogin(t, email, password)

	t.Run("setup - create box", func(t *testing.T) {
		boxReq := map[string]string{
			"packing_mode": "self",
			"item_name":    "Status Check Item",
			"item_note":    "Should be updated",
		}
		body, _ := json.Marshal(boxReq)
		req, _ := http.NewRequest(http.MethodPost, boxBaseURL, bytes.NewReader(body))
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", "application/json")

		client := http.Client{}
		resp, err := client.Do(req)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusCreated, resp.StatusCode)

		var createRes map[string]interface{}
		_ = json.NewDecoder(resp.Body).Decode(&createRes)
		boxID = createRes["id"].(string)
		assert.NotEmpty(t, boxID)
	})

	t.Run("setup - create box", func(t *testing.T) {
		boxReq := map[string]string{
			"packing_mode": "self",
			"item_name":    "Winter Coat",
			"item_note":    "Heavy parka",
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
		boxID = res["id"].(string)
		assert.NotEmpty(t, boxID)
	})

	t.Run("delete - valid request", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodDelete, fmt.Sprintf("%s/%s", boxBaseURL, boxID), nil)
		req.Header.Set("Authorization", "Bearer "+token)

		resp, err := http.DefaultClient.Do(req)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var res map[string]string
		_ = json.NewDecoder(resp.Body).Decode(&res)
		assert.Equal(t, "Box deleted", res["message"])
	})

	t.Run("delete - already deleted box", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodDelete, fmt.Sprintf("%s/%s", boxBaseURL, boxID), nil)
		req.Header.Set("Authorization", "Bearer "+token)

		resp, err := http.DefaultClient.Do(req)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	})

	t.Run("delete - unauthorized (no token)", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodDelete, fmt.Sprintf("%s/%s", boxBaseURL, boxID), nil)

		resp, err := http.DefaultClient.Do(req)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})

	t.Run("delete - foreign user", func(t *testing.T) {
		// Register and login as a second user (User B)
		foreignEmail := "userB+" + time.Now().Format("150405") + "@test.com"
		foreignPassword := "strongpass123"
		var foreignToken string

		// Register User B
		registerReq := map[string]string{
			"email":    foreignEmail,
			"password": foreignPassword,
		}
		body, _ := json.Marshal(registerReq)
		resp, err := http.Post(userBaseURL+"/users/register", "application/json", bytes.NewReader(body))
		assert.NoError(t, err)
		assert.Equal(t, http.StatusCreated, resp.StatusCode)

		// Login User B
		loginReq := map[string]string{
			"email":    foreignEmail,
			"password": foreignPassword,
		}
		body, _ = json.Marshal(loginReq)
		resp, err = http.Post(userBaseURL+"/users/login", "application/json", bytes.NewReader(body))
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var loginRes map[string]interface{}
		_ = json.NewDecoder(resp.Body).Decode(&loginRes)
		foreignToken = loginRes["token"].(string)
		assert.NotEmpty(t, foreignToken)

		req, _ := http.NewRequest(http.MethodDelete, fmt.Sprintf("%s/%s", boxBaseURL, boxID), nil)
		req.Header.Set("Authorization", "Bearer "+foreignToken)

		resp, err = http.DefaultClient.Do(req)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	})

	t.Run("delete - nonexistent box", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodDelete, fmt.Sprintf("%s/%s", boxBaseURL, "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"), nil)
		req.Header.Set("Authorization", "Bearer "+token)

		resp, err := http.DefaultClient.Do(req)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	})
}
