package integration

import (
	"bytes"
	"encoding/json"
	"net/http"
	"testing"
)

// тестируем получение команды
func TestGetTeam(t *testing.T) {
	// создаём команду заранее, чтобы точно существовала
	ResetDB() 
	teamPayload := map[string]interface{}{
		"team_name": "test_team",
		"members": []map[string]interface{}{
			{"user_id": "u1", "username": "Alice", "is_active": true},
			{"user_id": "u2", "username": "Bob", "is_active": true},
		},
	}
	body, _ := json.Marshal(teamPayload)
	resp, err := http.Post(baseURL+"/team/add", "application/json", bytes.NewBuffer(body))
	if err != nil {
		t.Fatalf("failed to create team: %v", err)
	}
	resp.Body.Close()

	cases := []struct {
		name string
		url  string
		want int
	}{
		{"пустое_имя", baseURL + "/team/get?team_name=", http.StatusBadRequest},
		{"существующая_команда", baseURL + "/team/get?team_name=test_team", http.StatusOK},
		{"не_существующая_команда", baseURL + "/team/get?team_name=no_such_team", http.StatusNotFound},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			resp, err := http.Get(c.url)
			if err != nil {
				t.Fatalf("request error: %v", err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != c.want {
				t.Errorf("expected %d, got %d", c.want, resp.StatusCode)
			}
		})
	}
}

// тестируем деактивацию команды
func TestDeactivateTeam(t *testing.T) {
	ResetDB() 
	// создаём команду заранее
	teamPayload := map[string]interface{}{
		"team_name": "deactivate_team",
		"members": []map[string]interface{}{
			{"user_id": "u3", "username": "Charlie", "is_active": true},
			{"user_id": "u4", "username": "Diana", "is_active": true},
		},
	}
	body, _ := json.Marshal(teamPayload)
	resp, err := http.Post(baseURL+"/team/add", "application/json", bytes.NewBuffer(body))
	if err != nil {
		t.Fatalf("failed to create team: %v", err)
	}
	resp.Body.Close()

	cases := []struct {
		name string
		url  string
		want int
	}{
		{"пустое_имя", baseURL + "/team/deactivate?team_name=", http.StatusBadRequest},
		{"существующая_команда", baseURL + "/team/deactivate?team_name=deactivate_team", http.StatusOK},
		{"не_существующая_команда", baseURL + "/team/deactivate?team_name=unknown_team", http.StatusNotFound},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			resp, err := http.Post(c.url, "application/json", bytes.NewBuffer(nil))
			if err != nil {
				t.Fatalf("request error: %v", err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != c.want {
				t.Errorf("expected %d, got %d", c.want, resp.StatusCode)
			}
		})
	}
}