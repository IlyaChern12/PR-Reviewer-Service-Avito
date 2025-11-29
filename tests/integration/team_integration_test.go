package integration

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"testing"
)

// тестируем получение команды
func TestGetTeam(t *testing.T) {
	// чистим базу
	ResetDB()

	// создаем тестовую команду
	teamTest := map[string]interface{}{
		"team_name": "test_team",
		"members": []map[string]interface{}{
			{"user_id": "u1", "username": "Ilya", "is_active": true},
			{"user_id": "u2", "username": "Sasha", "is_active": true},
		},
	}
	body, _ := json.Marshal(teamTest)
	resp, err := http.Post(baseURL+"/team/add", "application/json", bytes.NewBuffer(body))
	if err != nil {
		t.Fatalf("failed to create team: %v", err)
	}
	defer resp.Body.Close()

	// тест-кейсы
	cases := []struct {
		name       string
		url        string
		wantStatus int
		wantError  string
	}{
		{"пустое_имя", baseURL + "/team/get?team_name=", http.StatusBadRequest, "team name is empty"},
		{"существующая_команда", baseURL + "/team/get?team_name=test_team", http.StatusOK, ""},
		{"не_существующая_команда", baseURL + "/team/get?team_name=no_such_team", http.StatusNotFound, "resource not found"},
	}

	// запускаем какждый кейс
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			resp, err := http.Get(c.url)
			if err != nil {
				t.Fatalf("request error: %v", err)
			}
			defer resp.Body.Close()

			// assert
			// проверка статуса
			if resp.StatusCode != c.wantStatus {
				t.Fatalf("expected %d, got %d", c.wantStatus, resp.StatusCode)
			}

			bodyBytes, err := io.ReadAll(resp.Body)
			if err != nil {
				t.Fatalf("failed read body: %v", err)
			}

			// чтение тела успешного ответа
			if c.wantStatus == http.StatusOK {
				var jsonResp struct {
					TeamName string `json:"team_name"`
					Members  []struct {
						UserID   string `json:"user_id"`
						Username string `json:"username"`
						IsActive bool   `json:"is_active"`
					} `json:"members"`
				}

				if err := json.Unmarshal(bodyBytes, &jsonResp); err != nil {
					t.Fatalf("failed to decode json: %v", err)
				}

				if jsonResp.TeamName != "test_team" {
					t.Errorf("expected team_name test_team, got %s", jsonResp.TeamName)
				}

				if len(jsonResp.Members) != 2 {
					t.Errorf("expected 2 members, got %d", len(jsonResp.Members))
				}

				userMap := map[string]string{"u1": "Ilya", "u2": "Sasha"}
				for _, m := range jsonResp.Members {
					expected, ok := userMap[m.UserID]
					if !ok {
						t.Errorf("unexpected user %s", m.UserID)
						continue
					}
					if m.Username != expected {
						t.Errorf("expected %s, got %s", expected, m.Username)
					}
					if !m.IsActive {
						t.Errorf("expected user %s to be active", m.UserID)
					}
				}

				return
			}

			// чтение тела отрицательного ответа
			var errResp struct {
				Error struct {
					Code    string `json:"code"`
					Message string `json:"message"`
				} `json:"error"`
			}

			if err := json.Unmarshal(bodyBytes, &errResp); err != nil {
				t.Fatalf("failed to decode error json: %v. raw: %s", err, string(bodyBytes))
			}

			if errResp.Error.Message != c.wantError {
				t.Errorf("expected error message %q, got %q", c.wantError, errResp.Error.Message)
			}
		})
	}
}


// тестируем деактивацию команды
func TestDeactivateTeam(t *testing.T) {
	// сбрасываем бд
    ResetDB()

    // создаём команду заранее
    teamPayload := map[string]interface{}{
        "team_name": "deactivate_team",
        "members": []map[string]interface{}{
            {"user_id": "u1", "username": "Ivan", "is_active": true},
            {"user_id": "u2", "username": "Vasya", "is_active": true},
        },
    }
    body, _ := json.Marshal(teamPayload)
    resp, err := http.Post(baseURL+"/team/add", "application/json", bytes.NewBuffer(body))
    if err != nil {
        t.Fatalf("failed to create team: %v", err)
    }
    defer resp.Body.Close()

	// тест-кейсы
    cases := []struct {
        name       string
        jsonBody   map[string]string
        wantStatus int
        wantError  string
    }{
        {"пустое_имя", map[string]string{"team_name": ""}, http.StatusBadRequest, "team_name is empty"},
        {"существующая_команда", map[string]string{"team_name": "deactivate_team"}, http.StatusOK, ""},
        {"не_существующая_команда", map[string]string{"team_name": "unknown_team"}, http.StatusNotFound, "team not found"},
    }

	// применяем кейсы
    for _, c := range cases {
        t.Run(c.name, func(t *testing.T) {

            reqBody, _ := json.Marshal(c.jsonBody)
            resp, err := http.Post(baseURL+"/team/deactivate", "application/json", bytes.NewBuffer(reqBody))
            if err != nil {
                t.Fatalf("request error: %v", err)
            }
            defer resp.Body.Close()

            bodyBytes, _ := io.ReadAll(resp.Body)

			// assert
			// проверка статус кода
            if resp.StatusCode != c.wantStatus {
                t.Fatalf("expected %d, got %d", c.wantStatus, resp.StatusCode)
            }

			// проверяем ответ на успешный запрос
            if c.wantStatus == http.StatusOK {

                var okResp struct {
                    Message string `json:"message"`
                }
                if err := json.Unmarshal(bodyBytes, &okResp); err != nil {
                    t.Fatalf("failed to decode success json: %v. raw=%s", err, string(bodyBytes))
                }

                expectedMsg := "team deactivated and PR reviewers reassigned"
                if okResp.Message != expectedMsg {
                    t.Errorf("expected message=%q, got=%q", expectedMsg, okResp.Message)
                }

                // проверяем изменение состояния в бд
                rows, err := testDB.Query(
                    `SELECT user_id, is_active FROM users WHERE team_name = $1`,
                    "deactivate_team",
                )
                if err != nil {
                    t.Fatalf("failed to query users: %v", err)
                }
                defer rows.Close()

                found := 0
                for rows.Next() {
                    var userID string
                    var isActive bool
                    if err := rows.Scan(&userID, &isActive); err != nil {
                        t.Fatalf("scan error: %v", err)
                    }
                    found++
                    if isActive {
                        t.Errorf("user %s is still active, expected inactive", userID)
                    }
                }

                if found == 0 {
                    t.Errorf("no users found in DB for team deactivate_team")
                }

                return
            }

            // проверяем ошибочный ответ на запрос
            var errResp struct {
                Error string `json:"error"`
            }

            if err := json.Unmarshal(bodyBytes, &errResp); err != nil {
                t.Fatalf("failed to decode error json: %v. raw=%s", err, string(bodyBytes))
            }

            if errResp.Error != c.wantError {
                t.Errorf("expected error=%q, got=%q", c.wantError, errResp.Error)
            }
        })
    }
}
