# Framework2 – микросервисный API (Go + Gin)

Система учёта заказов/дефектов с разделением на микросервисы и API-gateway.

- `service_users` – сервис пользователей и аутентификации
- `service_orders` – сервис заказов
- `api_gateway` – единая точка входа (JWT, CORS, rate limit, X-Request-ID)
- `docs/` – OpenAPI спецификация и Postman-коллекция

---

## Архитектура

**API Gateway (`api_gateway`, порт 8080)**

- приём всех запросов от клиентов
- проксирование:
  - `/v1/users/**` → `service_users`
  - `/v1/orders/**` → `service_orders`
- проверка JWT (кроме регистрации и логина)
- CORS, rate limiting
- генерация и прокидывание заголовка `X-Request-ID`

**Сервис пользователей (`service_users`, порт 8081)**

- `POST /v1/users/register` – регистрация с базовой ролью
- `POST /v1/users/login` – логин, выдача JWT
- `GET /v1/users/me` – профиль текущего пользователя
- `PATCH /v1/users/me` – обновление имени
- `GET /v1/users` – список пользователей (только admin, с фильтрами)
- хранение данных в SQLite

**Сервис заказов (`service_orders`, порт 8082)**

- `POST /v1/orders` – создание заказа
- `GET /v1/orders` – список заказов текущего пользователя (пагинация, сортировка)
- `GET /v1/orders/{id}` – получение заказа по id
- `PATCH /v1/orders/{id}/status` – изменение статуса
- `POST /v1/orders/{id}/cancel` – отмена
- `DELETE /v1/orders/{id}` – удаление по правилам
- проверки прав по ролям (engineer, manager, director, customer, admin)
- хранение данных в SQLite
- доменные события в логах (`order.created`, `order.status_updated`)

---

## Переменные окружения

Общее:

```env
APP_ENV=dev        # dev / test / prod
JWT_SECRET=dev-secret-change-me
USERS_SERVICE_URL=http://localhost:8081
ORDERS_SERVICE_URL=http://localhost:8082



## Запуск без Docker

```bash
# service_users
cd service_users
go run .

# service_orders
cd service_orders
go run .

# api_gateway
cd api_gateway
go run .



## Запуск через Docker

В корне репозитория:
```bash
docker compose up --build
```

Остановка:
```bash
docker compose down

## Документация API

OpenAPI спецификация: docs/openapi.yaml
Можно открыть в Swagger Editor.