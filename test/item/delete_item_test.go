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

func TestDeleteItem(t *testing.T) {
	timestamp := time.Now().Format("150405")
	email := "deleteitem+" + timestamp + "@test.com"
	password := "deleteitempass123"
	token := test.RegisterAndLogin(t, email, password)

	var boxID string
	var itemID string

	t.Run("setup - create sort-packed box", func(t *testing.T) {
		body, _ := json.Marshal(map[string]string{
			"packing_mode": "sort",
			"item_name":    "initial",
			"item_note":    "ignored",
		})
		req, _ := http.NewRequest(http.MethodPost, boxBaseURL, bytes.NewReader(body))
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", "application/json")

		resp, err := http.DefaultClient.Do(req)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusCreated, resp.StatusCode)

		var res map[string]string
		_ = json.NewDecoder(resp.Body).Decode(&res)
		boxID = res["id"]
		assert.NotEmpty(t, boxID)
	})

	t.Run("setup - add item", func(t *testing.T) {
		item := map[string]interface{}{
			"name":        "ToDelete",
			"description": "This will be deleted",
			"quantity":    1,
		}
		body, _ := json.Marshal(item)
		req, _ := http.NewRequest(http.MethodPost, boxBaseURL+"/"+boxID+"/items", bytes.NewReader(body))
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", "application/json")

		resp, err := http.DefaultClient.Do(req)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusCreated, resp.StatusCode)

		var res map[string]string
		_ = json.NewDecoder(resp.Body).Decode(&res)
		itemID = res["id"]
		assert.NotEmpty(t, itemID)
	})

	t.Run("get item successfully", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, "http://localhost:8080/items/"+itemID, nil)
		req.Header.Set("Authorization", "Bearer "+token)

		resp, err := http.DefaultClient.Do(req)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("delete item successfully", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodDelete, "http://localhost:8080/items/"+itemID, nil)
		req.Header.Set("Authorization", "Bearer "+token)

		resp, err := http.DefaultClient.Do(req)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var res map[string]string
		_ = json.NewDecoder(resp.Body).Decode(&res)
		assert.Equal(t, "Item deleted successfully", res["message"])
	})

	t.Run("get item returns no item", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, "http://localhost:8080/items/"+itemID, nil)
		req.Header.Set("Authorization", "Bearer "+token)

		resp, err := http.DefaultClient.Do(req)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	})

	t.Run("deleting already deleted item returns not found", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodDelete, "http://localhost:8080/items/"+itemID, nil)
		req.Header.Set("Authorization", "Bearer "+token)

		resp, _ := http.DefaultClient.Do(req)
		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	})

	t.Run("unauthorized delete attempt", func(t *testing.T) {
		hackerEmail := "hacker+" + timestamp + "@test.com"
		hackerPass := "hack12345678"
		hackerToken := test.RegisterAndLogin(t, hackerEmail, hackerPass)

		req, _ := http.NewRequest(http.MethodDelete, "http://localhost:8080/items/"+itemID, nil)
		req.Header.Set("Authorization", "Bearer "+hackerToken)

		resp, _ := http.DefaultClient.Do(req)
		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	})

	t.Run("delete with malformed item ID", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodDelete, "http://localhost:8080/items/not-a-uuid", nil)
		req.Header.Set("Authorization", "Bearer "+token)

		resp, _ := http.DefaultClient.Do(req)
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})
}
