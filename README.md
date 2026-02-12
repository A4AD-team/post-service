# post-service

Управление постами форума.

[![Go](https://img.shields.io/badge/Go-1.23+-00ADD8?logo=go)](https://go.dev/)

## Функции

- Создание / редактирование / удаление постов
- Просмотр списка постов (new/hot/top)
- Счётчики: views, likes, comments
- Инкремент счётчиков через события

## Технологии

- Go 1.23+
- PostgreSQL + pgx
- Redis (кэш популярных постов)
- Kafka / Redis Pub/Sub (события)

## Структура

```
post-service/
├── cmd/
│   └── server/
├── internal/
│   ├── handler/
│   ├── repository/
│   └── event/
├── migrations/
└── docker-compose.yml
```

## Эндпоинты

- POST   `/api/v1/posts`
- GET    `/api/v1/posts/:id`
- GET    `/api/v1/posts` (фильтры, пагинация)
- PATCH  `/api/v1/posts/:id`
- DELETE `/api/v1/posts/:id`
- POST   `/api/v1/posts/:id/like`
