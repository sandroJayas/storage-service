package test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestGetBoxByID(t *testing.T) {
	timestamp := time.Now().Format("150405")
	email := "boxget+" + timestamp + "@test.com"
	password := "testpass123"
	var token string
	var selfBoxID string
	var sortBoxID string

	t.Run("setup - register and login", func(t *testing.T) {
		body, _ := json.Marshal(map[string]string{
			"email":    email,
			"password": password,
		})
		http.Post(userBaseURL+"/users/register", "application/json", bytes.NewReader(body))

		body, _ = json.Marshal(map[string]string{
			"email":    email,
			"password": password,
		})
		resp, _ := http.Post(userBaseURL+"/users/login", "application/json", bytes.NewReader(body))

		var res map[string]interface{}
		_ = json.NewDecoder(resp.Body).Decode(&res)
		token = res["token"].(string)
	})

	t.Run("setup - create self and sort boxes", func(t *testing.T) {
		// self
		boxReq := map[string]string{
			"packing_mode": "self",
			"item_name":    "Laptop",
			"item_note":    "MacBook Pro",
		}
		body, _ := json.Marshal(boxReq)
		req, _ := http.NewRequest(http.MethodPost, boxBaseURL, bytes.NewReader(body))
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", "application/json")
		resp, _ := http.DefaultClient.Do(req)

		var res map[string]interface{}
		_ = json.NewDecoder(resp.Body).Decode(&res)
		selfBoxID = res["id"].(string)

		// sort
		boxReq = map[string]string{
			"packing_mode": "sort",
			"item_name":    "Stuff",
			"item_note":    "Will be sorted",
		}
		body, _ = json.Marshal(boxReq)
		req, _ = http.NewRequest(http.MethodPost, boxBaseURL, bytes.NewReader(body))
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", "application/json")
		resp, _ = http.DefaultClient.Do(req)

		_ = json.NewDecoder(resp.Body).Decode(&res)
		sortBoxID = res["id"].(string)
	})

	t.Run("get self-packed box - should return item", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, boxBaseURL+"/"+selfBoxID, nil)
		req.Header.Set("Authorization", "Bearer "+token)
		resp, err := http.DefaultClient.Do(req)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var res map[string]map[string]interface{}
		_ = json.NewDecoder(resp.Body).Decode(&res)
		box := res["box"]

		assert.Equal(t, "self", box["packing_mode"])
		assert.Equal(t, selfBoxID, box["id"])

		items := box["items"].([]interface{})
		assert.Len(t, items, 1)

		item := items[0].(map[string]interface{})
		assert.Equal(t, "Laptop", item["name"])
		assert.Equal(t, "MacBook Pro", item["description"])
	})

	t.Run("get sort-packed box - should return no items", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, boxBaseURL+"/"+sortBoxID, nil)
		req.Header.Set("Authorization", "Bearer "+token)
		resp, err := http.DefaultClient.Do(req)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var res map[string]map[string]interface{}
		_ = json.NewDecoder(resp.Body).Decode(&res)
		box := res["box"]

		assert.Equal(t, "sort", box["packing_mode"])
		assert.Equal(t, sortBoxID, box["id"])
		assert.Nil(t, box["items"])
	})

	t.Run("unauthorized access should fail", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, boxBaseURL+"/"+selfBoxID, nil)
		resp, err := http.DefaultClient.Do(req)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})

	t.Run("non-existent box ID should return 404", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, boxBaseURL+"/00000000-0000-0000-0000-000000000000", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		resp, err := http.DefaultClient.Do(req)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	})

	t.Run("malformed UUID should return 400", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, boxBaseURL+"/invalid-uuid", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		resp, err := http.DefaultClient.Do(req)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("another user cannot access box", func(t *testing.T) {
		// Register a second user
		email = "boxhacker+" + timestamp + "@test.com"
		password = "maliciouspass"
		var secondToken string

		// Register second user
		body, _ := json.Marshal(map[string]string{
			"email":    email,
			"password": password,
		})
		_, err := http.Post(userBaseURL+"/users/register", "application/json", bytes.NewReader(body))
		assert.Nil(t, err)

		// Login second user
		body, _ = json.Marshal(map[string]string{
			"email":    email,
			"password": password,
		})
		resp, _ := http.Post(userBaseURL+"/users/login", "application/json", bytes.NewReader(body))
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var res map[string]interface{}
		_ = json.NewDecoder(resp.Body).Decode(&res)
		secondToken = res["token"].(string)
		assert.NotEmpty(t, secondToken)

		// Attempt to fetch a box owned by the first user
		req, _ := http.NewRequest(http.MethodGet, boxBaseURL+"/"+selfBoxID, nil)
		req.Header.Set("Authorization", "Bearer "+secondToken)
		resp, err = http.DefaultClient.Do(req)

		assert.Nil(t, err)
		assert.Equal(t, http.StatusNotFound, resp.StatusCode, "Should not be able to access another user's box")
	})

}
