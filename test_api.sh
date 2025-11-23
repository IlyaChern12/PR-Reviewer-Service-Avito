#!/usr/bin/env bash
set -e

BASE_URL="http://localhost:8080"   # ← при необходимости поменяй

echo "=== 1. Создание команды ==="
curl -s -X POST "$BASE_URL/team/add" \
  -H "Content-Type: application/json" \
  -d '{
        "team_name": "backend",
        "members": [
          { "user_id": "u1", "username": "Alice", "is_active": true },
          { "user_id": "u2", "username": "Bob", "is_active": true },
          { "user_id": "u3", "username": "Charlie", "is_active": true }
        ]
      }' | jq .

echo "=== 1.1 Повторное создание той же команды (ожидаем TEAM_EXISTS) ==="
curl -s -X POST "$BASE_URL/team/add" \
  -H "Content-Type: application/json" \
  -d '{
        "team_name": "backend",
        "members": []
      }' | jq .

echo "=== 2. Получение команды ==="
curl -s -G "$BASE_URL/team/get" \
  --data-urlencode "team_name=backend" | jq .

echo "=== 3. Деактивация пользователя u2 ==="
curl -s -X POST "$BASE_URL/users/setIsActive" \
  -H "Content-Type: application/json" \
  -d '{ "user_id": "u2", "is_active": false }' | jq .

echo "=== 4. Создание PR pr-1001 автором u1 ==="
curl -s -X POST "$BASE_URL/pullRequest/create" \
  -H "Content-Type: application/json" \
  -d '{
        "pull_request_id": "pr-1001",
        "pull_request_name": "Add search",
        "author_id": "u1"
      }' | jq .

echo "=== 4.1 Повторное создание PR (ожидаем PR_EXISTS) ==="
curl -s -X POST "$BASE_URL/pullRequest/create" \
  -H "Content-Type: application/json" \
  -d '{
        "pull_request_id": "pr-1001",
        "pull_request_name": "Duplicate",
        "author_id": "u1"
      }' | jq .

echo "=== 5. Переназначение ревьювера (old_user_id=u3) ==="
curl -s -X POST "$BASE_URL/pullRequest/reassign" \
  -H "Content-Type: application/json" \
  -d '{
        "pull_request_id": "pr-1001",
        "old_user_id": "u3"
      }' | jq .

echo "=== 5.1 Переназначение несуществующего ревьювера (ожидаем NOT_ASSIGNED) ==="
curl -s -X POST "$BASE_URL/pullRequest/reassign" \
  -H "Content-Type: application/json" \
  -d '{
        "pull_request_id": "pr-1001",
        "old_user_id": "u999"
      }' | jq .

echo "=== 6. Получение PR'ов, где пользователь назначен ревьювером ==="
curl -s -G "$BASE_URL/users/getReview" \
  --data-urlencode "user_id=u3" | jq .

echo "=== 7. Merge PR pr-1001 ==="
curl -s -X POST "$BASE_URL/pullRequest/merge" \
  -H "Content-Type: application/json" \
  -d '{ "pull_request_id": "pr-1001" }' | jq .

echo "=== 7.1 Повторный merge (идемпотентность) ==="
curl -s -X POST "$BASE_URL/pullRequest/merge" \
  -H "Content-Type: application/json" \
  -d '{ "pull_request_id": "pr-1001" }' | jq .

echo "=== 8. Переназначение после MERGED (ожидаем PR_MERGED) ==="
curl -s -X POST "$BASE_URL/pullRequest/reassign" \
  -H "Content-Type: application/json" \
  -d '{
        "pull_request_id": "pr-1001",
        "old_user_id": "u1"
      }' | jq .
