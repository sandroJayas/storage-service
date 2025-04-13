package test

import (
	"bytes"
	"encoding/json"
	"github.com/sandroJayas/storage-service/test"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

const boxBaseURL = "http://localhost:8080/boxes"
const userBaseURL = "http://localhost:8081"

func TestAddItem(t *testing.T) {
	timestamp := time.Now().Format("150405")
	email := "additem+" + timestamp + "@test.com"
	password := "additempass123"

	var sortBoxID string
	token := test.RegisterAndLogin(t, email, password)

	t.Run("setup - create self-packed box", func(t *testing.T) {
		reqBody := map[string]string{
			"packing_mode": "self",
			"item_name":    "dummy",
			"item_note":    "placeholder",
		}
		body, _ := json.Marshal(reqBody)
		req, _ := http.NewRequest(http.MethodPost, boxBaseURL, bytes.NewReader(body))
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", "application/json")

		resp, _ := http.DefaultClient.Do(req)
		assert.Equal(t, http.StatusCreated, resp.StatusCode)

		var res map[string]string
		_ = json.NewDecoder(resp.Body).Decode(&res)
		sortBoxID = res["id"]
		assert.NotEmpty(t, sortBoxID)
	})

	t.Run("add valid item to self-packed box", func(t *testing.T) {
		item := map[string]interface{}{
			"name":        "Shoes",
			"description": "Running shoes",
			"quantity":    1,
			"image_url":   "http://example.com/shoes.jpg",
		}
		body, _ := json.Marshal(item)
		req, _ := http.NewRequest(http.MethodPost, boxBaseURL+"/"+sortBoxID+"/items", bytes.NewReader(body))
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", "application/json")

		resp, err := http.DefaultClient.Do(req)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)

		var res map[string]string
		_ = json.NewDecoder(resp.Body).Decode(&res)
		assert.Empty(t, res["id"])
	})

	sortBoxID = test.CreateSortPackedBox(t, token)

	t.Run("add valid item to sort-packed box", func(t *testing.T) {
		item := map[string]interface{}{
			"name":        "Shoes",
			"description": "Running shoes",
			"quantity":    1,
			"image_url":   "http://example.com/shoes.jpg",
		}
		test.AddItemToBox(t, token, sortBoxID, item)
	})

	t.Run("fail to add item without auth", func(t *testing.T) {
		item := map[string]string{
			"name":        "Unauthorized",
			"description": "Should fail",
		}
		body, _ := json.Marshal(item)
		req, _ := http.NewRequest(http.MethodPost, boxBaseURL+"/"+sortBoxID+"/items", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err := http.DefaultClient.Do(req)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})

	t.Run("fail with invalid payload", func(t *testing.T) {
		body := []byte(`{"quantity": "wrong-type"}`)
		req, _ := http.NewRequest(http.MethodPost, boxBaseURL+"/"+sortBoxID+"/items", bytes.NewReader(body))
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", "application/json")

		resp, _ := http.DefaultClient.Do(req)
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("fail with malformed box ID", func(t *testing.T) {
		item := map[string]string{
			"name": "Invalid Box ID",
		}
		body, _ := json.Marshal(item)
		req, _ := http.NewRequest(http.MethodPost, boxBaseURL+"/not-a-uuid/items", bytes.NewReader(body))
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", "application/json")

		resp, _ := http.DefaultClient.Do(req)
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("unauthorized user cannot add to another user's box", func(t *testing.T) {
		email := "hacker+" + timestamp + "@test.com"
		password := "hackerpass"

		body, _ := json.Marshal(map[string]string{"email": email, "password": password})
		http.Post(userBaseURL+"/users/register", "application/json", bytes.NewReader(body))

		body, _ = json.Marshal(map[string]string{"email": email, "password": password})
		resp, _ := http.Post(userBaseURL+"/users/login", "application/json", bytes.NewReader(body))

		var res map[string]interface{}
		_ = json.NewDecoder(resp.Body).Decode(&res)
		hackerToken := res["token"].(string)
		item := map[string]interface{}{
			"name":        "Stolen Item",
			"description": "Running shoes",
			"quantity":    1,
			"image_url":   "http://example.com/shoes.jpg",
		}
		body, _ = json.Marshal(item)
		req, _ := http.NewRequest(http.MethodPost, boxBaseURL+"/"+sortBoxID+"/items", bytes.NewReader(body))
		req.Header.Set("Authorization", "Bearer "+hackerToken)
		req.Header.Set("Content-Type", "application/json")

		resp, _ = http.DefaultClient.Do(req)
		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	})
}
