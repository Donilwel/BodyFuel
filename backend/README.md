# Backend часть сервиса BodyFuel

## Краткая инструкция быстрого старта

1. Swagger документация развёрнута по адресу `http://localhost:8080/swagger/index.html` (при запросе к продовой части нужно заменить `localhost` на продовый хост)
2. Перед стартом необходимо создать (если его нет) или убедиться в правильности заполненных параметров файла `./config/config.yaml`, с шаблоном можно ознакомиться в файле `./config/config.yaml.template`
3. Для корректной работы необходимо иметь запущенные следующие сервисы:
   - **PostgreSQL** — данные подключения указать в `config.yaml` в секции `postgres`
   - **MinIO** — для хранения аватаров пользователей:
     1. Запустить Docker-образ: `docker run -p 9000:9000 minio/minio server /data`
     2. Создать бакет: `mc mb local/avatars`
     3. Открыть публичный доступ: `mc anonymous set public local/avatars`
4. Для отправки уведомлений заполнить секции в `config.yaml`:
   - `sendgrid` — API-ключ SendGrid для отправки email
   - `twilio` — Account SID, Auth Token и номер телефона для отправки SMS
   - `apns` — путь к `.p8` ключу, Key ID, Team ID и Bundle ID для iOS push-уведомлений (APNs HTTP/2)

## Архитектура сервиса

Backend-часть сервиса BodyFuel построена как монолитный сервис с чётким разделением на домены по принципам DDD + Clean Architecture. Все слои разделены и работают независимо друг от друга, взаимодействуя через абстракции (интерфейсы). Сервис содержит следующие слои:

1. **Слой API / Handlers** — обрабатывает HTTP-запросы, валидирует входные данные
2. **Слой сервисов / Use Cases** — вся бизнес-логика приложения
3. **Слой доменной бизнес-логики** — сущности и доменные правила
4. **Слой инфраструктуры** — прямой доступ к данным (PostgreSQL, MinIO)

### Описание сервисов

| Сервис | Описание |
|--------|----------|
| **Auth** | Регистрация и аутентификация пользователя, выдача JWT-токена |
| **Avatar** | Генерация presigned URL для загрузки аватара пользователя в MinIO |
| **CRUD** | Все CRUD-операции для пользователей, параметров, веса, упражнений, тренировок и устройств |
| **Workouts** | Генерация тренировочного плана по параметрам пользователя, фоновая автоматическая генерация |
| **Executor** | Фоновый воркер, выполняющий задачи из очереди: отправка email (SendGrid), SMS (Twilio) и iOS push-уведомлений (APNs) |
| **Nutricion** | (В разработке) Генерация плана питания для пользователя |

## API

Все защищённые эндпоинты требуют заголовок `Authorization: Bearer <token>`.

### Auth

| Метод | Путь | Описание |
|-------|------|----------|
| `POST` | `/auth/register` | Регистрация нового пользователя |
| `POST` | `/auth/login` | Вход и получение JWT-токена |

### User Info

| Метод | Путь | Описание |
|-------|------|----------|
| `GET` | `/user/info` | Получение информации о пользователе |
| `PATCH` | `/user/info` | Обновление информации о пользователе |
| `DELETE` | `/user/info` | Удаление аккаунта |

### User Params

| Метод | Путь | Описание |
|-------|------|----------|
| `GET` | `/user/params` | Получение параметров пользователя |
| `POST` | `/user/params` | Создание параметров пользователя |
| `PATCH` | `/user/params` | Обновление параметров |
| `DELETE` | `/user/params` | Удаление параметров |

### User Weight

| Метод | Путь | Описание |
|-------|------|----------|
| `GET` | `/user/weight` | Получение текущего веса |
| `GET` | `/user/weight/history` | История изменений веса |
| `POST` | `/user/weight` | Добавление записи о весе |
| `PATCH` | `/user/weight/:uuid` | Обновление записи |
| `DELETE` | `/user/weight/:uuid` | Удаление записи |

### User Calories

| Метод | Путь | Описание |
|-------|------|----------|
| `GET` | `/user/calories/history` | История записей о калориях (опционально: `start_date`, `end_date` в RFC3339) |
| `POST` | `/user/calories` | Создание записи о потреблённых/затраченных калориях |
| `PATCH` | `/user/calories/:uuid` | Обновление записи |
| `DELETE` | `/user/calories/:uuid` | Удаление записи |

### User Devices (Push-уведомления)

| Метод | Путь | Описание |
|-------|------|----------|
| `POST` | `/user/devices` | Регистрация device token устройства (APNs) |
| `GET` | `/user/devices` | Список зарегистрированных устройств |
| `DELETE` | `/user/devices/:uuid` | Удаление device token |

### Exercises

| Метод | Путь | Описание |
|-------|------|----------|
| `GET` | `/exercises` | Список упражнений с фильтрацией |
| `GET` | `/exercises/:uuid` | Получение упражнения по ID |
| `POST` | `/exercises` | Создание упражнения |
| `PATCH` | `/exercises/:uuid` | Обновление упражнения |
| `DELETE` | `/exercises/:uuid` | Удаление упражнения |

### Workouts

| Метод | Путь | Описание |
|-------|------|----------|
| `POST` | `/workouts` | Генерация тренировки по параметрам |
| `GET` | `/workouts/history` | История тренировок пользователя |
| `GET` | `/workouts/:uuid` | Получение тренировки с упражнениями |
| `PATCH` | `/workouts/:uuid` | Обновление тренировки |
| `DELETE` | `/workouts/:uuid` | Удаление тренировки |

### Avatars

| Метод | Путь | Описание |
|-------|------|----------|
| `GET` | `/avatars/presign` | Получение presigned URL для загрузки аватара |

### Tasks

| Метод | Путь | Описание |
|-------|------|----------|
| `GET` | `/tasks` | Список задач в очереди |
| `POST` | `/tasks/:uuid/restart` | Перезапуск упавшей задачи |
| `DELETE` | `/tasks/:uuid` | Удаление задачи |

## База данных

Сервис использует PostgreSQL. Все таблицы находятся в схеме `bodyfuel`. Единственный файл миграций: `migrations/00001_init_schema.sql`.

### Схема базы данных

#### Таблица `user_info`
Хранит основную информацию об аккаунте пользователя.

| Поле | Тип | Описание |
|------|-----|----------|
| `id` | UUID | Первичный ключ |
| `username` | TEXT | Никнейм (уникальный) |
| `name` | TEXT | Имя |
| `surname` | TEXT | Фамилия |
| `password` | TEXT | Хэш пароля |
| `email` | TEXT | Почта (уникальная) |
| `phone` | TEXT | Номер телефона |
| `created_at` | TIMESTAMPTZ | Дата создания аккаунта |

#### Таблица `user_params`
Хранит параметры пользователя и его цели.

| Поле | Тип | Описание |
|------|-----|----------|
| `id` | UUID | Первичный ключ |
| `id_user` | UUID | FK → `user_info.id` |
| `height` | INT | Рост в см |
| `photo` | TEXT | Ключ объекта аватара в MinIO |
| `wants` | ENUM | Цель: `lose_weight`, `build_muscle`, `stay_fit` |
| `lifestyle` | ENUM | Образ жизни: `not_active`, `active`, `sportive` |
| `target_workouts_weeks` | INT | Целевое количество тренировок в неделю |
| `target_calories_daily` | INT | Целевая дневная норма калорий |
| `target_weight` | FLOAT | Целевой вес (кг) |

#### Таблица `user_weight`
Хранит историю измерений веса пользователя.

| Поле | Тип | Описание |
|------|-----|----------|
| `id` | UUID | Первичный ключ |
| `id_user` | UUID | FK → `user_info.id` |
| `weight` | FLOAT | Вес в кг |
| `date` | TIMESTAMPTZ | Дата и время измерения |

#### Таблица `exercise`
Каталог упражнений для генерации тренировок.

| Поле | Тип | Описание |
|------|-----|----------|
| `id` | UUID | Первичный ключ |
| `name` | VARCHAR(100) | Название упражнения |
| `description` | TEXT | Описание техники выполнения |
| `level_preparation` | ENUM | Уровень: `beginner`, `medium`, `sportsman` |
| `type_exercise` | ENUM | Тип: `cardio`, `upper_body`, `lower_body`, `full_body`, `flexibility` |
| `place_exercise` | ENUM | Место: `home`, `gym`, `street` |
| `base_count_reps` | INT | Базовое количество повторений |
| `steps` | INT | Количество подходов |
| `avg_calories_per` | DECIMAL | Среднее сжигание калорий за повторение |
| `base_relax_time` | INT | Время отдыха между подходами (сек) |
| `link_gif` | TEXT | Ссылка на анимацию упражнения |

#### Таблица `workout`
Хранит тренировки пользователей.

| Поле | Тип | Описание |
|------|-----|----------|
| `id` | UUID | Первичный ключ |
| `user_id` | UUID | FK → `user_info.id` |
| `level` | ENUM | Уровень: `workout_light`, `workout_middle`, `workout_hard` |
| `status` | ENUM | Статус: `workout_created`, `workout_done`, `workout_in_active`, `workout_failed` |
| `total_calories` | INT | Фактически сожжённые калории |
| `prediction_calories` | INT | Прогнозируемые калории |
| `duration` | INT | Длительность (сек) |
| `created_at` | TIMESTAMP | Дата создания |
| `updated_at` | TIMESTAMP | Дата обновления |

#### Таблица `workouts_exercise`
Связь тренировок и упражнений (m2m).

| Поле | Тип | Описание |
|------|-----|----------|
| `workout_id` | UUID | FK → `workout.id` |
| `exercise_id` | UUID | FK → `exercise.id` |
| `modify_reps` | INT | Скорректированное количество повторений |
| `modify_relax_time` | INT | Скорректированное время отдыха |
| `calories` | INT | Калории за это упражнение в данной тренировке |
| `status` | ENUM | Статус: `pending`, `in_progress`, `completed`, `skipped` |
| `created_at` | TIMESTAMPTZ | Дата создания |
| `updated_at` | TIMESTAMPTZ | Дата обновления |

#### Таблица `tasks`
Очередь отложенных задач для сервиса Executor.

| Поле | Тип | Описание |
|------|-----|----------|
| `task_id` | UUID | Первичный ключ |
| `task_type_nm` | TEXT | Тип задачи: `send_code_email_task`, `send_code_phone_task`, `send_notification_email_task`, `send_notification_phone_task`, `send_push_notification_task` |
| `task_state` | TEXT | Состояние: `running`, `failed` |
| `max_attempts` | INT | Максимальное количество попыток |
| `attempts` | INT | Текущее количество попыток |
| `retry_at` | TIMESTAMPTZ | Время следующей попытки |
| `attribute` | JSONB | Данные задачи (email / phone / device_token / subject / body) |
| `created_at` | TIMESTAMPTZ | Дата создания |
| `updated_at` | TIMESTAMPTZ | Дата обновления |

#### Таблица `user_devices`
Хранит device token устройств пользователей для APNs push-уведомлений.

| Поле | Тип | Описание |
|------|-----|----------|
| `id` | UUID | Первичный ключ |
| `user_id` | UUID | FK → `user_info.id` |
| `device_token` | TEXT | APNs device token |
| `platform` | TEXT | Платформа: `ios`, `android` |
| `created_at` | TIMESTAMPTZ | Дата регистрации |
| `updated_at` | TIMESTAMPTZ | Дата обновления |

#### Таблица `user_calories`
Хранит записи о потреблённых или затраченных калориях пользователя.

| Поле | Тип | Описание |
|------|-----|----------|
| `id` | UUID | Первичный ключ |
| `user_id` | UUID | FK → `user_info.id` |
| `calories` | INT | Количество калорий (0–10000) |
| `description` | TEXT | Комментарий (опционально) |
| `date` | TIMESTAMPTZ | Дата и время записи |
| `created_at` | TIMESTAMPTZ | Дата создания |
| `updated_at` | TIMESTAMPTZ | Дата обновления |

## Уведомления

Executor-сервис опрашивает таблицу `tasks` с интервалом `tasks_tracking_duration` (из конфига) и отправляет уведомления через три канала:

| Канал | Провайдер | Типы задач |
|-------|-----------|------------|
| Email | SendGrid | `send_code_email_task`, `send_notification_email_task` |
| SMS | Twilio | `send_code_phone_task`, `send_notification_phone_task` |
| Push (iOS) | APNs HTTP/2 | `send_push_notification_task` |

При генерации тренировки автоматически создаются задачи для всех доступных каналов: email (если заполнен), SMS (если заполнен номер телефона) и push-уведомление для каждого зарегистрированного устройства.

При ошибке отправки задача повторяется с экспоненциальным/фибоначчи backoff. После превышения `max_attempts` задача переводится в состояние `failed`.