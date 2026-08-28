# Delivery Service — MVP Сервиса Доставки Еды

Бэкенд-сервис для платформы «Delivery Service», реализующий интеграцию с партнерскими заведениями (рестораны и магазины) и обработку заказов пользователей.

---

## 🚀 Быстрый старт (Запуск)

Проект полностью упакован в Docker Compose и поднимается одной командой:

```bash
docker compose up --build -d

---



C4Container
    title C4 Container Diagram - Delivery Service

    Person(user, "Пользователь (Web)", "Использует веб-интерфейс Delivery Service")
    System_Boundary(c1, "Delivery Service Platform") {
        Container(api, "API Service", "Go, Chi, pgx", "Обрабатывает запросы клиентов и ресторанов, управляет заказами")
        ContainerDb(db, "PostgreSQL", "Postgres 16", "Хранит пользователей, заведения, меню, заказы и позиции")
    }
    System_Ext(mockRest, "Mock Restaurant / Shop", "Go / HTTP", "Принимает вебхуки о заказах и автоматически меняет статусы")

    Rel(user, api, "HTTP JSON", "API / Menus / Orders")
    Rel(api, db, "SQL / pgx", "Читает/Пишет данные")
    Rel(api, mockRest, "HTTP Webhook (POST/PATCH)", "Уведомляет о заказе и обновляет статус")
    Rel(mockRest, api, "HTTP PATCH", "Асинхронная смена статуса заказа")

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