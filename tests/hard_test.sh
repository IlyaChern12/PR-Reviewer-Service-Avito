#!/usr/bin/env bash

BASE_URL="http://localhost:8080"
LOG_DIR="./logs"
mkdir -p "$LOG_DIR"

LOG_DETAIL="$LOG_DIR/requests.json"
LOG_TIMES="$LOG_DIR/times.log"
LOG_CODES="$LOG_DIR/codes.log"

>> "$LOG_DETAIL"
>> "$LOG_TIMES"
>> "$LOG_CODES"

echo "=== запуск нагрузочного теста на $BASE_URL ==="

SUFFIX=$RANDOM
NUM_TEAMS=20
USERS_PER_TEAM=10
PRS_PER_TEAM=3
RPS_DELAY=0.2

request() {
  local endpoint=$1
  local method=$2
  local url=$3
  local body=$4

  local start end duration code body_response
  start=$(($(date +%s%N)/1000000))

  local tmpfile=$(mktemp)
  if [ "$method" == "GET" ]; then
    code=$(curl -s -w "%{http_code}" -o "$tmpfile" "$url")
  else
    body_single_line=$(echo "$body" | jq -c . 2>/dev/null || echo "{}")
    code=$(curl -s -w "%{http_code}" -o "$tmpfile" -X "$method" \
      -H "Content-Type: application/json" \
      -d "$body_single_line" "$url")
  fi

  end=$(($(date +%s%N)/1000000))
  duration=$((end - start))
  body_response=$(cat "$tmpfile")
  rm "$tmpfile"

  jq -n \
    --arg ts "$(date '+%Y-%m-%d %H:%M:%S')" \
    --arg endpoint "$endpoint" \
    --arg method "$method" \
    --arg url "$url" \
    --arg request_body "$body" \
    --arg code "$code" \
    --arg response_body "$body_response" \
    --arg duration "$duration" \
    '{timestamp: $ts, endpoint: $endpoint, method: $method, url: $url, request_body: $request_body, response_code: ($code|tonumber), response_body: $response_body, duration_ms: ($duration|tonumber)}' \
    >> "$LOG_DETAIL"

  echo "$duration" >> "$LOG_TIMES"
  echo "$endpoint,$code,$duration" >> "$LOG_CODES"

  sleep $RPS_DELAY
}

# создаем команды и пользователей
for t in $(seq 1 $NUM_TEAMS); do
  TEAM_NAME="load_team_${t}_$SUFFIX"
  MEMBERS="["
  for u in $(seq 1 $USERS_PER_TEAM); do
    MEMBERS="$MEMBERS{\"user_id\":\"u${t}_${u}_$SUFFIX\",\"username\":\"User${u}\",\"is_active\":true},"
  done
  MEMBERS="${MEMBERS%,}]"
  TEAM_PAYLOAD="{\"team_name\":\"$TEAM_NAME\",\"members\":$MEMBERS}"

  request "team/add" POST "$BASE_URL/team/add" "$TEAM_PAYLOAD"
done

# создаем PR
for t in $(seq 1 $NUM_TEAMS); do
  for p in $(seq 1 $PRS_PER_TEAM); do
    PR_ID="pr${t}_${p}_$SUFFIX"
    AUTHOR_ID="u${t}_1_$SUFFIX"
    PR_PAYLOAD="{\"pull_request_id\":\"$PR_ID\",\"pull_request_name\":\"Feature ${p}\",\"author_id\":\"$AUTHOR_ID\"}"
    request "pullRequest/create" POST "$BASE_URL/pullRequest/create" "$PR_PAYLOAD"
  done
done

# reassign: u2 -> u3
for t in $(seq 1 $NUM_TEAMS); do
  for p in $(seq 1 $PRS_PER_TEAM); do
    PR_ID="pr${t}_${p}_$SUFFIX"
    OLD_USER="u${t}_2_$SUFFIX"
    NEW_USER="u${t}_3_$SUFFIX"
    REASSIGN="{\"pull_request_id\":\"$PR_ID\",\"old_user_id\":\"$OLD_USER\",\"new_user_id\":\"$NEW_USER\"}"
    request "pullRequest/reassign" POST "$BASE_URL/pullRequest/reassign" "$REASSIGN"
  done
done

# users/getReview
for t in $(seq 1 $NUM_TEAMS); do
  for u in $(seq 1 $USERS_PER_TEAM); do
    USER_ID="u${t}_${u}_$SUFFIX"
    request "users/getReview" GET "$BASE_URL/users/getReview?user_id=$USER_ID" "{}"
  done
done

# users/setIsActive
for t in $(seq 1 $NUM_TEAMS); do
  for u in $(seq 1 $USERS_PER_TEAM); do
    USER_ID="u${t}_${u}_$SUFFIX"
    SET_ACTIVE="{\"user_id\":\"$USER_ID\",\"is_active\":true}"
    request "users/setIsActive" POST "$BASE_URL/users/setIsActive" "$SET_ACTIVE"
  done
done

# get команды
for t in $(seq 1 $NUM_TEAMS); do
  TEAM_NAME="load_team_${t}_$SUFFIX"
  request "team/get" GET "$BASE_URL/team/get?team_name=$TEAM_NAME" "{}"
done

# деактивация команды
for t in $(seq 1 $NUM_TEAMS); do
  TEAM_NAME="load_team_${t}_$SUFFIX"
  DEACT_PAYLOAD="{\"team_name\":\"$TEAM_NAME\"}"
  request "team/deactivate" POST "$BASE_URL/team/deactivate" "$DEACT_PAYLOAD"
done

# merge PR
for t in $(seq 1 $NUM_TEAMS); do
  for p in $(seq 1 $PRS_PER_TEAM); do
    MERGE_PAYLOAD="{\"pull_request_id\":\"pr${t}_${p}_$SUFFIX\"}"
    request "pullRequest/merge" POST "$BASE_URL/pullRequest/merge" "$MERGE_PAYLOAD"
  done
done

# health и stats
request "health" GET "$BASE_URL/health" "{}"
request "stats" GET "$BASE_URL/stats" "{}"

# метрики
TOTAL_REQS=$(wc -l < "$LOG_CODES")
AVG_TIME=$(awk '{sum+=$1} END {print sum/NR}' "$LOG_TIMES")
MAX_TIME=$(awk 'max<$1 {max=$1} END {print max}' "$LOG_TIMES")
MIN_TIME=$(awk 'min=="" || min>$1 {min=$1} END {print min}' "$LOG_TIMES")
SUCCESS=$(awk -F, '$2 ~ /^2/ {count++} END {print count}' "$LOG_CODES")
FAIL=$((TOTAL_REQS - SUCCESS))
SUCCESS_RATE=$(awk "BEGIN {printf \"%.3f\", ($SUCCESS/$TOTAL_REQS)*100}")

echo
echo "=== резюме нагрузочного теста ==="
echo "Всего запросов: $TOTAL_REQS"
echo "Среднее время отклика: $AVG_TIME ms"
echo "Максимальное время отклика: $MAX_TIME ms"
echo "Минимальное время отклика: $MIN_TIME ms"
echo "Процент успешных запросов: $SUCCESS_RATE%"
echo "Количество неуспешных запросов: $FAIL"

echo
echo "=== sli по endpoint ==="
awk -F, '
{code_sum[$1]+=($2~/^2/?1:0); total[$1]++; time_sum[$1]+=$3}
END {
  for (e in total) {
    printf "%s: avg_time=%.2fms, success=%.2f%% (%d/%d)\n",
           e, time_sum[e]/total[e], code_sum[e]/total[e]*100, code_sum[e], total[e]
  }
}' "$LOG_CODES"