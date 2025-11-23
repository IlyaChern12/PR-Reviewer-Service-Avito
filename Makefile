# сборка образов
build:
	docker-compose build

# запуск приложения
up:
	docker-compose up -d db app

# остановка
down:
	docker-compose down

# пересборка с запуском
rebuild: clean build up migrate-up

# вывод логов
logs:
	docker-compose logs -f

# очистка всего
clean:
	docker-compose down -v --rmi all

# применить миграции
migrate-up:
	docker-compose exec app sh -c '\
	    until pg_isready -h $$DB_HOST -U $$DB_USER -d $$DB_NAME; do echo waiting for db; sleep 1; done; \
	    migrate -path /app/migrations -database "postgres://$$DB_USER:$$DB_PASSWORD@$$DB_HOST:5432/$$DB_NAME?sslmode=disable" up'

# откат миграций
migrate-down:
	docker-compose exec app sh -c '\
	    until pg_isready -h $$DB_HOST -U $$DB_USER -d $$DB_NAME; do echo waiting for db; sleep 1; done; \
	    migrate -path /app/migrations -database "postgres://$$DB_USER:$$DB_PASSWORD@$$DB_HOST:5432/$$DB_NAME?sslmode=disable" down'

# линтер
lint:
	golangci-lint run ./...

# интеграционные тесты
test:
	@export $(shell grep TEST_DATABASE_URL .env) && go test -v ./tests/integration
