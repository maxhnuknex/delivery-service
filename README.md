# Delivery Service — MVP Сервиса Доставки Еды

Бэкенд-сервис доставки еды, реализующий интеграцию с партнерскими заведениями (рестораны и магазины) и обработку заказов пользователей.

---

## 🚀 Быстрый старт (Запуск)

Проект полностью упакован в Docker Compose и поднимается одной командой:
```bash
docker compose up --build -d
```

После запуска будут доступны:
* Основной API: `http://localhost:8080/api/v1`
* Мок-ресторан (интеграционный сервис): `http://localhost:8081`
* База данных PostgreSQL: `localhost:5432`

## 🏛️ Архитектура (C4 Diagram - Level 2)

Ниже представлена схема контейнеров системы, демонстрирующая взаимодействие клиента, основного API, базы данных и внешнего сервиса заведения.

```mermaid
C4Container
    title C4 Container Diagram - Delivery Service

    Person(user, "Пользователь (Web)", "Использует веб-интерфейс сервиса")
    System_Boundary(c1, "Delivery Service") {
        Container(api, "API Service", "Go, Chi, pgx", "Обрабатывает запросы клиентов и ресторанов, управляет заказами")
        ContainerDb(db, "PostgreSQL", "Postgres 16", "Хранит пользователей, заведения, меню, заказы и позиции")
    }
    System_Ext(mockRest, "Mock Restaurant / Shop", "Go / HTTP", "Принимает вебхуки о заказах и автоматически меняет статусы")

    Rel(user, api, "HTTP JSON", "API / Menus / Orders")
    Rel(api, db, "SQL / pgx", "Читает/Пишет данные")
    Rel(api, mockRest, "HTTP Webhook (POST/PATCH)", "Уведомляет о заказе и обновляет статус")
    Rel(mockRest, api, "HTTP PATCH", "Асинхронная смена статуса заказа")
```

## 🛒 Customer Journey Map (CJM)

### 1. Пользовательский сценарий (Покупка еды/товаров)

```mermaid
sequenceDiagram
    autonumber
    actor User as Пользователь
    participant API as API Сервис
    participant DB as PostgreSQL
    participant Rest as Ресторан / Склад

    User->>API: GET /restaurants?type=RESTAURANT (Выбор заведения)
    API->>DB: Запрос активных заведений
    DB-->>API: Список заведений
    API-->>User: JSON список
    
    User->>API: GET /restaurants/{id}/menu (Просмотр меню)
    API->>DB: Выборка доступных блюд
    DB-->>API: Меню с ценами и остатками
    API-->>User: Список блюд (или [] если пусто)

    User->>API: POST /orders (Создание заказа)
    Note over API,DB: Проверка остатков (для SHOP) и статуса ресторана
    API->>DB: Сохранение заказа и позиций
    DB-->>API: Заказ создан (ID)
    API->>Rest: HTTP Webhook (Уведомление о заказе)
    API-->>User: 201 Created (Детали заказа)
```

### 2. Сценарий со стороны заведения

```mermaid
sequenceDiagram
    autonumber
    participant Rest as Сервис Заведения (Мок)
    participant API as API Сервис
    participant DB as PostgreSQL

    Note over Rest,API: При старте мок-сервис регистрируется в системе
    Rest->>API: POST /restaurants (Регистрация + Webhook URL)
    API->>DB: Сохранение заведения
    
    API->>Rest: POST /webhook (Отправка нового заказа)
    Rest->>Rest: Обработка заказа (Симуляция готовки)
    
    loop Каждые 5 секунд
        Rest->>API: PATCH /orders/{id}/status (COOKING -> DELIVERING -> COMPLETED)
        API->>DB: Обновление статуса заказа
        API-->>Rest: 204 No Content
    end
```

## 🗄️ Схема Базы данных и Миграции

Структура базы данных спроектирована с учетом масштабирования после завершения MVP. Инициализация и миграции выполняются автоматически через утилиту `Goose` при старте контейнеров.

* `users` — список пользователей платформы.
* `restaurants` — заведения и магазины (поля: `type` (RESTAURANT/SHOP), `is_active`, `webhook_url`).
* `menu_items` — меню заведений (поля: `price`, `stock_quantity` для контроля остатков магазинов, `is_available`).
* `orders` — заказы пользователей (поля: `status`, `delivery_address`, `total_price`).
* `order_items` — состав заказа (связь многие-к-многим между заказом и меню).

## 📝 Дополнительная документация и Качество кода

* **OpenAPI Спецификация:** Полное описание эндпоинтов и кодов ошибок находится в файле `openapi.yaml` в корне репозитория.
* **Статический анализ (Линтер):** Проект проверяется с помощью `golangci-lint`. Конфигурация находится в файле `.golangci.yml`.

## Тестирование

В проекте реализовано три уровня тестирования: юнит-тесты для репозиториев и сервисов, а также сквозные E2E-тесты.

### 1. Юнит-тесты (Репозитории и базы данных)
Для запуска тестов репозиториев используется изолированное тестовое окружение. Поднимите только тестовую базу данных:
```bash
docker compose -f docker-compose.e2e.yaml up db_e2e migrator_e2e -d
```

Запустите тесты:

```bash
go test -v ./internal/repository/postgres/...
```

### 2. Юнит-тесты (Сервисный слой)

Тестирование бизнес-логики с помощью моков (не требует запущенной базы данных):

```bash
go test -v ./internal/service/...
```

### 3. E2E (End-to-End) тесты

E2E тесты проверяют работу всего приложения целиком (HTTP API, базу данных, миграции и интеграцию со слоями).

1. Поднимите полное E2E окружение (API, базу данных, мигратор и мок-ресторан):

```bash
docker compose -f docker-compose.e2e.yaml --env-file .env.e2e up --build -d
```

2. Запустите E2E тесты:

```bash
go test -v ./e2e/...
```

## 💡 Допущения и упрощения (MVP)

* **Аутентификация:** Согласно ТЗ, авторизация пользователей и заведений в рамках MVP не реализовывалась.
* **Управление транзакциями:** Списание остатков для магазинов (SHOP) и создание заказа объединены на уровне сервисного слоя для обеспечения атомарности.
