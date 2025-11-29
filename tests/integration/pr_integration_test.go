package integration

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"testing"
)

// тестируем создание PR с разными сценариями
func TestCreatePR(t *testing.T) {
	// чистим бд
	ResetDB()

	// создаём команду
	teamPayload := map[string]interface{}{
		"team_name": "team_create_pr",
		"members": []map[string]interface{}{
			{"user_id": "u1", "username": "AuthorUser", "is_active": true},
			{"user_id": "u2", "username": "ReviewerUser", "is_active": true},
		},
	}
	body, _ := json.Marshal(teamPayload)
	resp, err := http.Post(baseURL+"/team/add", "application/json", bytes.NewBuffer(body))
	if err != nil {
		t.Fatalf("failed to create team: %v", err)
	}
	defer resp.Body.Close()

	// кейсы
	cases := []struct {
		name    string
		payload map[string]string
		want    int
	}{
		{"успешное_создание", map[string]string{"pull_request_id": "pr-success", "pull_request_name": "add feature", "author_id": "u1"}, http.StatusCreated},
		{"pr_уже_существует", map[string]string{"pull_request_id": "pr-success", "pull_request_name": "duplicate", "author_id": "u1"}, http.StatusConflict},
		{"автор_не_существует", map[string]string{"pull_request_id": "pr-fail", "pull_request_name": "test feature", "author_id": "u999"}, http.StatusNotFound},
	}

	// применяем кейсы
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			body, _ := json.Marshal(c.payload)
			resp, err := http.Post(baseURL+"/pullRequest/create", "application/json", bytes.NewBuffer(body))
			if err != nil {
				t.Fatalf("request error: %v", err)
			}
			defer resp.Body.Close()

			bodyBytes, _ := io.ReadAll(resp.Body)

			// assert
			// проверяем статус код
			if resp.StatusCode != c.want {
				t.Fatalf("expected status %d, got %d, body: %s", c.want, resp.StatusCode, string(bodyBytes))
			}

			// проверка успешного ответа
			if c.want == http.StatusCreated {
				// проверяем тело 
				var okResp struct {
					PR struct {
						PullRequestID   string   `json:"pull_request_id"`
						PullRequestName string   `json:"pull_request_name"`
						AuthorID        string   `json:"author_id"`
						Status          string   `json:"status"`
						AssignedReviewers []string `json:"assigned_reviewers"`
					} `json:"pr"`
				}
				if err := json.Unmarshal(bodyBytes, &okResp); err != nil {
					t.Fatalf("failed to decode success json: %v. raw=%s", err, string(bodyBytes))
				}

				if okResp.PR.PullRequestID != c.payload["pull_request_id"] {
					t.Errorf("expected pull_request_id=%s, got=%s", c.payload["pull_request_id"], okResp.PR.PullRequestID)
				}
				if okResp.PR.PullRequestName != c.payload["pull_request_name"] {
					t.Errorf("expected pull_request_name=%s, got=%s", c.payload["pull_request_name"], okResp.PR.PullRequestName)
				}
				if okResp.PR.AuthorID != c.payload["author_id"] {
					t.Errorf("expected author_id=%s, got=%s", c.payload["author_id"], okResp.PR.AuthorID)
				}

				// проверка БД
				var dbAuthorID, dbPRName string
				err := testDB.QueryRow(
					`SELECT author_id, pull_request_name FROM pull_requests WHERE pull_request_id = $1`,
					c.payload["pull_request_id"],
				).Scan(&dbAuthorID, &dbPRName)
				if err != nil {
					t.Fatalf("pull request not found in DB: %v", err)
				}
				if dbAuthorID != c.payload["author_id"] {
					t.Errorf("DB author_id mismatch: expected %s, got %s", c.payload["author_id"], dbAuthorID)
				}
				if dbPRName != c.payload["pull_request_name"] {
					t.Errorf("DB pull_request_name mismatch: expected %s, got %s", c.payload["pull_request_name"], dbPRName)
				}

				return
			}

			// проверка ошибочного запроса
			var errResp struct {
				Error struct {
					Code    string `json:"code"`
					Message string `json:"message"`
				} `json:"error"`
			}
			if err := json.Unmarshal(bodyBytes, &errResp); err != nil {
				t.Fatalf("failed to decode error json: %v. raw=%s", err, string(bodyBytes))
			}

			if errResp.Error.Code == "" || errResp.Error.Message == "" {
				t.Errorf("expected non-empty error code and message for case %s", c.name)
			}
		})
	}
}
