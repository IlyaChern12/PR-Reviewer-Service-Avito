package integration

import (
	"bytes"
	"encoding/json"
	"net/http"
	"testing"
)

// тестируем создание PR с разными сценариями
func TestCreatePR(t *testing.T) {
	ResetDB() // чистим тестовую БД перед тестом

	// создаём команду через API или напрямую через сервисы
	teamPayload := map[string]interface{}{
		"team_name": "team_create_pr",
		"members": []map[string]interface{}{
			{"user_id": "u10", "username": "AuthorUser", "is_active": true},
			{"user_id": "u11", "username": "ReviewerUser", "is_active": true},
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

	// кейсы
	cases := []struct {
		name    string
		payload map[string]string
		want    int
	}{
		{"успешное_создание", map[string]string{"pull_request_id": "pr-success", "pull_request_name": "add feature", "author_id": "u10"}, http.StatusCreated},
		{"pr_уже_существует", map[string]string{"pull_request_id": "pr-success", "pull_request_name": "duplicate", "author_id": "u10"}, http.StatusConflict},
		{"автор_не_существует", map[string]string{"pull_request_id": "pr-fail", "pull_request_name": "test feature", "author_id": "u999"}, http.StatusNotFound},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			body, _ := json.Marshal(c.payload)
			resp, err := http.Post(baseURL+"/pullRequest/create", "application/json", bytes.NewBuffer(body))
			if err != nil {
				t.Fatalf("request error: %v", err)
			}
			defer func() {
				if err := resp.Body.Close(); err != nil {
					t.Fatalf("failed to close body: %v", err)
				}
			}()

			if resp.StatusCode != c.want {
				t.Errorf("expected %d, got %d", c.want, resp.StatusCode)
			}
		})
	}
}
