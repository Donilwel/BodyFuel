# BodyFuel — Backend

REST API бэкенд фитнес-сервиса **BodyFuel**. Написан на Go, реализует управление тренировками, дневником питания, весом, персонализированными рекомендациями и системой push/email/SMS уведомлений.

---

## Содержание

- [Быстрый старт](#быстрый-старт)
- [Запуск через Docker Compose](#запуск-через-docker-compose)
- [Ручной запуск](#ручной-запуск)
- [Конфигурация](#конфигурация)
- [Архитектура](#архитектура)
- [Сервисы](#сервисы)
- [База данных](#база-данных)
- [API Reference](#api-reference)
- [Бизнес-логика](#бизнес-логика)
- [Уведомления](#уведомления)
- [AI-интеграция](#ai-интеграция)
- [Подключение внешних сервисов](#подключение-внешних-сервисов)
- [Разработка](#разработка)

---

## Быстрый старт

### Требования

| Инструмент | Версия |
|-----------|--------|
| Go | ≥ 1.24 |
| PostgreSQL | ≥ 14 |
| MinIO | любая |
| Docker + Compose | для контейнерного запуска |

### Запуск через Docker Compose

Самый простой способ — поднять всё окружение одной командой:

```bash
docker compose up --build
```

Compose запустит три сервиса:
- **bodyfuel-app** — Go-приложение на порту `8080`
- **bodyfuel-postgres** — PostgreSQL на порту `5432`
- **bodyfuel-minio** — MinIO на порту `9000` (консоль: `9001`)

Миграции применяются автоматически при старте PostgreSQL через `docker-entrypoint-initdb.d`.

Swagger UI доступен сразу после старта:
```
http://localhost:8080/swagger/index.html
```

---

## Ручной запуск

**1. Клонировать репозиторий и перейти в папку бэкенда:**
```bash
cd backend
```

**2. Создать конфигурационный файл:**
```bash
cp config/config.yaml.template config/config.yaml
# Отредактировать config/config.yaml под своё окружение
```

**3. Применить миграции:**
```bash
go run github.com/pressly/goose/v3/cmd/goose@latest \
  postgres "host=localhost port=5432 user=danila password=postgres dbname=backend_db sslmode=disable" \
  up -dir migrations
```

**4. Запустить приложение:**
```bash
go run ./cmd/main.go -config ./config/config.yaml
```

---

## Конфигурация

Полный шаблон — `config/config.yaml.template`. Все параметры можно переопределить переменными окружения.

### Секция `app`

```yaml
app:
  http_server:
    host: "0.0.0.0"          # IP для прослушивания
    port: 8080                 # HTTP-порт API
    api_host: "localhost:8080" # Хост для Swagger (отображается в UI)
    metric_port: 8081          # Порт для метрик Prometheus
    tls: false                 # Включить HTTPS
    cert_path: ""              # Путь к TLS-сертификату
    key_path: ""               # Путь к TLS-ключу
  graceful_timeout: "5s"       # Таймаут graceful shutdown
  tasks_tracking_duration: "13s" # Интервал опроса очереди задач
  workouts_config:
    workout_pull_user_interval: "60s" # Интервал автогенерации тренировок
    limit_generate_workouts: 3        # Лимит авто-тренировок в день
```

### Секция `postgres`

```yaml
postgres:
  host: "localhost"
  port: 5432
  database: "backend_db"
  user: "danila"
  password: "postgres"
```

Переменные окружения: `POSTGRES_HOST`, `POSTGRES_PORT`, `POSTGRES_DATABASE`, `POSTGRES_USER`, `POSTGRES_PASSWORD`.

### Секция `minio`

```yaml
minio:
  endpoint: "http://localhost:9000"
  access_key: "minioadmin"
  secret_key: "minioadmin"
  bucket: "avatars"
  region: "us-east-1"
  public_url: "http://localhost:9000/avatars" # Публичный URL для аватаров
  presign_ttl: "5m"                           # TTL presigned URL
```

### Секция `sendgrid` (email)

```yaml
sendgrid:
  api_key: "SG.xxx"
  from_email: "noreply@bodyfuel.app"
  from_name: "BodyFuel"
```

Переменная окружения: `SENDGRID_API_KEY`.

### Секция `twilio` (SMS)

```yaml
twilio:
  account_sid: "ACxxx"
  auth_token: "xxx"
  from_phone: "+79001234567"
```

### Секция `apns` (iOS push)

```yaml
apns:
  key_path: "./keys/AuthKey.p8"  # Путь к .p8 ключу
  key_id: "XXXXXXXXXX"
  team_id: "XXXXXXXXXX"
  bundle_id: "com.bodyfuel.app"
  sandbox: true                  # false для продакшена
```

### Секция `openai` (AI-функции)

```yaml
openai:
  api_key: "sk-xxx"
```

Переменная окружения: `OPENAI_API_KEY`.

---

## Архитектура

Приложение построено по принципам **Clean Architecture + DDD**. Каждый слой взаимодействует с соседним только через интерфейсы.

```
cmd/
└── main.go                   # Точка входа, swagger-аннотации

internal/
├── app/                      # Сборка приложения (wire-up зависимостей)
├── config/                   # Конфигурационные структуры
├── domain/
│   └── entities/             # Доменные сущности и бизнес-правила
├── dto/                      # Data Transfer Objects (фильтры, спецификации)
├── errors/                   # Доменные ошибки
├── handlers/
│   └── v1/                   # HTTP-хендлеры (Gin), swagger-комментарии
│       └── models/           # Request/Response модели
├── infrastructure/
│   └── repositories/
│       ├── postgres/         # Реализация репозиториев (sqlx + squirrel)
│       │   ├── builders/     # SQL query builders
│       │   └── models/       # DB row-модели
│       └── minio/            # Клиент MinIO (S3)
└── service/
    ├── auth/                 # Аутентификация, токены, верификация
    ├── avatar/               # Presigned URL для аватаров
    ├── crud/                 # CRUD для всех доменных объектов
    ├── executor/             # Фоновый воркер задач
    ├── nutricion/            # Дневник питания, AI-анализ фото
    ├── recomendation/        # Персональные AI-рекомендации
    └── workouts/             # Генерация тренировок

pkg/
├── JWT/                      # Генерация и валидация JWT
├── ai/                       # Клиент OpenAI (Vision + Chat)
├── logging/                  # Структурированное логирование (zerolog)
└── notifications/
    ├── apns/                 # iOS push (APNs HTTP/2)
    ├── sendgrid/             # Email (SendGrid)
    └── twilio/               # SMS (Twilio)
```

---

## Сервисы

| Сервис | Пакет | Описание |
|--------|-------|----------|
| **Auth** | `service/auth` | Регистрация, вход, refresh-токен, верификация email/телефона, сброс пароля |
| **CRUD** | `service/crud` | CRUD для профиля, параметров, веса, упражнений, тренировок, устройств, калорий |
| **Workouts** | `service/workouts` | Генерация персональных тренировок, фоновая автогенерация по расписанию |
| **Executor** | `service/executor` | Фоновый воркер: опрашивает таблицу `tasks` и отправляет email / SMS / push |
| **Nutrition** | `service/nutricion` | Дневник питания, анализ фото через GPT-4o Vision, дневник, отчёты |
| **Recommendations** | `service/recomendation` | Генерация персональных рекомендаций через GPT-4o на основе профиля |
| **Avatar** | `service/avatar` | Presigned PUT URL для загрузки аватара напрямую в MinIO |

---

## База данных

Все таблицы находятся в схеме `bodyfuel`. Единственный файл миграций: `migrations/00001_init_schema.sql` (goose).

### `user_info` — аккаунты пользователей

| Колонка | Тип | Описание |
|---------|-----|----------|
| `id` | UUID PK | Идентификатор |
| `username` | TEXT UNIQUE | Никнейм |
| `name` | TEXT | Имя |
| `surname` | TEXT | Фамилия |
| `password` | TEXT | bcrypt-хэш пароля |
| `email` | TEXT UNIQUE | Email |
| `phone` | TEXT | Номер телефона |
| `created_at` | TIMESTAMPTZ | Дата регистрации |

### `user_params` — физические параметры и цели

| Колонка | Тип | Описание |
|---------|-----|----------|
| `id` | UUID PK | Идентификатор |
| `id_user` | UUID FK | → `user_info.id` |
| `height` | INT | Рост (см) |
| `photo` | TEXT | Ключ объекта аватара в MinIO |
| `wants` | ENUM | Цель: `lose_weight`, `build_muscle`, `stay_fit` |
| `lifestyle` | ENUM | Активность: `not_active`, `active`, `sportive` |
| `target_workouts_weeks` | INT | Тренировок в неделю (цель) |
| `target_calories_daily` | INT | Норма калорий в день |
| `target_weight` | FLOAT | Целевой вес (кг) |

### `user_weight` — история веса

| Колонка | Тип | Описание |
|---------|-----|----------|
| `id` | UUID PK | Идентификатор |
| `id_user` | UUID FK | → `user_info.id` |
| `weight` | FLOAT | Вес (кг) |
| `date` | TIMESTAMPTZ | Дата измерения |

### `exercise` — справочник упражнений

| Колонка | Тип | Описание |
|---------|-----|----------|
| `id` | UUID PK | Идентификатор |
| `name` | VARCHAR(100) | Название |
| `description` | TEXT | Описание техники |
| `level_preparation` | ENUM | `beginner`, `medium`, `sportsman` |
| `type_exercise` | ENUM | `cardio`, `upper_body`, `lower_body`, `full_body`, `flexibility` |
| `place_exercise` | ENUM | `home`, `gym`, `street` |
| `base_count_reps` | INT | Базовое число повторений |
| `steps` | INT | Количество подходов |
| `avg_calories_per` | DECIMAL | Калорий за одно повторение |
| `base_relax_time` | INT | Отдых между подходами (сек) |
| `link_gif` | TEXT | Ссылка на анимацию |

### `workout` — тренировки

| Колонка | Тип | Описание |
|---------|-----|----------|
| `id` | UUID PK | Идентификатор |
| `user_id` | UUID FK | → `user_info.id` |
| `level` | ENUM | `workout_light`, `workout_middle`, `workout_hard` |
| `status` | ENUM | `workout_created`, `workout_in_active`, `workout_done`, `workout_failed` |
| `prediction_calories` | INT | Прогноз калорий |
| `total_calories` | INT | Фактически сожжено |
| `duration` | INT | Длительность (нс) |
| `created_at` | TIMESTAMPTZ | Дата создания |
| `updated_at` | TIMESTAMPTZ | Последнее обновление |

### `workouts_exercise` — упражнения внутри тренировки

| Колонка | Тип | Описание |
|---------|-----|----------|
| `workout_id` | UUID FK | → `workout.id` |
| `exercise_id` | UUID FK | → `exercise.id` |
| `modify_reps` | INT | Скорректированные повторения |
| `modify_relax_time` | INT | Скорректированное время отдыха (сек) |
| `calories` | INT | Калории за это упражнение |
| `status` | ENUM | `pending`, `in_progress`, `completed`, `skipped` |
| `created_at` | TIMESTAMPTZ | Дата создания |
| `updated_at` | TIMESTAMPTZ | Последнее обновление |

### `tasks` — очередь фоновых задач

| Колонка | Тип | Описание |
|---------|-----|----------|
| `task_id` | UUID PK | Идентификатор |
| `task_type_nm` | TEXT | Тип: `send_code_email_task`, `send_code_phone_task`, `send_notification_email_task`, `send_notification_phone_task`, `send_push_notification_task` |
| `task_state` | TEXT | `running`, `failed` |
| `max_attempts` | INT | Максимум попыток |
| `attempts` | INT | Текущее число попыток |
| `retry_at` | TIMESTAMPTZ | Время следующей попытки |
| `attribute` | JSONB | Полезная нагрузка (email / phone / device_token / subject / body / code) |
| `created_at` | TIMESTAMPTZ | Создана |
| `updated_at` | TIMESTAMPTZ | Обновлена |

### `user_devices` — устройства для push-уведомлений

| Колонка | Тип | Описание |
|---------|-----|----------|
| `id` | UUID PK | Идентификатор |
| `user_id` | UUID FK | → `user_info.id` |
| `device_token` | TEXT UNIQUE | APNs device token |
| `platform` | TEXT | `ios`, `android` |
| `created_at` | TIMESTAMPTZ | Дата регистрации |
| `updated_at` | TIMESTAMPTZ | Последнее обновление |

### `user_calories` — трекинг калорий

| Колонка | Тип | Описание |
|---------|-----|----------|
| `id` | UUID PK | Идентификатор |
| `user_id` | UUID FK | → `user_info.id` |
| `calories` | INT | Количество калорий |
| `description` | TEXT | Комментарий |
| `date` | TIMESTAMPTZ | Дата записи |
| `created_at` | TIMESTAMPTZ | Создана |
| `updated_at` | TIMESTAMPTZ | Обновлена |

### `user_refresh_tokens` — refresh-токены

| Колонка | Тип | Описание |
|---------|-----|----------|
| `id` | UUID PK | Идентификатор |
| `user_id` | UUID FK | → `user_info.id` |
| `token_hash` | TEXT UNIQUE | SHA-256 хэш токена |
| `expires_at` | TIMESTAMPTZ | Срок действия (30 дней) |
| `created_at` | TIMESTAMPTZ | Создан |

### `user_verification_codes` — коды верификации

| Колонка | Тип | Описание |
|---------|-----|----------|
| `id` | UUID PK | Идентификатор |
| `user_id` | UUID FK | → `user_info.id` |
| `code_hash` | TEXT | SHA-256 хэш 6-значного кода |
| `code_type` | TEXT | `email`, `phone`, `recover` |
| `expires_at` | TIMESTAMPTZ | Срок действия (10 минут) |
| `used_at` | TIMESTAMPTZ | Время использования (NULL = не использован) |
| `created_at` | TIMESTAMPTZ | Создан |

### `user_food` — дневник питания

| Колонка | Тип | Описание |
|---------|-----|----------|
| `id` | UUID PK | Идентификатор |
| `user_id` | UUID FK | → `user_info.id` |
| `description` | TEXT | Название / описание блюда |
| `calories` | INT | Калории |
| `protein` | NUMERIC(6,2) | Белки (г) |
| `carbs` | NUMERIC(6,2) | Углеводы (г) |
| `fat` | NUMERIC(6,2) | Жиры (г) |
| `meal_type` | TEXT | `breakfast`, `lunch`, `dinner`, `snack` |
| `photo_url` | TEXT | URL изображения |
| `date` | DATE | Дата приёма пищи |
| `created_at` | TIMESTAMPTZ | Создана |
| `updated_at` | TIMESTAMPTZ | Обновлена |

### `user_recommendation` — рекомендации

| Колонка | Тип | Описание |
|---------|-----|----------|
| `id` | UUID PK | Идентификатор |
| `user_id` | UUID FK | → `user_info.id` |
| `type` | TEXT | `workout`, `nutrition`, `rest`, `general` |
| `description` | TEXT | Текст рекомендации |
| `priority` | SMALLINT | Приоритет: 1 — высокий, 2 — средний, 3 — низкий |
| `is_read` | BOOLEAN | Прочитана ли рекомендация |
| `generated_at` | TIMESTAMPTZ | Время генерации |
| `created_at` | TIMESTAMPTZ | Создана |

---

## API Reference

Базовый путь: `/api/v1`

Swagger UI: `http://localhost:8080/swagger/index.html`

Все защищённые эндпоинты требуют заголовок:
```
Authorization: Bearer <access_token>
```

---

### Auth

| Метод | Путь | Авторизация | Описание |
|-------|------|:-----------:|----------|
| `POST` | `/auth/register` | — | Регистрация нового пользователя |
| `POST` | `/auth/login` | — | Вход, возвращает `access_token` + `refresh_token` |
| `POST` | `/auth/refresh` | — | Обмен refresh-токена на новую пару токенов (ротация) |
| `POST` | `/auth/recover` | — | Запрос кода для сброса пароля (код на email) |
| `POST` | `/auth/reset-password` | — | Сброс пароля по email + коду |
| `POST` | `/auth/send-verification` | ✓ | Отправка кода подтверждения на email или телефон |
| `POST` | `/auth/verify-email` | ✓ | Подтверждение email по 6-значному коду |
| `POST` | `/auth/verify-phone` | ✓ | Подтверждение телефона по 6-значному коду |

**Регистрация** `POST /auth/register`
```json
{
  "username": "john_doe",
  "name": "John",
  "surname": "Doe",
  "email": "john@example.com",
  "phone": "+79001234567",
  "password": "secret123"
}
```

**Вход** `POST /auth/login`
```json
{ "username": "john_doe", "password": "secret123" }
```
Ответ:
```json
{ "access_token": "eyJ...", "refresh_token": "a3f9..." }
```

**Обновление токенов** `POST /auth/refresh`
```json
{ "refresh_token": "a3f9..." }
```
Ответ — новая пара `access_token` + `refresh_token`. Старый refresh-токен сразу инвалидируется.

**Запрос кода верификации** `POST /auth/send-verification`
```json
{ "code_type": "email" }
```
Доступные типы: `email`, `phone`. Код действует 10 минут.

**Подтверждение email** `POST /auth/verify-email`
```json
{ "code": "123456", "code_type": "email" }
```

**Восстановление пароля** — двухшаговый процесс:

1. `POST /auth/recover` → `{ "email": "john@example.com" }` → код приходит на email
2. `POST /auth/reset-password` → `{ "email": "john@example.com", "code": "123456", "new_password": "newsecret" }`

---

### User Info

| Метод | Путь | Авторизация | Описание |
|-------|------|:-----------:|----------|
| `GET` | `/user/info` | ✓ | Профиль текущего пользователя |
| `PATCH` | `/user/info` | ✓ | Обновление имени, фамилии, email, телефона |
| `DELETE` | `/user/info` | ✓ | Удаление аккаунта |

---

### User Params

| Метод | Путь | Авторизация | Описание |
|-------|------|:-----------:|----------|
| `POST` | `/user/params` | ✓ | Создание параметров (рост, цель, образ жизни) |
| `GET` | `/user/params` | ✓ | Получение параметров |
| `PATCH` | `/user/params` | ✓ | Обновление параметров |
| `DELETE` | `/user/params` | ✓ | Удаление параметров |

**Создание** `POST /user/params`
```json
{
  "height": 180,
  "wants": "lose_weight",
  "lifestyle": "active",
  "target_workouts_weeks": 4,
  "target_calories_daily": 2000,
  "target_weight": 75.0
}
```

Значения `wants`: `lose_weight`, `build_muscle`, `stay_fit`

Значения `lifestyle`: `not_active`, `active`, `sportive`

---

### User Weight

| Метод | Путь | Авторизация | Описание |
|-------|------|:-----------:|----------|
| `GET` | `/user/weight` | ✓ | Последнее измерение веса |
| `GET` | `/user/weight/history` | ✓ | Вся история изменений веса |
| `POST` | `/user/weight` | ✓ | Добавить новое измерение |
| `DELETE` | `/user/weight/:uuid` | ✓ | Удалить запись о весе |

**Добавление** `POST /user/weight`
```json
{ "weight": 82.5 }
```

---

### User Calories

| Метод | Путь | Авторизация | Описание |
|-------|------|:-----------:|----------|
| `GET` | `/user/calories/history` | ✓ | История записей (фильтр: `start_date`, `end_date` в RFC3339) |
| `POST` | `/user/calories` | ✓ | Новая запись |
| `PATCH` | `/user/calories/:uuid` | ✓ | Обновление записи |
| `DELETE` | `/user/calories/:uuid` | ✓ | Удаление записи |

**Создание** `POST /user/calories`
```json
{
  "calories": 450,
  "description": "Обед",
  "date": "2025-04-12T13:00:00Z"
}
```

---

### User Devices

| Метод | Путь | Авторизация | Описание |
|-------|------|:-----------:|----------|
| `POST` | `/user/devices` | ✓ | Зарегистрировать устройство (APNs device token) |
| `GET` | `/user/devices` | ✓ | Список зарегистрированных устройств |
| `DELETE` | `/user/devices/:uuid` | ✓ | Удалить устройство |

**Регистрация** `POST /user/devices`
```json
{
  "device_token": "abc123...",
  "platform": "ios"
}
```

---

### Exercises

| Метод | Путь | Авторизация | Описание |
|-------|------|:-----------:|----------|
| `GET` | `/exercises` | ✓ | Список упражнений (фильтрация по типу, уровню, месту) |
| `GET` | `/exercises/:uuid` | ✓ | Упражнение по ID |
| `POST` | `/exercises` | ✓ | Создать упражнение |
| `PATCH` | `/exercises/:uuid` | ✓ | Обновить упражнение |
| `DELETE` | `/exercises/:uuid` | ✓ | Удалить упражнение |

**Создание** `POST /exercises`
```json
{
  "name": "Приседания",
  "description": "Классические приседания со штангой",
  "level_preparation": "medium",
  "type_exercise": "lower_body",
  "place_exercise": "gym",
  "base_count_reps": 10,
  "steps": 3,
  "avg_calories_per": 0.5,
  "base_relax_time": 60,
  "link_gif": "https://example.com/squat.gif"
}
```

Значения `level_preparation`: `beginner`, `medium`, `sportsman`

Значения `type_exercise`: `cardio`, `upper_body`, `lower_body`, `full_body`, `flexibility`

Значения `place_exercise`: `home`, `gym`, `street`

---

### Workouts

| Метод | Путь | Авторизация | Описание |
|-------|------|:-----------:|----------|
| `POST` | `/workouts` | ✓ | Сгенерировать тренировку |
| `GET` | `/workouts/history` | ✓ | История тренировок |
| `GET` | `/workouts/:uuid` | ✓ | Детали тренировки с упражнениями |
| `PATCH` | `/workouts/:uuid` | ✓ | Обновить статус / длительность |
| `DELETE` | `/workouts/:uuid` | ✓ | Удалить тренировку |

**Генерация** `POST /workouts`
```json
{
  "place_exercise": "gym",
  "type_exercise": "upper_body",
  "level": "medium",
  "exercises_count": 6
}
```

Система подбирает упражнения из справочника с учётом параметров пользователя (`user_params`) и заданных фильтров. Тренировка сохраняется и возвращается со списком упражнений.

**Обновление** `PATCH /workouts/:uuid`
```json
{
  "status": "workout_done",
  "duration": "1h15m"
}
```

Статусы тренировки: `workout_created`, `workout_in_active`, `workout_done`, `workout_failed`

---

### Workout Exercises

| Метод | Путь | Авторизация | Описание |
|-------|------|:-----------:|----------|
| `GET` | `/workouts/:workoutId/exercises` | ✓ | Список упражнений тренировки |
| `POST` | `/workouts/:workoutId/exercises` | ✓ | Добавить упражнение в тренировку |
| `PATCH` | `/workouts/exercises/:uuid` | ✓ | Обновить статус / данные упражнения |
| `DELETE` | `/workouts/exercises/:uuid` | ✓ | Удалить упражнение из тренировки |

**Добавление** `POST /workouts/:workoutId/exercises`
```json
{
  "exercise_id": "550e8400-e29b-41d4-a716-446655440000",
  "modify_reps": 12,
  "modify_relax_time": 90
}
```

**Обновление** `PATCH /workouts/exercises/:uuid`
```json
{
  "status": "completed",
  "calories": 80
}
```

Статусы упражнения: `pending`, `in_progress`, `completed`, `skipped`

---

### Nutrition

| Метод | Путь | Авторизация | Описание |
|-------|------|:-----------:|----------|
| `POST` | `/nutrition/analyze` | ✓ | Анализ фото еды через GPT-4o Vision |
| `POST` | `/nutrition/entries` | ✓ | Добавить запись в дневник питания |
| `PATCH` | `/nutrition/entries/:uuid` | ✓ | Обновить запись |
| `DELETE` | `/nutrition/entries/:uuid` | ✓ | Удалить запись |
| `GET` | `/nutrition/diary` | ✓ | Дневник за день с суммарными макросами |
| `GET` | `/nutrition/report` | ✓ | Отчёт за период с усреднением |

**Анализ фото** `POST /nutrition/analyze`

Отправляет изображение в GPT-4o Vision и возвращает расчёт КБЖУ:
```json
{ "image_url": "https://example.com/food.jpg" }
```
Ответ:
```json
{
  "description": "Куриная грудка с рисом",
  "calories": 420,
  "protein": 38.5,
  "carbs": 42.0,
  "fat": 6.2
}
```

**Создание записи** `POST /nutrition/entries`
```json
{
  "description": "Куриная грудка с рисом",
  "calories": 420,
  "protein": 38.5,
  "carbs": 42.0,
  "fat": 6.2,
  "meal_type": "lunch",
  "photo_url": "https://example.com/food.jpg",
  "date": "2025-04-12T13:00:00Z"
}
```

Тип приёма пищи (`meal_type`): `breakfast`, `lunch`, `dinner`, `snack`

**Дневник** `GET /nutrition/diary?date=2025-04-12`

Возвращает все записи за день и суммарные значения КБЖУ:
```json
{
  "date": "2025-04-12",
  "entries": [...],
  "total_calories": 1850,
  "total_protein": 142.5,
  "total_carbs": 195.0,
  "total_fat": 48.3
}
```

**Отчёт** `GET /nutrition/report?from=2025-04-01&to=2025-04-12`
```json
{
  "from": "2025-04-01",
  "to": "2025-04-12",
  "days": 12,
  "total_calories": 22200,
  "avg_calories_per_day": 1850.0,
  ...
}
```

---

### Recommendations

| Метод | Путь | Авторизация | Описание |
|-------|------|:-----------:|----------|
| `GET` | `/recommendations` | ✓ | Список рекомендаций (пагинация: `page`, `limit`) |
| `POST` | `/recommendations/refresh` | ✓ | Сгенерировать новые рекомендации через GPT |
| `PATCH` | `/recommendations/:uuid/read` | ✓ | Отметить рекомендацию прочитанной |

**Список** `GET /recommendations?page=1&limit=10`
```json
[
  {
    "id": "550e8400-...",
    "type": "nutrition",
    "description": "Увеличьте потребление белка до 2г на кг веса",
    "priority": 1,
    "is_read": false,
    "generated_at": "2025-04-12T10:00:00Z"
  }
]
```

**Обновление рекомендаций** `POST /recommendations/refresh`

GPT-4o анализирует профиль пользователя (`user_params`: вес, рост, цель, образ жизни) и генерирует 3–5 персонализированных рекомендаций. Старые рекомендации заменяются новыми.

Типы рекомендаций: `workout`, `nutrition`, `rest`, `general`

Приоритеты: `1` — высокий, `2` — средний, `3` — низкий

---

### Tasks

| Метод | Путь | Авторизация | Описание |
|-------|------|:-----------:|----------|
| `GET` | `/tasks` | ✓ | Список задач |
| `GET` | `/tasks/:uuid` | ✓ | Задача по ID |
| `POST` | `/tasks/:uuid/restart` | ✓ | Перезапустить упавшую задачу |
| `DELETE` | `/tasks/:uuid` | ✓ | Удалить задачу |

---

### Avatars

| Метод | Путь | Авторизация | Описание |
|-------|------|:-----------:|----------|
| `POST` | `/avatars` | ✓ | Получить presigned PUT URL для загрузки аватара |

**Запрос** `POST /avatars`
```json
{ "content_type": "image/jpeg" }
```
Ответ:
```json
{
  "upload_url": "http://minio:9000/avatars/users/...",
  "object_key": "users/uuid/avatar.jpg",
  "public_url": "http://localhost:9000/avatars/users/uuid/avatar.jpg"
}
```

Клиент загружает файл напрямую в MinIO через `PUT upload_url`, бэкенд не участвует в передаче данных.

---

## Бизнес-логика

### Аутентифик��ция и токены

Система использует два типа токенов:

| Токен | TTL | Хранение |
|-------|-----|----------|
| **Access token** (JWT) | 24 часа | Только на клиенте |
| **Refresh token** (random hex 128 символов) | 30 дней | SHA-256 хэш в таблице `user_refresh_tokens` |

**Флоу токенов:**

1. `POST /auth/login` → клиент получает пару `access_token` + `refresh_token`
2. Все запросы к API: заголовок `Authorization: Bearer <access_token>`
3. При истечении access-токена: `POST /auth/refresh` с refresh-токеном → новая пара
4. Refresh-токен **ротируется** при каждом использовании (старый удаляется, выдаётся новый)
5. При сбросе пароля **все** refresh-токены пользователя уничтожаются

**Пароли:** хэшируются через bcrypt перед сохранением в БД. Оригинал нигде не хранится.

---

### Верификация email и телефона

**Поля в таблице `user_info`:**
- `email_verified_at TIMESTAMPTZ NULL` — заполняется при успешной верификации email
- `phone_verified_at TIMESTAMPTZ NULL` — заполняется при успешной верификации телефона

**Флоу верификации:**

1. Клиент вызывает `POST /auth/send-verification` с `code_type: "email"` или `"phone"`
2. Сервис генерирует случайный 6-значный код, сохраняет SHA-256 хэш в `user_verification_codes`, TTL — 10 минут
3. Создаётся задача в очереди (`send_code_email_task` / `send_code_phone_task`)
4. Executor отправляет код через SendGrid или Twilio
5. Клиент вызывает `POST /auth/verify-email` или `POST /auth/verify-phone` с кодом
6. Сервер проверяет хэш. При совпадении: помечает код как использованный, проставляет `email_verified_at` / `phone_verified_at`

**Правила кодов:**
- Код одноразовый (повторное использование → ошибка `ErrVerificationCodeAlreadyUsed`)
- Истёкший код → ошибка `ErrVerificationCodeExpired`
- Неверный код → ошибка `ErrInvalidVerificationCode`

**Верификация и уведомления** — ключевое ограничение:

| Тип задачи | Требует верификации |
|-----------|---------------------|
| `send_code_email_task` (сам код верификации) | Нет — всегда отправляется |
| `send_code_phone_task` (сам код верификации) | Нет — всегда отправляется |
| `send_notification_email_task` (общие уведомления) | **Да** — email должен быть верифицирован |
| `send_notification_phone_task` (SMS-уведомления) | **Да** — телефон должен быть верифицирован |
| `send_push_notification_task` | Нет |

Если канал не верифицирован — уведомление удаляется без отправки и без ретраев.

---

### Восстановление пароля

1. `POST /auth/recover` — принимает email, **всегда возвращает 200** (защита от перебора пользователей)
2. Если пользователь с таким email существует — создаётся код типа `recover` и задача `send_code_email_task`
3. `POST /auth/reset-password` — принимае�� email + код + новый пароль
4. При успехе: новый пароль сохраняется (bcrypt), все refresh-токены пользователя инвалидируются

---

### Дневник питания

**Записи о еде** (`user_food`) привязаны к дате. Типы приёмов пищи: `breakfast`, `lunch`, `dinner`, `snack`.

**Дневник** (`GET /nutrition/diary?date=YYYY-MM-DD`):
- Возвращает все записи за день + суммарные КБЖУ (calories, protein, carbs, fat)
- Дата по умолчанию — сегодня

**Отчёт** (`GET /nutrition/report?from=...&to=...`):
- Возвращает все записи за период + суммарные КБЖУ + среднесуточные калории
- `avg_calories_per_day = total_calories / кол-во_дней_в_периоде`

**Рекомендации рецептов** (`GET /nutrition/recipes?date=YYYY-MM-DD`):
- Автоматически читает дневник за указанную дату
- Передаёт суммарные КБЖУ в GPT-4o mini
- GPT предлагает 3–5 рецептов, которые дополнят рацион до баланса
- Каждый рецепт: название, описание с шагами приготовления, КБЖУ (protein/fat/carbs), время приготовления в минутах

---

### Персонализированные рекомендации

**Источник данных для AI:** `user_params` (вес, рост, цель `want`, образ жизни `lifestyle`). Если параметры не заполнены — используются дефолты (`general fitness`, `moderate`).

**`POST /recommendations/refresh`:**
1. Читает `user_params`
2. Отправляет профиль в GPT-4o mini
3. Удаляет все старые рекомендации пользователя
4. Сохраняет новые (3–5 штук) с типами: `workout`, `nutrition`, `rest`, `general` и приоритетами 1–3

**`PATCH /recommendations/:uuid/read`** — помечает рекомендацию прочитанной (`read_at`).

---

### Тренировки

**Генерация** (`POST /workouts`):
- Принимает: место (`home`/`gym`/`street`), тип упражнений, уровень (`beginner`/`medium`/`sportsman`), количество упражнений
- Фильтрует упражнения из БД по параметрам, создаёт `workout` + записи `workout_exercises`

**Порядок упражнений в тренировке (фазы):**

| Фаза | Тип упражнений | Назначение |
|------|----------------|------------|
| 0 | `flexibility` | Разминка — всегда первой |
| 1 | `upper_body`, `lower_body`, `full_body` | Основной силовой блок |
| 2 | `cardio` | Кардио — всегда последним |

Правило: кардио не может стоять между силовыми упражнениями; растяжка не может стоять в середине тренировки.

**Правило скипа упражнений:**

| Статус скипа | Поведение |
|---|---|
| Пропущено 1 раз | Упражнение предлагается ещё 1 раз |
| Пропущено 2+ раз за 7 дней | Упражнение **блокируется** на 7 дней с момента последнего скипа |
| Прошло 7 дней с последнего скипа | Блокировка снимается, упражнение доступно снова |

Система считает скипы из таблицы `workouts_exercise` (статус `skipped`), агрегируя по `exercise_id` для данного пользователя.

**Адаптация к питанию (нутрициологический контекст):**

Интенсивность тренировки автоматически корректируется относительно баланса калорий за сегодня:

| Баланс (съедено − цель) | Коэффициент интенсивности | Логика |
|---|---|---|
| Профицит > +300 ккал | ×1.2 | Пользователь переел — можно нагрузить сильнее |
| Нормальный (±300 ккал) | ×1.0 | Без изменений |
| Дефицит > −300 ккал, цель не «набор массы» | ×0.8 | Не хватает топлива — щадящая тренировка |
| Дефицит > −300 ккал, цель `build_muscle` | ×1.0 | При наборе не снижаем нагрузку |

**Предпочтительный тип упражнений по прогрессу веса:**

| Дельта веса (текущий − целевой) | Рекомендуемый тип |
|---|---|
| > +1 кг (нужно похудеть), цель не `build_muscle` | `cardio` |
| > +1 кг (нужно похудеть), цель `build_muscle` | `full_body` |
| < −1 кг (нужно набрать) | `upper_body` |
| Около цели (±1 кг) | Без изменений |

**Автогенерация** (background):
- Сервис `workouts` работает в фоне, каждые `workout_pull_user_interval` секунд
- Генерирует тренировки для пользователей, у которых нет активных тренировок
- Учитывает нутрициологический контекст, прогресс веса и историю скипов
- Лимит: `limit_generate_workouts` автотренировок в день на пользователя
- После генерации создаёт задачи уведомлений (email + SMS + push на все устройства пользователя)

---

## Уведомления

Executor-сервис работает в фоне и опрашивает таблицу `tasks` с интервалом `tasks_tracking_duration`. При обнаружении задачи со статусом `running` и наступившим `retry_at` выполняет отправку.

| Канал | Провайдер | Типы задач |
|-------|-----------|------------|
| Email | SendGrid | `send_code_email_task`, `send_notification_email_task` |
| SMS | Twilio | `send_code_phone_task`, `send_notification_phone_task` |
| Push (iOS) | APNs HTTP/2 | `send_push_notification_task` |

**Стратегия повторов при ошибке:**

| Тип задачи | Backoff |
|-----------|---------|
| Коды (email/phone) | Fibonacci backoff, база 20 сек |
| Уведомления (email/phone) | Exponential + jitter, база 10 сек |
| Push | Linear backoff, база 20 сек |

После превышения `max_attempts` задача переводится в `failed`. Перезапустить можно через `POST /tasks/:uuid/restart`.

**Когда создаются задачи автоматически:**
- `POST /auth/send-verification` → `send_code_email_task` или `send_code_phone_task`
- `POST /auth/recover` → `send_code_email_task` (если email существует)
- Автогенерация тренировки → `send_notification_email_task` + `send_notification_phone_task` + `send_push_notification_task` на каждое устройство

---

## AI-интеграция

Сервис использует **GPT-4o mini** (OpenAI) для трёх задач:

### Анализ фото еды (`POST /nutrition/analyze`)

Отправляет URL изображения в GPT-4o Vision. Промпт требует вернуть строгий JSON с полями `description`, `calories`, `protein`, `carbs`, `fat`. Значения — оценка на порцию, показанную на фото.

Требование: URL изображения должен быть публично доступен.

### Рекомендации рецептов (`GET /nutrition/recipes`)

Читает дневник за день, считает итоговые КБЖУ и просит GPT предложить блюда, которые сбалансируют рацион. Возвращает 3–5 рецептов с макросами и временем приготовления.

### Генерация персональных рекомендаций (`POST /recommendations/refresh`)

GPT-4o получает профиль пользователя (вес, рост, цель, активность, **целевой вес и прогресс к нему**) и генерирует 3–5 конкретных рекомендаций в формате JSON-массива. Каждая рекомендация содержит тип, описание и приоритет.

Система использует последний залогированный вес из `user_weight` для актуальности данных (не только параметры из `user_params`). В промпте явно указывается, сколько кг нужно сбросить/набрать до целевого веса — GPT включает в рекомендации советы по питанию или тренировкам исходя из этого прогресса.

Если `OPENAI_API_KEY` не задан, запросы к AI-эндпоинтам вернут ошибку `500`.

---

## Подключение внешних сервисов

### OpenAI (анализ фото и рекомендации)

**1. Зарегистрируйся на OpenAI:**
- Перейди на [platform.openai.com](https://platform.openai.com)
- Создай аккаунт или войди

**2. Создай API-ключ:**
- Открой [platform.openai.com/api-keys](https://platform.openai.com/api-keys)
- Нажми **Create new secret key**
- Скопируй ключ — он показывается только один раз

**3. Пополни баланс:**
- [platform.openai.com/settings/organization/billing](https://platform.openai.com/settings/organization/billing)
- Приложение использует **gpt-4o-mini** (один из самых дешёвых): ~$0.15 за 1M токенов входа

**4. Добавь в конфиг:**
```yaml
openai:
  api_key: "sk-proj-..."
```
Или через переменную окружения:
```bash
export OPENAI_API_KEY="sk-proj-..."
```

**Что использует ключ:**
- `POST /nutrition/analyze` — Vision API, анализирует URL фотографии еды
- `POST /recommendations/refresh` — Chat API, генерирует персональные рекомендации

Если ключ не задан — оба эндпоинта вернут `500`.

---

### SendGrid (email: верификация и восстановление пароля)

**1. Зарегистрируйся:**
- [sendgrid.com](https://sendgrid.com) → Sign Up (есть бесплатный tier: 100 писем/день)

**2. Подтверди домен-отправитель (или единственный email):**
- В SendGrid: **Settings → Sender Authentication**
- Для быстрого старта — **Single Sender Verification**: подтверди один конкретный email-адрес
- Для продакшена — **Domain Authentication**: добавь DNS-записи для своего домена

**3. Создай API-ключ:**
- **Settings → API Keys → Create API Key**
- Права: **Restricted Access → Mail Send → Full Access**
- Скопируй ключ (показывается один раз)

**4. Добавь в конфиг:**
```yaml
sendgrid:
  api_key: "SG.xxxxxxxxxxxxxxxxxx"
  from_email: "noreply@yourdomain.com"  # должен совпадать с подтверждённым отправителем
  from_name: "BodyFuel"
```
Или через переменные окружения:
```bash
export SENDGRID_API_KEY="SG.xxxxxxxxxxxxxxxxxx"
export SENDGRID_FROM_EMAIL="noreply@yourdomain.com"
export SENDGRID_FROM_NAME="BodyFuel"
```

**Когда отправляются письма:**
- `POST /auth/send-verification` с `code_type: "email"` — код подтверждения email
- `POST /auth/recover` — код восстановления пароля

Если ключ не задан — задачи будут создаваться в БД, но отправка упадёт с ошибкой и задача перейдёт в `failed` после `max_attempts` попыток.

---

### Twilio (SMS: верификация телефона)

**1. Зарегистрируйся:**
- [twilio.com/try-twilio](https://twilio.com/try-twilio) → бесплатный аккаунт с $15 триал-кредитом

**2. Получи учётные данные:**
- На главной странице консоли (**Console Dashboard**) сразу видны:
  - **Account SID** (начинается с `AC...`)
  - **Auth Token** (показать → скопировать)

**3. Получи номер для отправки SMS:**
- **Phone Numbers → Manage → Buy a number**
- В триале — один бесплатный номер
- Важно: в триале можно отправлять SMS только на **верифицированные** номера (добавить в **Verified Caller IDs**)

**4. Добавь в конфиг:**
```yaml
twilio:
  account_sid: "ACxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"
  auth_token: "xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"
  from_phone: "+1234567890"  # номер в формате E.164
```
Или через переменные окружения:
```bash
export TWILIO_ACCOUNT_SID="ACxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"
export TWILIO_AUTH_TOKEN="xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"
export TWILIO_FROM_PHONE="+1234567890"
```

**Когда отправляются SMS:**
- `POST /auth/send-verification` с `code_type: "phone"` — код подтверждения телефона

---

### APNs (iOS push-уведомления) — опционально

Без APNs приложение работает полностью. Push просто не отправляются.

**1. Нужен Apple Developer аккаунт** ([developer.apple.com](https://developer.apple.com), $99/год)

**2. Создай Authentication Key:**
- **Certificates, Identifiers & Profiles → Keys → (+)**
- Включи **Apple Push Notifications service (APNs)**
- Нажми **Continue** → **Register**
- Скачай файл `.p8` — он скачивается **один раз**
- Запомни **Key ID** (10 символов)

**3. Найди Team ID:**
- **Membership → Team ID** (10 символов)

**4. Bundle ID приложения:**
- **Identifiers** → выбери своё приложение → скопируй Bundle ID (например `com.yourcompany.bodyfuel`)

**5. Добавь в конфиг:**
```yaml
apns:
  key_path: "./keys/AuthKey_XXXXXXXXXX.p8"  # путь к скачанному .p8 файлу
  key_id: "XXXXXXXXXX"
  team_id: "XXXXXXXXXX"
  bundle_id: "com.yourcompany.bodyfuel"
  sandbox: true   # true для тестов (TestFlight), false для App Store
```
Или через переменные окружения:
```bash
export APNS_KEY_PATH="./keys/AuthKey_XXXXXXXXXX.p8"
export APNS_KEY_ID="XXXXXXXXXX"
export APNS_TEAM_ID="XXXXXXXXXX"
export APNS_BUNDLE_ID="com.yourcompany.bodyfuel"
export APNS_SANDBOX="true"
```

**Когда отправляются push:**
- После успешной генерации тренировки — уведомление на все зарегистрированные устройства пользователя
- Устройства регистрируются через `POST /user/devices`

---

### Минимальный запуск (без внешних сервисов)

Для локальной разработки OpenAI, SendGrid, Twilio и APNs **не обязательны**. Приложение стартует и работает, просто:
- AI-эндпоинты (`/nutrition/analyze`, `/recommendations/refresh`) вернут ошибку
- Задачи отправки кодов создадутся в БД, но executor не сможет их выполнить
- Push-уведомления отправляться не будут

Минимальный рабочий конфиг — только PostgreSQL и MinIO, всё остальное опционально.

---

## Разработка

### Запуск тестов

```bash
go test ./...
```

### Регенерация Swagger-документации

```bash
# Установить swag (если не установлен)
go install github.com/swaggo/swag/cmd/swag@latest

# Сгенерировать docs/
swag init -g cmd/main.go --output docs/
```

### Структура конфига для локальной разработки

Минимальный рабочий `config/config.yaml` (без уведомлений и AI):

```yaml
app:
  http_server:
    host: "0.0.0.0"
    port: 8080
    api_host: "localhost:8080"
  graceful_timeout: "5s"
  tasks_tracking_duration: "13s"
  workouts_config:
    workout_pull_user_interval: "60s"
    limit_generate_workouts: 3

sage:
  level: "debug"
  app_id: 1
  deployment_environment: dev
  system: bodyfuel
  stand_type: dev
  pod_name: bodyfuel-local

postgres:
  host: "localhost"
  port: 5432
  database: "backend_db"
  user: "danila"
  password: "postgres"

minio:
  endpoint: "http://localhost:9000"
  access_key: "minioadmin"
  secret_key: "minioadmin"
  bucket: "avatars"
  region: "us-east-1"
  public_url: "http://localhost:9000/avatars"
  presign_ttl: "5m"
```

### Добавление нового эндпоинта

1. Объявить интерфейс метода в сервисе (`internal/service/<name>/service.go`)
2. Реализовать метод репозитория (`internal/infrastructure/repositories/postgres/`)
3. Добавить Swagger-комментарии к хендлеру
4. Зарегистрировать маршрут в `register*Handlers()`
5. Добавить метод в интерфейс в `api.go` (если новый сервис)
6. Прокинуть зависимость в `app.go`

### Переменные окружения (полный список)

| Переменная | Секция | Описание |
|-----------|--------|----------|
| `HOST` | app | IP для прослушивания |
| `PORT` | app | HTTP-порт |
| `POSTGRES_HOST` | postgres | Хост БД |
| `POSTGRES_PORT` | postgres | Порт БД |
| `POSTGRES_DATABASE` | postgres | Название БД |
| `POSTGRES_USER` | postgres | Пользователь БД |
| `POSTGRES_PASSWORD` | postgres | Пароль БД |
| `MINIO_ENDPOINT` | minio | URL MinIO |
| `MINIO_ACCESS_KEY` | minio | Access key |
| `MINIO_SECRET_KEY` | minio | Secret key |
| `MINIO_BUCKET` | minio | Имя бакета |
| `SENDGRID_API_KEY` | sendgrid | API-ключ SendGrid |
| `SENDGRID_FROM_EMAIL` | sendgrid | Email отправителя |
| `TWILIO_ACCOUNT_SID` | twilio | Account SID |
| `TWILIO_AUTH_TOKEN` | twilio | Auth Token |
| `TWILIO_FROM_PHONE` | twilio | Номер отправителя |
| `APNS_KEY_PATH` | apns | Путь к .p8 ключу |
| `APNS_KEY_ID` | apns | Key ID |
| `APNS_TEAM_ID` | apns | Team ID |
| `APNS_BUNDLE_ID` | apns | Bundle ID приложения |
| `APNS_SANDBOX` | apns | `true` для тестов |
| `OPENAI_API_KEY` | openai | API-ключ OpenAI |
