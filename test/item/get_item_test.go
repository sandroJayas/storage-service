package test

import (
	"encoding/json"
	"github.com/sandroJayas/storage-service/test"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestGetItem(t *testing.T) {
	timestamp := time.Now().Format("150405")
	email := "getitem+" + timestamp + "@test.com"
	password := "getitempass123"
	token := test.RegisterAndLogin(t, email, password)

	var itemID string

	t.Run("setup - create box and add item", func(t *testing.T) {
		boxID := test.CreateSortPackedBox(t, token)

		itemReq := map[string]interface{}{
			"name":        "Book",
			"description": "A good book",
			"quantity":    1,
			"image_url":   "http://example.com/book.jpg",
		}
		res := test.AddItemToBox(t, token, boxID, itemReq)
		itemID = res["id"]
		assert.NotEmpty(t, itemID)
	})

	t.Run("get item successfully", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, "http://localhost:8080/items/"+itemID, nil)
		req.Header.Set("Authorization", "Bearer "+token)

		resp, err := http.DefaultClient.Do(req)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var item map[string]interface{}
		_ = json.NewDecoder(resp.Body).Decode(&item)

		assert.Equal(t, itemID, item["id"])
		assert.Equal(t, "Book", item["name"])
		assert.Equal(t, "A good book", item["description"])
		assert.Equal(t, float64(1), item["quantity"]) // JSON numbers are decoded as float64
		assert.Equal(t, "http://example.com/book.jpg", item["image_url"])
	})

	t.Run("fail with invalid item ID", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, "http://localhost:8080/items/not-a-uuid", nil)
		req.Header.Set("Authorization", "Bearer "+token)

		resp, err := http.DefaultClient.Do(req)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("fail when item does not exist", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, "http://localhost:8080/items/123e4567-e89b-12d3-a456-426614174999", nil)
		req.Header.Set("Authorization", "Bearer "+token)

		resp, err := http.DefaultClient.Do(req)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	})

	t.Run("unauthorized user cannot access another user's item", func(t *testing.T) {
		attackerToken := test.RegisterAndLogin(t, "intruder+"+timestamp+"@test.com", "pass45678987")
		req, _ := http.NewRequest(http.MethodGet, "http://localhost:8080/items/"+itemID, nil)
		req.Header.Set("Authorization", "Bearer "+attackerToken)

		resp, err := http.DefaultClient.Do(req)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	})
}
