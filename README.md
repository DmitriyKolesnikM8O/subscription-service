# Subscription Service

REST-сервис для управления подписками пользователей с возможностью агрегации данных и расчета суммарной стоимости.

## Функциональность

- CRUD-операции для управления подписками
- Расчет суммарной стоимости подписок за период
- Фильтрация по пользователю и сервису
- Swagger-документация API
- Миграции базы данных

## Технологии

- Go 1.24
- PostgreSQL 14
- Echo framework
- Swaggo для документации
- Docker и Docker Compose

## Запуск сервиса

### Требования

- Docker 20.10+
- Docker Compose 2.0+

### Инструкция по запуску

1. Склонируйте репозиторий:
```bash
git clone <repository-url>
cd subscription-service
```

2. Создайте файл `.env` в корне проекта (для примера уже представлен):
```bash
cp .env.example .env
```

3. Запустите сервис:
```bash
docker-compose up -d --build
```

Сервис будет доступен по адресу: `http://localhost:8080`

## API Документация

Swagger UI доступен по адресу:  
`http://localhost:8080/swagger/index.html`

## Структура проекта

```
.
├── cmd/                  # Основной пакет приложения
├── config/               # Конфигурационные файлы
├── docs/                 # Swagger документация
├── internal/             # Внутренние пакеты
│   ├── controller/       # HTTP контроллеры
│   ├── entity/           # Сущности БД
│   ├── repo/             # Репозитории для работы с БД
│   └── service/          # Бизнес-логика
├── migrations/           # SQL-миграции
├── .env.example          # Пример конфигурации
├── docker-compose.yml    # Конфигурация Docker
└── Dockerfile            # Конфигурация образа
```

## Примеры запросов

### Создание подписки
```bash
curl -X POST "http://localhost:8080/api/v1/subscriptions" \
  -H "Content-Type: application/json" \
  -d '{
    "service_name": "Yandex Plus",
    "price": 400,
    "user_id": "6060ifee-2bf1-4721-ae6f-7636e79a0cba",
    "start_date": "07-2025"
  }'
```

### Получение списка подписок пользователя
```bash
curl "http://localhost:8080/api/v1/subscriptions?user_id=6060ifee-2bf1-4721-ae6f-7636e79a0cba"
```

### Расчет стоимости подписок
```bash
curl "http://localhost:8080/api/v1/subscriptions/total-cost?\
user_id=6060ifee-2bf1-4721-ae6f-7636e79a0cba&\
start_date=07-2025&end_date=09-2025"
```

## Переменные окружения

| Переменная       | Описание                     | Пример значения        |
|------------------|------------------------------|------------------------|
| DB_HOST          | Хост PostgreSQL              | postgres               |
| DB_PORT          | Порт PostgreSQL              | 5432                   |
| DB_USER          | Пользователь БД              | postgres               |
| DB_PASSWORD      | Пароль БД                    | postgres               |
| DB_NAME          | Имя БД                       | subscriptions          |
| HTTP_PORT        | Порт HTTP-сервера            | 8080                   |

## Миграции

Миграции базы данных автоматически применяются при старте контейнера PostgreSQL.  
Файлы миграций должны находиться в директории `./migrations`.

## Логирование

Логи приложения доступны:
- В контейнере: `docker-compose logs -f app`
- На хосте: в директории `./logs` (смонтирована в контейнер)