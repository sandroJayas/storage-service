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
const userBaseURL = "http://localhost:8081"

func TestCreateBox(t *testing.T) {
	timestamp := time.Now().Format("150405")

	email := "box+" + timestamp + "@test.com"
	password := "testpass123"
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

	t.Run("create box - self - full validation", func(t *testing.T) {
		boxReq := map[string]string{
			"packing_mode": "self",
			"item_name":    "Jacket",
			"item_note":    "Leather winter jacket",
		}
		body, _ := json.Marshal(boxReq)

		// Create box
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

		// Validate via GET /boxes/:id
		req, _ = http.NewRequest(http.MethodGet, boxBaseURL+"/"+boxID, nil)
		req.Header.Set("Authorization", "Bearer "+token)

		resp, err = http.DefaultClient.Do(req)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var getRes map[string]map[string]interface{}
		_ = json.NewDecoder(resp.Body).Decode(&getRes)

		box := getRes["box"]
		assert.Equal(t, "self", box["packing_mode"])

		items := box["items"].([]interface{})
		assert.Equal(t, 1, len(items))

		item := items[0].(map[string]interface{})
		assert.Equal(t, "Jacket", item["name"])
		assert.Equal(t, "Leather winter jacket", item["description"])
	})

	t.Run("create box - sort - should not include items", func(t *testing.T) {
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

		req, _ = http.NewRequest(http.MethodGet, boxBaseURL+"/"+boxID, nil)
		req.Header.Set("Authorization", "Bearer "+token)

		resp, err = http.DefaultClient.Do(req)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var getRes map[string]map[string]interface{}
		_ = json.NewDecoder(resp.Body).Decode(&getRes)

		box := getRes["box"]
		assert.Equal(t, "sort", box["packing_mode"])
		assert.Nil(t, box["items"], "sort-packed box should not include items")
	})

	t.Run("create box - duplicate request returns unique ID", func(t *testing.T) {
		boxReq := map[string]string{
			"packing_mode": "sort",
			"item_name":    "Shoes",
			"item_note":    "Repeat submission",
		}
		body, _ := json.Marshal(boxReq)

		var createdIDs []string

		for i := 0; i < 2; i++ {
			req, _ := http.NewRequest(http.MethodPost, boxBaseURL, bytes.NewReader(body))
			req.Header.Set("Authorization", "Bearer "+token)
			req.Header.Set("Content-Type", "application/json")

			resp, err := http.DefaultClient.Do(req)
			assert.NoError(t, err)
			assert.Equal(t, http.StatusCreated, resp.StatusCode)

			var res map[string]interface{}
			_ = json.NewDecoder(resp.Body).Decode(&res)
			id := res["id"].(string)
			assert.NotEmpty(t, id)
			createdIDs = append(createdIDs, id)
		}

		// Fetch all boxes and verify both IDs exist
		req, _ := http.NewRequest(http.MethodGet, boxBaseURL, nil)
		req.Header.Set("Authorization", "Bearer "+token)

		resp, err := http.DefaultClient.Do(req)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var listRes map[string][]map[string]interface{}
		_ = json.NewDecoder(resp.Body).Decode(&listRes)

		boxes := listRes["boxes"]
		assert.GreaterOrEqual(t, len(boxes), len(createdIDs))

		idSet := map[string]bool{}
		for _, box := range boxes {
			id := box["id"].(string)
			idSet[id] = true
		}

		for _, expectedID := range createdIDs {
			assert.True(t, idSet[expectedID], "box with ID %s not found in list", expectedID)
		}
	})

	t.Run("create box - invalid packing_mode", func(t *testing.T) {
		boxReq := map[string]string{
			"packing_mode": "invalid_mode",
			"item_name":    "Invalid",
			"item_note":    "Bad mode",
		}
		body, _ := json.Marshal(boxReq)

		req, _ := http.NewRequest(http.MethodPost, boxBaseURL, bytes.NewReader(body))
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", "application/json")

		resp, err := http.DefaultClient.Do(req)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("create box - missing packing_mode", func(t *testing.T) {
		boxReq := map[string]string{
			"item_name": "Shoes",
			"item_note": "Missing packing_mode",
		}
		body, _ := json.Marshal(boxReq)

		req, _ := http.NewRequest(http.MethodPost, boxBaseURL, bytes.NewReader(body))
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", "application/json")

		resp, err := http.DefaultClient.Do(req)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("create box - unauthorized request", func(t *testing.T) {
		boxReq := map[string]string{
			"packing_mode": "self",
			"item_name":    "Unauthorized",
			"item_note":    "Should fail",
		}
		body, _ := json.Marshal(boxReq)

		req, _ := http.NewRequest(http.MethodPost, boxBaseURL, bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err := http.DefaultClient.Do(req)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})

	t.Run("create box - malformed token", func(t *testing.T) {
		boxReq := map[string]string{
			"packing_mode": "self",
			"item_name":    "Invalid Token",
			"item_note":    "Bad JWT",
		}
		body, _ := json.Marshal(boxReq)

		req, _ := http.NewRequest(http.MethodPost, boxBaseURL, bytes.NewReader(body))
		req.Header.Set("Authorization", "Bearer invalid.token.value")
		req.Header.Set("Content-Type", "application/json")

		resp, err := http.DefaultClient.Do(req)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})
}
