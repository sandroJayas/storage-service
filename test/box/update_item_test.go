package test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestUpdateItem(t *testing.T) {
	timestamp := time.Now().Format("150405")
	email := "updateitem+" + timestamp + "@test.com"
	password := "supersecurepass"

	var token string
	var boxID string
	var itemID string

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

	t.Run("setup - create self box", func(t *testing.T) {
		boxReq := map[string]string{
			"packing_mode": "self",
			"item_name":    "Watch",
			"item_note":    "Smartwatch",
		}
		body, _ := json.Marshal(boxReq)
		req, _ := http.NewRequest(http.MethodPost, boxBaseURL, bytes.NewReader(body))
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", "application/json")

		resp, _ := http.DefaultClient.Do(req)

		var res map[string]interface{}
		_ = json.NewDecoder(resp.Body).Decode(&res)
		boxID = res["id"].(string)
	})

	t.Run("setup - get item ID from box", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, boxBaseURL+"/"+boxID, nil)
		req.Header.Set("Authorization", "Bearer "+token)

		resp, _ := http.DefaultClient.Do(req)
		var res map[string]map[string]interface{}
		_ = json.NewDecoder(resp.Body).Decode(&res)

		item := res["box"]["items"].([]interface{})[0].(map[string]interface{})
		itemID = item["id"].(string)
		assert.NotEmpty(t, itemID)
	})

	t.Run("update name and description", func(t *testing.T) {
		updateReq := map[string]string{
			"name":        "Updated Watch",
			"description": "Updated Smartwatch",
		}
		body, _ := json.Marshal(updateReq)
		req, _ := http.NewRequest(http.MethodPatch, boxBaseURL+"/"+boxID+"/items/"+itemID, bytes.NewReader(body))
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", "application/json")

		resp, err := http.DefaultClient.Do(req)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("update quantity and image URL", func(t *testing.T) {
		updateReq := map[string]interface{}{
			"quantity":  2,
			"image_url": "http://example.com/item.jpg",
		}
		body, _ := json.Marshal(updateReq)
		req, _ := http.NewRequest(http.MethodPatch, boxBaseURL+"/"+boxID+"/items/"+itemID, bytes.NewReader(body))
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", "application/json")

		resp, err := http.DefaultClient.Do(req)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("verify updates", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, boxBaseURL+"/"+boxID, nil)
		req.Header.Set("Authorization", "Bearer "+token)

		resp, _ := http.DefaultClient.Do(req)
		var res map[string]map[string]interface{}
		_ = json.NewDecoder(resp.Body).Decode(&res)

		item := res["box"]["items"].([]interface{})[0].(map[string]interface{})
		assert.Equal(t, "Updated Watch", item["name"])
		assert.Equal(t, "Updated Smartwatch", item["description"])
		assert.Equal(t, float64(2), item["quantity"])
		assert.Equal(t, "http://example.com/item.jpg", item["image_url"])
	})

	t.Run("unauthorized user cannot update item", func(t *testing.T) {
		email2 := "otheruser+" + timestamp + "@test.com"
		password2 := "pass5678"
		body, _ := json.Marshal(map[string]string{"email": email2, "password": password2})
		http.Post(userBaseURL+"/users/register", "application/json", bytes.NewReader(body))
		body, _ = json.Marshal(map[string]string{"email": email2, "password": password2})
		resp, _ := http.Post(userBaseURL+"/users/login", "application/json", bytes.NewReader(body))

		var res map[string]interface{}
		_ = json.NewDecoder(resp.Body).Decode(&res)
		secondToken := res["token"].(string)

		updateReq := map[string]string{"name": "Hacked"}
		body, _ = json.Marshal(updateReq)
		req, _ := http.NewRequest(http.MethodPatch, boxBaseURL+"/"+boxID+"/items/"+itemID, bytes.NewReader(body))
		req.Header.Set("Authorization", "Bearer "+secondToken)
		req.Header.Set("Content-Type", "application/json")

		resp, err := http.DefaultClient.Do(req)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	})
}
