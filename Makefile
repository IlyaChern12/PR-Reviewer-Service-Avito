# сборка образов
build:
	docker-compose build

# запуск приложения
up:
	docker-compose up -d

# остановка
down:
	docker-compose down

# пересборка с запуском
rebuild: build up

# вывод логов
logs:
	docker-compose logs -f

# очистка всего
clean:
	docker-compose down -v --rmi all

