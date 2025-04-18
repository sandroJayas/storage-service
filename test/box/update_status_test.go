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

func TestUpdateStatus(t *testing.T) {
	timestamp := time.Now().Format("150405")

	email := "user+" + timestamp + "@test.com"
	emailEmployee := "admin+" + timestamp + "@test.com"
	password := "supersecure123"

	var boxID string
	token := test.RegisterAndLogin(t, email, password)
	tokenEmployee := test.RegisterAndLoginEmployee(t, emailEmployee, password)

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

	t.Run("update - valid status change", func(t *testing.T) {
		body, _ := json.Marshal(map[string]string{
			"status": "pending_pickup",
		})
		req, _ := http.NewRequest(http.MethodPatch, boxBaseURL+"/"+boxID+"/status", bytes.NewReader(body))
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", "application/json")

		resp, err := http.DefaultClient.Do(req)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("verify status changed", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, boxBaseURL+"/"+boxID, nil)
		req.Header.Set("Authorization", "Bearer "+token)

		resp, err := http.DefaultClient.Do(req)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var res map[string]map[string]interface{}
		_ = json.NewDecoder(resp.Body).Decode(&res)
		assert.Equal(t, "pending_pickup", res["box"]["status"])
	})

	t.Run("update - returned - forbidden status for non-employee", func(t *testing.T) {
		body, _ := json.Marshal(map[string]string{
			"status": "returned",
		})
		req, _ := http.NewRequest(http.MethodPatch, boxBaseURL+"/"+boxID+"/status", bytes.NewReader(body))
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", "application/json")

		resp, err := http.DefaultClient.Do(req)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusForbidden, resp.StatusCode)
	})

	t.Run("update - stored- forbidden status for non-employee", func(t *testing.T) {
		body, _ := json.Marshal(map[string]string{
			"status": "stored",
		})
		req, _ := http.NewRequest(http.MethodPatch, boxBaseURL+"/"+boxID+"/status", bytes.NewReader(body))
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", "application/json")

		resp, err := http.DefaultClient.Do(req)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusForbidden, resp.StatusCode)
	})

	t.Run("update - stored status for employee", func(t *testing.T) {
		body, _ := json.Marshal(map[string]string{
			"status": "stored",
		})
		req, _ := http.NewRequest(http.MethodPatch, boxBaseURL+"/"+boxID+"/status", bytes.NewReader(body))
		req.Header.Set("Authorization", "Bearer "+tokenEmployee)
		req.Header.Set("Content-Type", "application/json")

		resp, err := http.DefaultClient.Do(req)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("update - returned status for employee", func(t *testing.T) {
		body, _ := json.Marshal(map[string]string{
			"status": "returned",
		})
		req, _ := http.NewRequest(http.MethodPatch, boxBaseURL+"/"+boxID+"/status", bytes.NewReader(body))
		req.Header.Set("Authorization", "Bearer "+tokenEmployee)
		req.Header.Set("Content-Type", "application/json")

		resp, err := http.DefaultClient.Do(req)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("update - invalid status", func(t *testing.T) {
		body, _ := json.Marshal(map[string]string{
			"status": "unknown_status",
		})
		req, _ := http.NewRequest(http.MethodPatch, boxBaseURL+"/"+boxID+"/status", bytes.NewReader(body))
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", "application/json")

		resp, err := http.DefaultClient.Do(req)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("update - invalid box ID", func(t *testing.T) {
		body, _ := json.Marshal(map[string]string{
			"status": "in_storage",
		})
		req, _ := http.NewRequest(http.MethodPatch, boxBaseURL+"/invalid-box-id/status", bytes.NewReader(body))
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", "application/json")

		resp, err := http.DefaultClient.Do(req)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("update - without token (unauthorized)", func(t *testing.T) {
		body, _ := json.Marshal(map[string]string{
			"status": "in_storage",
		})
		req, _ := http.NewRequest(http.MethodPatch, boxBaseURL+"/"+boxID+"/status", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err := http.DefaultClient.Do(req)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})
	t.Run("foreign user - cannot update someone else's box", func(t *testing.T) {
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

		// Try to update User A’s box using User B’s token
		updateReq := map[string]string{
			"status": "pending_pack",
		}
		body, _ = json.Marshal(updateReq)
		url := fmt.Sprintf("%s/%s/status", boxBaseURL, boxID)

		req, _ := http.NewRequest(http.MethodPatch, url, bytes.NewReader(body))
		req.Header.Set("Authorization", "Bearer "+foreignToken)
		req.Header.Set("Content-Type", "application/json")

		resp, err = http.DefaultClient.Do(req)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	})

}
