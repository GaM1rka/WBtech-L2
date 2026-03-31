# Calendar HTTP Server (task18)

Простой HTTP-сервер календаря событий.

## Структура

- `cmd/server/main.go` - запуск сервера, graceful shutdown
- `config/config.go` - загрузка конфигурации (`.env` и переменные окружения)
- `internal/repository` - in-memory хранилище (map)
- `internal/service` - бизнес-логика (CRUD, фильтры)
- `internal/handler` - HTTP-обработчики и middleware

## Запуск

1. `cd task18`
2. `make test` - запуск unit-тестов
3. `make run` - запуск сервера

Порт можно задать в `.env`:

- `CALENDAR_PORT=8080`

или переменной окружения:

- `CALENDAR_PORT=9090 make run`

или флагом:

- `go run ./cmd/server -port=9091`

## API

- `POST /create_event`
- `POST /update_event`
- `POST /delete_event`
- `GET /events_for_day?user_id=1&date=2024-01-01`
- `GET /events_for_week?user_id=1&date=2024-01-01`
- `GET /events_for_month?user_id=1&date=2024-01-01`

### Поддерживаемые форматы запросов (POST)

- `application/json`
- `application/x-www-form-urlencoded`

Поля:

- `user_id` (int, обязательное)
- `date` (`YYYY-MM-DD`, обязательное)
- `event` (text, обязательное)
- `id` (для update/delete)

## Примеры

### create_event

```bash
curl -X POST http://localhost:8080/create_event \
  -H 'Content-Type: application/json' \
  -d '{"user_id":1,"date":"2024-03-31","event":"test event"}'
```

### update_event

```bash
curl -X POST http://localhost:8080/update_event \
  -H 'Content-Type: application/x-www-form-urlencoded' \
  -d 'id=1&user_id=1&date=2024-03-31&event=updated'
```

### delete_event

```bash
curl -X POST http://localhost:8080/delete_event \
  -H 'Content-Type: application/x-www-form-urlencoded' \
  -d 'id=1&user_id=1'
```

### events_for_day

```bash
curl 'http://localhost:8080/events_for_day?user_id=1&date=2024-03-31'
```

### events_for_week

```bash
curl 'http://localhost:8080/events_for_week?user_id=1&date=2024-03-31'
```

### events_for_month

```bash
curl 'http://localhost:8080/events_for_month?user_id=1&date=2024-03-31'
```

## Docker

`make docker-run`
