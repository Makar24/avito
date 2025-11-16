# Сервис назначения ревьюеров для Pull Request'ов

Микросервис для автоматического назначения ревьюеров на Pull Request'ы и управления командами и участниками.

## Требования

- Go 1.21+
- Docker и Docker Compose
- PostgreSQL (если запуск без Docker)

## Быстрый старт

### Запуск через Docker Compose

```bash
docker-compose up
```

Сервис будет доступен на `http://localhost:8080`

### Локальный запуск

1. Убедитесь, что PostgreSQL запущен и доступен
2. Установите зависимости:
```bash
go mod download
# или
go mod tidy
```
3. Запустите сервис:
```bash
make run
# или
go run .
```

**Примечание**: Если вы видите ошибки линтера о недостающих пакетах, выполните `go mod download` для загрузки зависимостей.

## API Endpoints

### Команды

- `POST /team/add` - Создать команду с участниками (создаёт/обновляет пользователей)
  ```json
  {
    "team_name": "payments",
    "members": [
      {
        "user_id": "u1",
        "username": "Alice",
        "is_active": true
      },
      {
        "user_id": "u2",
        "username": "Bob",
        "is_active": true
      }
    ]
  }
  ```

- `GET /team/get?team_name=payments` - Получить команду с участниками

### Пользователи

- `POST /users/setIsActive` - Установить флаг активности пользователя
  ```json
  {
    "user_id": "u2",
    "is_active": false
  }
  ```

- `GET /users/getReview?user_id=u2` - Получить PR'ы, где пользователь назначен ревьювером

### Pull Request'ы

- `POST /pullRequest/create` - Создать PR (автоматически назначаются до 2 ревьюверов из команды автора)
  ```json
  {
    "pull_request_id": "pr-1001",
    "pull_request_name": "Add search",
    "author_id": "u1"
  }
  ```

- `POST /pullRequest/reassign` - Переназначить ревьювера
  ```json
  {
    "pull_request_id": "pr-1001",
    "old_user_id": "u2"
  }
  ```

- `POST /pullRequest/merge` - Выполнить merge PR (идемпотентная операция)
  ```json
  {
    "pull_request_id": "pr-1001"
  }
  ```

### Health Check

- `GET /health` - Проверка работоспособности сервиса

## Структура проекта

```
.
├── main.go              # Точка входа приложения
├── models/              # Модели данных
├── database/            # Работа с БД и миграции
├── repository/          # Слой доступа к данным
├── service/             # Бизнес-логика
├── handlers/            # HTTP обработчики
├── docker-compose.yml   # Конфигурация Docker Compose
├── Dockerfile           # Образ приложения
├── Makefile            # Команды для сборки и запуска
├── openapi.yaml        # OpenAPI спецификация API
└── README.md           # Документация
```

## Переменные окружения

- `DATABASE_URL` - Строка подключения к PostgreSQL (по умолчанию: `host=localhost user=postgres password=postgres dbname=avito sslmode=disable`)
- `PORT` - Порт для HTTP сервера (по умолчанию: `8080`)

