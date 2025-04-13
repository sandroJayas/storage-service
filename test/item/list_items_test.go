package test

import (
	"encoding/json"
	"github.com/sandroJayas/storage-service/test"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestListItems(t *testing.T) {
	timestamp := time.Now().Format("150405")
	email := "listitems+" + timestamp + "@test.com"
	password := "listitemspass123"
	token := test.RegisterAndLogin(t, email, password)

	sortBoxID := test.CreateSortPackedBox(t, token)

	t.Run("list items returns expected count", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, boxBaseURL+"/"+sortBoxID+"/items", nil)
		req.Header.Set("Authorization", "Bearer "+token)

		resp, err := http.DefaultClient.Do(req)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var res map[string][]map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&res)
		assert.NoError(t, err)
		assert.GreaterOrEqual(t, len(res["items"]), 0)
	})

	t.Run("add multiple items", func(t *testing.T) {
		for i := 0; i < 3; i++ {
			item := map[string]interface{}{
				"name":        "Item " + time.Now().Format("050405"),
				"description": "Desc",
				"quantity":    1,
				"image_url":   "http://example.com/image.jpg",
			}
			test.AddItemToBox(t, token, sortBoxID, item)
		}
	})

	t.Run("list items returns expected count", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, boxBaseURL+"/"+sortBoxID+"/items", nil)
		req.Header.Set("Authorization", "Bearer "+token)

		resp, err := http.DefaultClient.Do(req)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var res map[string][]map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&res)
		assert.NoError(t, err)
		assert.GreaterOrEqual(t, len(res["items"]), 3)
	})

	t.Run("unauthorized request fails", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, boxBaseURL+"/"+sortBoxID+"/items", nil)
		resp, err := http.DefaultClient.Do(req)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})

	t.Run("malformed box ID returns 400", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, boxBaseURL+"/not-a-uuid/items", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		resp, err := http.DefaultClient.Do(req)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("unauthorized user cannot view another user's box", func(t *testing.T) {
		email := "intruder+" + timestamp + "@test.com"
		password := "pass12345678"
		intruderToken := test.RegisterAndLogin(t, email, password)

		req, _ := http.NewRequest(http.MethodGet, boxBaseURL+"/"+sortBoxID+"/items", nil)
		req.Header.Set("Authorization", "Bearer "+intruderToken)
		resp, err := http.DefaultClient.Do(req)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	})
}
