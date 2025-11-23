package integration

import (
	"bytes"
	"encoding/json"
	"net/http"
	"os"
	"testing"
)

var baseURL string

func init() {
	baseURL = os.Getenv("PR_SERVICE_BASE_URL")
	if baseURL == "" {
		baseURL = "http://localhost:8080"
	}
}

// тестируем получение PR на ревью
func TestGetUserReviewPRs(t *testing.T) {
	ResetDB()
	// создаём команду с пользователем заранее
	teamPayload := map[string]interface{}{
		"team_name": "team1",
		"members": []map[string]interface{}{
			{
				"user_id":   "u1",
				"username":  "Alice",
				"is_active": true,
			},
		},
	}
	body, _ := json.Marshal(teamPayload)
	resp, err := http.Post(baseURL+"/team/add", "application/json", bytes.NewBuffer(body))
	if err != nil {
		t.Fatalf("request error: %v", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			t.Fatalf("failed to close body: %v", err)
		}
	}()

	cases := []struct {
		name   string
		userID string
		want   int
	}{
		{"существующий_пользователь", "u1", http.StatusOK},
		{"не_существующий_пользователь", "u999", http.StatusNotFound},
		{"пользователь_без_PR", "u1", http.StatusOK},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			reqURL := baseURL + "/users/getReview?user_id=" + c.userID
			resp, err := http.Get(reqURL)
			if err != nil {
				t.Fatalf("ошибка запроса: %v", err)
			}
			defer func() {
				if err := resp.Body.Close(); err != nil {
					t.Fatalf("failed to close body: %v", err)
				}
			}()

			if resp.StatusCode != c.want {
				t.Errorf("ожидали %d, получили %d", c.want, resp.StatusCode)
			}
		})
	}
}

// тестируем установку активности пользователя
func TestSetUserIsActive(t *testing.T) {
	// создаём команду с пользователем заранее
	ResetDB()
	teamPayload := map[string]interface{}{
		"team_name": "team1",
		"members": []map[string]interface{}{
			{
				"user_id":   "u2",
				"username":  "Bob",
				"is_active": true,
			},
		},
	}
	body, _ := json.Marshal(teamPayload)
	resp, err := http.Post(baseURL+"/team/add", "application/json", bytes.NewBuffer(body))
	if err != nil {
		t.Fatalf("request error: %v", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			t.Fatalf("failed to close body: %v", err)
		}
	}()

	cases := []struct {
		name    string
		payload map[string]interface{}
		want    int
	}{
		{"активировать_существующего_пользователя", map[string]interface{}{"user_id": "u2", "is_active": true}, http.StatusOK},
		{"деактивировать_существующего_пользователя", map[string]interface{}{"user_id": "u2", "is_active": false}, http.StatusOK},
		{"не_существующий_пользователь", map[string]interface{}{"user_id": "u999", "is_active": true}, http.StatusNotFound},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			body, _ := json.Marshal(c.payload)
			reqURL := baseURL + "/users/setIsActive"

			resp, err := http.Post(reqURL, "application/json", bytes.NewBuffer(body))
			if err != nil {
				t.Fatalf("ошибка запроса: %v", err)
			}
			defer func() {
				if err := resp.Body.Close(); err != nil {
					t.Fatalf("failed to close body: %v", err)
				}
			}()

			if resp.StatusCode != c.want {
				t.Errorf("ожидали %d, получили %d", c.want, resp.StatusCode)
			}
		})
	}
}
