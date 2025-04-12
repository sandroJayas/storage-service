package test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestListUserBoxes(t *testing.T) {
	timestamp := time.Now().Format("150405")
	email := "list+" + timestamp + "@test.com"
	password := "strongpass123"
	var token string

	t.Run("setup - register + login", func(t *testing.T) {
		// Register
		register := map[string]string{"email": email, "password": password}
		body, _ := json.Marshal(register)
		resp, err := http.Post(userBaseURL+"/users/register", "application/json", bytes.NewReader(body))
		assert.NoError(t, err)
		assert.Equal(t, http.StatusCreated, resp.StatusCode)

		// Login
		login := map[string]string{"email": email, "password": password}
		body, _ = json.Marshal(login)
		resp, err = http.Post(userBaseURL+"/users/login", "application/json", bytes.NewReader(body))
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var res map[string]interface{}
		_ = json.NewDecoder(resp.Body).Decode(&res)
		token = res["token"].(string)
		assert.NotEmpty(t, token)
	})

	t.Run("list - no boxes yet", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, boxBaseURL, nil)
		req.Header.Set("Authorization", "Bearer "+token)

		resp, err := http.DefaultClient.Do(req)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var res map[string][]interface{}
		_ = json.NewDecoder(resp.Body).Decode(&res)
		assert.Equal(t, 0, len(res["boxes"]))
	})

	t.Run("create multiple boxes", func(t *testing.T) {
		boxes := []map[string]string{
			{"packing_mode": "self", "item_name": "Shoes", "item_note": "Sneakers"},
			{"packing_mode": "sort", "item_name": "Books", "item_note": "Sci-fi"},
			{"packing_mode": "self", "item_name": "Hoodie", "item_note": "Black and warm"},
		}

		for _, box := range boxes {
			body, _ := json.Marshal(box)
			req, _ := http.NewRequest(http.MethodPost, boxBaseURL, bytes.NewReader(body))
			req.Header.Set("Authorization", "Bearer "+token)
			req.Header.Set("Content-Type", "application/json")
			resp, err := http.DefaultClient.Do(req)

			assert.NoError(t, err)
			assert.Equal(t, http.StatusCreated, resp.StatusCode)
		}
	})

	t.Run("list - should return all user boxes with correct fields", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, boxBaseURL, nil)
		req.Header.Set("Authorization", "Bearer "+token)

		resp, err := http.DefaultClient.Do(req)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var res map[string][]map[string]interface{}
		_ = json.NewDecoder(resp.Body).Decode(&res)

		boxes := res["boxes"]
		assert.Equal(t, 3, len(boxes))

		for _, box := range boxes {
			assert.NotEmpty(t, box["id"])
			assert.Contains(t, []interface{}{"self", "sort"}, box["packing_mode"])

			if box["packing_mode"] == "sort" {
				assert.Nil(t, box["items"])
			} else {
				items := box["items"].([]interface{})
				assert.Equal(t, len(items), 1)
				item := items[0].(map[string]interface{})
				assert.NotEmpty(t, item["name"])
				assert.NotEmpty(t, item["description"])
			}
		}
	})

	t.Run("unauthorized - no token", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, boxBaseURL, nil)
		resp, err := http.DefaultClient.Do(req)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})

	t.Run("unauthorized - malformed token", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, boxBaseURL, nil)
		req.Header.Set("Authorization", "Bearer this.is.not.valid")
		resp, err := http.DefaultClient.Do(req)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})

	t.Run("soft-deleted boxes should not appear", func(t *testing.T) {
		box := map[string]string{
			"packing_mode": "self",
			"item_name":    "To be deleted",
			"item_note":    "Should be soft-deleted",
		}
		body, _ := json.Marshal(box)
		req, _ := http.NewRequest(http.MethodPost, boxBaseURL, bytes.NewReader(body))
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", "application/json")

		resp, err := http.DefaultClient.Do(req)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusCreated, resp.StatusCode)

		var res map[string]string
		_ = json.NewDecoder(resp.Body).Decode(&res)
		boxID := res["id"]

		req, _ = http.NewRequest(http.MethodDelete, boxBaseURL+"/"+boxID, nil)
		req.Header.Set("Authorization", "Bearer "+token)
		resp, err = http.DefaultClient.Do(req)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		req, _ = http.NewRequest(http.MethodGet, boxBaseURL, nil)
		req.Header.Set("Authorization", "Bearer "+token)
		resp, err = http.DefaultClient.Do(req)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var listRes map[string][]map[string]interface{}
		_ = json.NewDecoder(resp.Body).Decode(&listRes)
		for _, b := range listRes["boxes"] {
			assert.NotEqual(t, boxID, b["id"])
		}
	})
}
