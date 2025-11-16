.PHONY: build run clean docker-up docker-down

# Сборка проекта
build:
	go build -o bin/main .

# Запуск локально (требует запущенной PostgreSQL)
run:
	go run .

# Очистка
clean:
	rm -rf bin/

# Запуск через docker-compose
docker-up:
	docker-compose up --build

# Остановка docker-compose
docker-down:
	docker-compose down

# Остановка и удаление volumes
docker-clean:
	docker-compose down -v


