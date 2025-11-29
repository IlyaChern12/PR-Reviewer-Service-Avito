package integration

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"testing"
)

// тестируем получение PR на ревью
func TestGetUserReviewPRs(t *testing.T) {
	// cбрасываем базу
	ResetDB()

	// создаём команду с пользователем
	testTeam := map[string]interface{}{
		"team_name": "team1",
		"members": []map[string]interface{}{
			{
				"user_id":   "u1",
				"username":  "Ilya",
				"is_active": true,
			},
		},
	}
	body, _ := json.Marshal(testTeam)
	resp, err := http.Post(baseURL+"/team/add", "application/json", bytes.NewBuffer(body))
	if err != nil {
		t.Fatalf("request error: %v", err)
	}
	resp.Body.Close()

	// тест кейсы
	cases := []struct {
		name        string
		userID      string
		wantStatus  int
		expectEmpty bool // пустой список PR
		expectError bool // JSON с error
	}{
		{"существующий_пользователь", "u1", http.StatusOK, true, false},
		{"не_существующий_пользователь", "u999", http.StatusNotFound, false, true},
		{"пользователь_без_PR", "u1", http.StatusOK, true, false},
	}

	// посылаем запросы
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {

			reqURL := baseURL + "/users/getReview?user_id=" + c.userID
			resp, err := http.Get(reqURL)
			if err != nil {
				t.Fatalf("request error: %v", err)
			}
			defer resp.Body.Close()

			// assert
			// проверка статуса
			if resp.StatusCode != c.wantStatus {
				t.Fatalf("expected status %d, got %d", c.wantStatus, resp.StatusCode)
			}

			// проверка тела запроса
			bodyBytes, err := io.ReadAll(resp.Body)
			if err != nil {
				t.Fatalf("reading JSON error: %v", err)
			}

			if c.expectError {
				var errJSON map[string]interface{}
				if err := json.Unmarshal(bodyBytes, &errJSON); err != nil {
					t.Fatalf("parsing JSON error: %v\nbody=%s", err, string(bodyBytes))
				}
				if _, ok := errJSON["error"]; !ok {
					t.Fatalf("expected error, got: %#v", errJSON)
				}
				return
			}

			// тело успешного ответа
			var jsonBody map[string]interface{}
			if err := json.Unmarshal(bodyBytes, &jsonBody); err != nil {
				t.Fatalf("parsing JSON error: %v\nbody=%s", err, string(bodyBytes))
			}

			userIDResp, ok := jsonBody["user_id"].(string)
			if !ok || userIDResp != c.userID {
				t.Fatalf("expected user_id=%s, got: %v", c.userID, jsonBody["user_id"])
			}

			prs, ok := jsonBody["pull_requests"].([]interface{})
			if !ok {
				t.Fatalf("expected pull_requests as array, got: %#v", jsonBody["pull_requests"])
			}

			// проверка через количество записей в бд
			var dbCount int
			err = testDB.QueryRow(
				`SELECT COUNT(*) FROM pull_request_reviewers WHERE user_id = $1`, c.userID,
			).Scan(&dbCount)
			if err != nil {
				t.Fatalf("response DB error: %v", err)
			}


			if c.expectEmpty && len(prs) != 0 {
				t.Fatalf("expected empty list of PR, got: %v", prs)
			}

			if !c.expectEmpty && len(prs) != dbCount {
				t.Fatalf("amount of PR is not correct: api=%d, db=%d", len(prs), dbCount)
			}
		})
	}
}

// тестируем установку активности пользователя
func TestSetUserIsActive(t *testing.T) {
	// чистим базу
	ResetDB()

	// создаём команду с пользователем
	teamPayload := map[string]interface{}{
		"team_name": "team1",
		"members": []map[string]interface{}{
			{
				"user_id":   "u1",
				"username":  "Ilya",
				"is_active": true,
			},
		},
	}
	body, _ := json.Marshal(teamPayload)
	resp, err := http.Post(baseURL+"/team/add", "application/json", bytes.NewBuffer(body))
	if err != nil {
		t.Fatalf("request error: %v", err)
	}
	resp.Body.Close()

	// кейсы
	cases := []struct {
		name      string
		payload   map[string]interface{}
		wantCode  int
		checkDB   bool
		wantValue bool
	}{
		{
			"активировать_существующего_пользователя",
			map[string]interface{}{"user_id": "u1", "is_active": true},
			http.StatusOK,
			true,
			true,
		},
		{
			"деактивировать_существующего_пользователя",
			map[string]interface{}{"user_id": "u1", "is_active": false},
			http.StatusOK,
			true,
			false,
		},
		{
			"не_существующий_пользователь",
			map[string]interface{}{"user_id": "u999", "is_active": true},
			http.StatusNotFound,
			false,
			false,
		},
	}

	// применение тест-кейсов
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {

			body, _ := json.Marshal(c.payload)
			resp, err := http.Post(baseURL+"/users/setIsActive", "application/json", bytes.NewBuffer(body))
			if err != nil {
				t.Fatalf("response error: %v", err)
			}
			defer resp.Body.Close()

			// assert
			// проверка статуса
			if resp.StatusCode != c.wantCode {
				t.Fatalf("expected status: %d, got %d", c.wantCode, resp.StatusCode)
			}

			// чтение json
			var jsonBody map[string]interface{}
			if err := json.NewDecoder(resp.Body).Decode(&jsonBody); err != nil {
				t.Fatalf("не удалось прочитать JSON: %v", err)
			}

			// чтение тела запроса
			switch resp.StatusCode {
			case http.StatusOK:
				userObj, ok := jsonBody["user"].(map[string]interface{})
				if !ok {
					t.Fatalf("expected JSON with user, got: %#v", jsonBody)
				}

				if userObj["user_id"] != c.payload["user_id"] {
					t.Fatalf("expected user_id=%v, got %v", c.payload["user_id"], userObj["user_id"])
				}

				if userObj["is_active"] != c.wantValue {
					t.Fatalf("expected is_active=%v, got %v", c.wantValue, userObj["is_active"])
				}

			case http.StatusNotFound:
				errObj, ok := jsonBody["error"].(map[string]interface{})
				if !ok {
					t.Fatalf("expected JSON {error:{...}}, got: %#v", jsonBody)
				}

				if errObj["code"] == nil {
					t.Fatalf("error.code is missing: %#v", jsonBody)
				}

			default:
				t.Fatalf("unexpected status code: %d", resp.StatusCode)
			}


			// проверка записи в бд
			if c.checkDB {
				var dbVal bool
				row := testDB.QueryRow(`SELECT is_active FROM users WHERE user_id = $1`, c.payload["user_id"])
				if err := row.Scan(&dbVal); err != nil {
					t.Fatalf("error query from db: %v", err)
				}

				if dbVal != c.wantValue {
					t.Errorf("expected is_active=%v, got %v", c.wantValue, dbVal)
				}
			}
		})
	}
}
