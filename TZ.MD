# [НАЗВАНИЕ] — KARVON
## Stack: Go + GIN · PostgreSQL · Postman
## MVP v3.0 · Июнь 2025

---

# РАЗДЕЛ 1 — АРХИТЕКТУРА

## Стек и библиотеки

| Слой | Технология |
|------|-----------|
| HTTP-фреймворк | Echo v4 |
| База данных | PostgreSQL 15+ |
| ORM / query builder | sqlx + чистый SQL (без ORM) |
| Миграции | golang-migrate |
| Аутентификация | JWT (access + refresh tokens) |
| SMS | telegram api and whatsapp|
| Платежи | Rahmat (webhook) |
| Файлы | MinIO или локальный storage (S3-совместимый) |
| Конфиг | .env через godotenv |
| Логирование | zerolog |
| Контейнеры | Docker + Docker Compose |
| API-документация | Postman Collection |

---

## Структура проекта

```
/cmd/api          — точка входа (main.go)
/internal
  /config         — загрузка .env
  /handler        — HTTP-хэндлеры (по модулям)
  /service        — бизнес-логика
  /repository     — запросы к БД
  /middleware     — JWT, роли, rate limit
  /model          — структуры данных (Go structs)
  /dto            — Request / Response объекты
/migrations       — SQL-файлы миграций
/pkg
  /eskiz          — клиент Eskiz SMS
  /rahmat         — клиент Rahmat платежи
  /storage        — работа с файлами
/docker-compose.yml
/.env.example
```

---

## Переменные окружения (.env)

```
APP_PORT=8080
DB_URL=postgres://user:pass@localhost:5432/dbname?sslmode=disable
JWT_SECRET=...
JWT_ACCESS_TTL=15m
JWT_REFRESH_TTL=30d
ESKIZ_EMAIL=...
ESKIZ_PASSWORD=...
RAHMAT_MERCHANT_ID=...
RAHMAT_SECRET_KEY=...
STORAGE_TYPE=local   # или minio
STORAGE_PATH=./uploads
```

---

## Формат всех API-ответов

**Успех:**
```json
{
  "success": true,
  "data": { ... }
}
```

**Ошибка:**
```json
{
  "success": false,
  "error": {
    "code": "INSUFFICIENT_TOKENS",
    "message": "Недостаточно токенов"
  }
}
```

**Пагинация (списки):**
```json
{
  "success": true,
  "data": [ ... ],
  "meta": {
    "total": 150,
    "page": 1,
    "per_page": 20
  }
}
```

Base URL: `GET /api/v1/...`

Аутентификация: заголовок `Authorization: Bearer <access_token>`

---

## Коды ошибок (ErrorCode)

| Код | Смысл |
|-----|-------|
| `UNAUTHORIZED` | Не авторизован |
| `FORBIDDEN` | Нет прав |
| `NOT_FOUND` | Объект не найден |
| `VALIDATION_ERROR` | Ошибка валидации тела запроса |
| `OTP_EXPIRED` | OTP истёк |
| `OTP_INVALID` | Неверный OTP |
| `OTP_MAX_ATTEMPTS` | Превышено число попыток |
| `COMPANY_NOT_VERIFIED` | Компания не верифицирована |
| `INSUFFICIENT_TOKENS` | Нет токенов |
| `FREE_LISTING_USED` | Бесплатное объявление уже использовано |
| `PAYMENT_FAILED` | Ошибка платежа |
| `ALREADY_EXISTS` | Дубликат (например, компания с таким ИНН) |

---

# РАЗДЕЛ 2 — СХЕМА БАЗЫ ДАННЫХ

Все таблицы создаются через файлы миграций в `/migrations`.

---

### users
Основной аккаунт. Одна запись на телефон.

| Колонка | Тип | Описание |
|---------|-----|----------|
| id | UUID PK | gen_random_uuid() |
| phone | VARCHAR(20) UNIQUE NOT NULL | Номер телефона |
| name | VARCHAR(100) | Имя / псевдоним |
| email | VARCHAR(100) | |
| whatsapp | VARCHAR(20) | |
| telegram | VARCHAR(50) | |
| city | VARCHAR(100) | |
| country | VARCHAR(100) | |
| token_balance | INT DEFAULT 5 | Баланс токенов |
| role | ENUM: user / moderator / super_admin | DEFAULT user |
| is_blocked | BOOLEAN DEFAULT false | |
| created_at | TIMESTAMPTZ | |
| updated_at | TIMESTAMPTZ | |

---

### otp_codes
Временные OTP-коды для входа.

| Колонка | Тип | Описание |
|---------|-----|----------|
| id | UUID PK | |
| phone | VARCHAR(20) NOT NULL | |
| code | VARCHAR(6) NOT NULL | 6-значный код |
| expires_at | TIMESTAMPTZ | +5 минут от created_at |
| used | BOOLEAN DEFAULT false | |
| attempts | INT DEFAULT 0 | Счётчик неверных попыток |
| created_at | TIMESTAMPTZ | |

---

### companies
Юрлица, привязанные к пользователю. Один user → много companies.

| Колонка | Тип | Описание |
|---------|-----|----------|
| id | UUID PK | |
| user_id | UUID FK → users | Владелец |
| name | VARCHAR(200) NOT NULL | Название компании |
| inn | VARCHAR(20) NOT NULL | ИНН |
| region | VARCHAR(100) | |
| phone | VARCHAR(20) | Телефон компании |
| inn_doc_url | TEXT | Ссылка на файл ИНН |
| reg_doc_url | TEXT | Ссылка на свидетельство |
| status | ENUM: pending / approved / rejected / docs_requested | DEFAULT pending |
| rejection_reason | TEXT | Причина отклонения |
| moderator_id | UUID FK → users | Кто обрабатывал |
| verified_at | TIMESTAMPTZ | Дата одобрения |
| created_at | TIMESTAMPTZ | |
| updated_at | TIMESTAMPTZ | |

---

### cargo_listings
Объявления о грузах. Только от верифицированных компаний.

| Колонка | Тип | Описание |
|---------|-----|----------|
| id | UUID PK | |
| company_id | UUID FK → companies | |
| user_id | UUID FK → users | Публикатор |
| type | ENUM: domestic / international | Тип перевозки |
| cargo_name | VARCHAR(200) | Наименование груза |
| cargo_type | VARCHAR(50) | навалом / паллеты / бочки и т.д. |
| is_adr | BOOLEAN | Опасный груз |
| has_temp_regime | BOOLEAN | |
| temp_min / temp_max | DECIMAL | Температурный режим |
| weight_ton | DECIMAL | |
| volume_m3 | DECIMAL | |
| places_count | INT | |
| length_m / width_m / height_m | DECIMAL | Габариты |
| loading_type | ENUM: ftl / ltl / partial | |
| body_types | TEXT[] | Массив типов кузова |
| vehicle_type | VARCHAR(50) | тягач+полуприцеп и т.д. |
| tractor_axles | INT | |
| trailer_axles | INT | |
| only_recoupling | BOOLEAN | Только перецепка |
| from_city | VARCHAR(100) | |
| from_country | VARCHAR(100) | |
| from_date | DATE | |
| loading_method | VARCHAR(50) | |
| to_city | VARCHAR(100) | |
| to_country | VARCHAR(100) | |
| to_date | DATE | |
| unloading_method | VARCHAR(50) | |
| transit_countries | TEXT[] | |
| border_crossing | VARCHAR(100) | КПП |
| customs_type | ENUM: export / import / transit | |
| incoterms | VARCHAR(20) | |
| permits | TEXT[] | TIR / CMR / ADR и т.д. |
| rate_mode | ENUM: negotiable / fixed / announcement | |
| rate_amount | DECIMAL | |
| rate_currency | ENUM: UZS / USD | |
| rate_vat | BOOLEAN | |
| payment_terms | TEXT | |
| payment_method | VARCHAR(50) | |
| prepayment_percent | INT | |
| duration_hours | INT | 24 / 48 / 72 |
| expires_at | TIMESTAMPTZ | created_at + duration |
| notes | TEXT | |
| status | ENUM: active / archived / completed | DEFAULT active |
| is_template | BOOLEAN DEFAULT false | |
| template_name | VARCHAR(100) | |
| is_boosted | BOOLEAN DEFAULT false | |
| boost_expires_at | TIMESTAMPTZ | |
| is_paid | BOOLEAN DEFAULT false | Платное или бесплатное |
| views_count | INT DEFAULT 0 | |
| contacts_bought_count | INT DEFAULT 0 | |
| created_at | TIMESTAMPTZ | |
| updated_at | TIMESTAMPTZ | |

---

### cargo_waypoints
Промежуточные пункты маршрута груза.

| Колонка | Тип | Описание |
|---------|-----|----------|
| id | UUID PK | |
| cargo_listing_id | UUID FK → cargo_listings | |
| sort_order | INT | Порядок |
| waypoint_type | ENUM: border / customs / reload / passthrough | |
| city | VARCHAR(100) | |
| country | VARCHAR(100) | |

---

### fleet_vehicles
Автопарк перевозчика.

| Колонка | Тип | Описание |
|---------|-----|----------|
| id | UUID PK | |
| user_id | UUID FK → users | |
| composition | ENUM: tractor_semi / truck_trailer | |
| brand | VARCHAR(100) | Марка |
| model | VARCHAR(100) | Модель |
| plate_number | VARCHAR(20) | Гос. номер |
| year | INT | Год выпуска |
| country_reg | VARCHAR(100) | Страна регистрации |
| body_type | VARCHAR(50) | |
| volume_m3 | DECIMAL | |
| load_capacity_ton | DECIMAL | |
| length_m / width_m / height_m | DECIMAL | |
| tractor_axles | INT | |
| trailer_axles | INT | |
| has_tir | BOOLEAN | |
| has_cmr | BOOLEAN | |
| has_adr | BOOLEAN | |
| has_ekmt | BOOLEAN | |
| allowed_countries | TEXT[] | |
| driver_language | TEXT[] | |
| created_at | TIMESTAMPTZ | |

---

### transport_listings
Объявления свободного транспорта.

| Колонка | Тип | Описание |
|---------|-----|----------|
| id | UUID PK | |
| vehicle_id | UUID FK → fleet_vehicles | |
| user_id | UUID FK → users | |
| current_location | VARCHAR(150) | |
| from_city | VARCHAR(100) | |
| to_city | VARCHAR(100) | |
| ready_date | DATE | |
| status | ENUM: active / archived | |
| is_boosted | BOOLEAN | |
| boost_expires_at | TIMESTAMPTZ | |
| is_paid | BOOLEAN | |
| views_count | INT DEFAULT 0 | |
| created_at | TIMESTAMPTZ | |

---

### warehouse_listings
Карточки складов.

| Колонка | Тип | Описание |
|---------|-----|----------|
| id | UUID PK | |
| company_id | UUID FK → companies | |
| user_id | UUID FK → users | |
| warehouse_type | ENUM: regular / cold / customs | |
| name | VARCHAR(200) | |
| region | VARCHAR(100) | |
| address | TEXT | |
| lat / lng | DECIMAL | Геометка |
| phone_main | VARCHAR(20) | |
| phone_extra | VARCHAR(20) | |
| contact_person | VARCHAR(100) | |
| email | VARCHAR(100) | |
| website | VARCHAR(200) | |
| specialization | TEXT[] | Тип груза |
| area_total_m2 | DECIMAL | |
| area_free_m2 | DECIMAL | |
| ceiling_height_m | DECIMAL | |
| heating_type | ENUM: heated / unheated / open / closed | |
| storage_type | TEXT[] | стеллажное / напольное / паллетоместа |
| temp_min / temp_max | DECIMAL | Только для cold |
| cold_chamber_types | TEXT[] | Только для cold |
| customs_license_number | VARCHAR(100) | Только для customs |
| customs_license_issued | DATE | |
| customs_license_expires | DATE | |
| infrastructure | TEXT[] | Погрузчик / ж/д / парковка и т.д. |
| services | TEXT[] | Кроссдок / упаковка и т.д. |
| working_hours | JSONB | { "mon": "9:00-18:00", ... } |
| is_boosted | BOOLEAN | |
| boost_expires_at | TIMESTAMPTZ | |
| is_paid | BOOLEAN | |
| status | ENUM: active / archived | |
| views_count | INT DEFAULT 0 | |
| created_at | TIMESTAMPTZ | |

---

### listing_media
Фото и документы для всех типов объявлений.

| Колонка | Тип | Описание |
|---------|-----|----------|
| id | UUID PK | |
| entity_type | ENUM: cargo / transport / warehouse / vehicle | |
| entity_id | UUID | ID объявления |
| file_url | TEXT | Путь к файлу |
| file_type | ENUM: photo / document | |
| original_name | VARCHAR(200) | |
| sort_order | INT | |
| created_at | TIMESTAMPTZ | |

---

### contact_views
История просмотренных контактов.

| Колонка | Тип | Описание |
|---------|-----|----------|
| id | UUID PK | |
| user_id | UUID FK → users | Кто смотрел |
| listing_type | ENUM: cargo / transport / warehouse | |
| listing_id | UUID | |
| tokens_spent | INT | 0 если подписка |
| viewed_at | TIMESTAMPTZ | |

Индекс: (user_id, listing_type, listing_id) — для быстрой проверки повторного просмотра.

---

### token_transactions
Лог всех движений токенов.

| Колонка | Тип | Описание |
|---------|-----|----------|
| id | UUID PK | |
| user_id | UUID FK → users | |
| type | ENUM: credit / debit | |
| amount | INT | Количество токенов |
| reason | ENUM: registration / purchase / contact_view / manual | |
| reference_id | UUID | ID платежа или просмотра |
| balance_after | INT | Баланс после операции |
| created_at | TIMESTAMPTZ | |

---

### payment_orders
Заявки на оплату.

| Колонка | Тип | Описание |
|---------|-----|----------|
| id | UUID PK | |
| user_id | UUID FK → users | |
| payment_type | ENUM: tokens / subscription / listing / boost | |
| item_id | VARCHAR(100) | Идентификатор тарифа |
| amount | DECIMAL | Сумма |
| currency | ENUM: UZS / USD | |
| status | ENUM: pending / paid / failed / cancelled | DEFAULT pending |
| rahmat_order_id | VARCHAR(200) | ID от Rahmat |
| payment_method | VARCHAR(50) | uzcard / humo / visa и т.д. |
| paid_at | TIMESTAMPTZ | |
| created_at | TIMESTAMPTZ | |

---

### subscriptions
Активные подписки пользователей.

| Колонка | Тип | Описание |
|---------|-----|----------|
| id | UUID PK | |
| user_id | UUID FK → users | |
| plan | ENUM: week / month / year | |
| starts_at | TIMESTAMPTZ | |
| expires_at | TIMESTAMPTZ | |
| is_active | BOOLEAN DEFAULT true | |
| payment_order_id | UUID FK → payment_orders | |
| created_at | TIMESTAMPTZ | |

---

### pricing_config
Цены управляются супер-админом. Не хардкодятся в коде.

| Колонка | Тип | Описание |
|---------|-----|----------|
| id | UUID PK | |
| key | VARCHAR(100) UNIQUE | Уникальный ключ тарифа |
| label | VARCHAR(200) | Название (для UI) |
| price_uzs | DECIMAL | Цена в сумах |
| price_usd | DECIMAL | Цена в долларах |
| tokens_amount | INT | Только для токен-пакетов |
| duration_days | INT | Для подписок и Boost |
| is_active | BOOLEAN | |
| updated_by | UUID FK → users | Кто последний менял |
| updated_at | TIMESTAMPTZ | |

Начальные ключи: `tokens_mini`, `tokens_basic`, `tokens_standard`, `tokens_max`, `sub_week`, `sub_month`, `sub_year`, `listing_paid`, `boost_1day`, `boost_3day`, `boost_7day`

---

### saved_routes
Сохранённые маршруты перевозчика.

| Колонка | Тип | Описание |
|---------|-----|----------|
| id | UUID PK | |
| user_id | UUID FK → users | |
| from_city | VARCHAR(100) | |
| to_city | VARCHAR(100) | |
| notifications_enabled | BOOLEAN DEFAULT true | |
| created_at | TIMESTAMPTZ | |

---

### favorites
Избранные объявления.

| Колонка | Тип | Описание |
|---------|-----|----------|
| id | UUID PK | |
| user_id | UUID FK → users | Кто добавил |
| listing_type | ENUM: cargo / transport / warehouse | |
| listing_id | UUID | |
| created_at | TIMESTAMPTZ | |

Уникальный индекс: (user_id, listing_type, listing_id)

---

### notifications
Внутренние уведомления платформы (без SMS).

| Колонка | Тип | Описание |
|---------|-----|----------|
| id | UUID PK | |
| user_id | UUID FK → users | Получатель |
| type | VARCHAR(50) | new_cargo_on_route / contact_purchased / sub_expiring / verification_approved / etc. |
| title | VARCHAR(200) | |
| body | TEXT | |
| is_read | BOOLEAN DEFAULT false | |
| meta | JSONB | Дополнительные данные (listing_id и т.д.) |
| created_at | TIMESTAMPTZ | |

---

# РАЗДЕЛ 3 — MIDDLEWARE

| Middleware | Назначение |
|-----------|-----------|
| `AuthMiddleware` | Проверяет JWT, кладёт user в context |
| `RoleMiddleware(roles...)` | Проверяет role пользователя: user / moderator / super_admin |
| `CompanyVerifiedMiddleware` | Проверяет, есть ли у user хоть одна verified company |
| `RateLimitMiddleware` | Ограничение запросов (особенно для /auth/send-otp) |
| `CORSMiddleware` | Разрешённые origins |
| `LoggerMiddleware` | Логирование запросов |

---

# РАЗДЕЛ 4 — ПЛАН РАЗРАБОТКИ (10 ДНЕЙ)

---

## ДЕНЬ 1 — Инициализация проекта и база данных

**Цель:** Рабочий скелет, все таблицы созданы через миграции, проект стартует.

### Задачи

1. **Инициализировать Go-проект**
   - `go mod init [название]`
   - Установить зависимости: echo, sqlx, pgx, golang-migrate, godotenv, zerolog, uuid

2. **Docker Compose**
   - Сервис PostgreSQL 15
   - Сервис приложения (опционально, можно позже)
   - Порты, volumes для данных БД

3. **Конфигурация**
   - Загрузка всех переменных из `.env`
   - Структура Config с валидацией (обязательные поля)

4. **Создать все миграции**
   - Один файл на таблицу (или сгруппировать логически)
   - Порядок: users → otp_codes → companies → cargo_listings → cargo_waypoints → fleet_vehicles → transport_listings → warehouse_listings → listing_media → contact_views → token_transactions → payment_orders → subscriptions → pricing_config → saved_routes → favorites → notifications
   - Создать начальный seed для `pricing_config` — заполнить все тарифы нулевыми ценами

5. **Базовый HTTP-сервер**
   - Echo приложение стартует на APP_PORT
   - Healthcheck: `GET /health` → `{"status": "ok"}`
   - Подключение к PostgreSQL с проверкой ping

6. **Postman**
   - Создать Collection `[НАЗВАНИЕ] API v1`
   - Переменные: `base_url`, `access_token`, `refresh_token`
   - Добавить запрос `GET /health`

### Бизнес-правила
- Если БД недоступна при старте → приложение завершается с ошибкой (fail fast)
- Миграции запускаются автоматически при старте (или отдельной командой)

---

## ДЕНЬ 2 — Аутентификация (OTP + JWT)

**Цель:** Пользователь может зарегистрироваться и войти через телефон + OTP.

### Задачи

1. **Интеграция Eskiz.uz**
   - Получить токен через email/password Eskiz
   - Метод отправки SMS с OTP-кодом
   - Токен Eskiz кешируется в памяти (TTL ~24h), обновляется автоматически

2. **Генерация OTP**
   - 6-значный числовой код
   - Запись в таблицу `otp_codes`, TTL = 5 минут
   - Старые коды для этого телефона — деактивировать (used = true)

3. **Rate limit для OTP**
   - Максимум 3 активных запроса OTP в час с одного телефона
   - При превышении — ошибка `OTP_MAX_ATTEMPTS`, указать сколько секунд ждать

4. **JWT**
   - Access token: содержит `user_id`, `role`, TTL = 15 минут
   - Refresh token: содержит `user_id`, TTL = 30 дней, хранится в таблице (или stateless)
   - При refresh — выдаётся новая пара токенов

5. **AuthMiddleware**
   - Читает `Authorization: Bearer <token>`
   - Валидирует JWT, достаёт user из БД
   - Если пользователь заблокирован — возвращает 403

### API Endpoints

| Метод | Путь | Auth | Описание |
|-------|------|------|----------|
| POST | `/api/v1/auth/send-otp` | Нет | Отправить OTP на телефон |
| POST | `/api/v1/auth/verify-otp` | Нет | Проверить OTP, получить токены |
| POST | `/api/v1/auth/refresh` | Нет | Обновить access token |
| POST | `/api/v1/auth/logout` | JWT | Инвалидировать refresh token |

**POST /auth/send-otp**
- Body: `{ "phone": "+998901234567" }`
- Response: `{ "expires_in": 300 }` (секунды до истечения кода)
- Логика: если пользователя нет — создаётся запись с пустым именем (регистрация завершится на следующем шаге)

**POST /auth/verify-otp**
- Body: `{ "phone": "+998901234567", "code": "123456" }`
- Response: `{ "access_token", "refresh_token", "is_new_user": true/false }`
- Если `is_new_user = true` → фронтенд показывает шаг "Введите имя"
- Если неверный код → `attempts++`, при attempts >= 5 → код блокируется
- Новому пользователю начислить 5 токенов (insert в token_transactions)

**POST /auth/verify-otp (завершение регистрации)**
- Отдельный endpoint или флаг `name` в том же запросе — обсудить с фронтом
- Рекомендация: `POST /api/v1/auth/complete-registration` → `{ "name": "Алишер" }`
- Обновляет users.name. Доступен только для новых пользователей без имени.

### Бизнес-правила
- Телефон нормализовать перед сохранением: убрать пробелы, скобки, дефисы
- OTP проверяется максимум за 5 минут (expires_at)
- После успешной верификации — OTP помечается used = true
- Если пользователь уже существует — это вход, не регистрация. Токены выдаются в обоих случаях.

### Postman
- Добавить все 4 запроса
- В `verify-otp` — тест: автоматически сохранять `access_token` и `refresh_token` в переменные коллекции

---

## ДЕНЬ 3 — Профиль и компании

**Цель:** Пользователь управляет своим профилем и добавляет юрлица.

### Задачи

1. **Профиль пользователя**
   - Получить и обновить личные данные
   - Поля для редактирования: name, email, whatsapp, telegram, city, country

2. **Загрузка файлов**
   - Сервис для сохранения файлов (локально или MinIO)
   - Принимает multipart/form-data
   - Возвращает URL файла
   - Ограничения: PDF и изображения, максимум 10 МБ на файл

3. **Компании**
   - Один пользователь → несколько компаний
   - Каждая компания верифицируется отдельно
   - Одно бесплатное объявление — на весь аккаунт (проверяется по user_id, не company_id)

### API Endpoints

| Метод | Путь | Auth | Описание |
|-------|------|------|----------|
| GET | `/api/v1/users/me` | JWT | Мой профиль |
| PUT | `/api/v1/users/me` | JWT | Обновить профиль |
| GET | `/api/v1/users/me/stats` | JWT | Статистика (токены, подписка, активные объявления) |
| POST | `/api/v1/companies` | JWT | Добавить компанию (шаг 1 и 2 формы) |
| GET | `/api/v1/companies` | JWT | Мои компании |
| GET | `/api/v1/companies/:id` | JWT | Одна компания |
| PUT | `/api/v1/companies/:id` | JWT | Обновить данные компании |
| POST | `/api/v1/upload` | JWT | Загрузить файл → вернуть URL |

**POST /companies**
- Body: `{ name, inn, region, phone, inn_doc_url, reg_doc_url }`
- `inn_doc_url` и `reg_doc_url` — ссылки, полученные из `/upload`
- Статус по умолчанию — `pending`
- Создаётся уведомление модераторам (через notifications или очередь)

**GET /companies**
- Возвращает все компании пользователя
- Включает: name, inn, status, rejection_reason, verified_at

### Бизнес-правила
- ИНН уникален в рамках платформы. Если такой ИНН уже есть в другой верифицированной компании — ошибка `ALREADY_EXISTS`
- Пользователь не может редактировать компанию пока она в статусе `pending`
- При статусе `docs_requested` — пользователь может обновить документы (inn_doc_url, reg_doc_url)

---

## ДЕНЬ 4 — Верификация компаний (Модератор)

**Цель:** Модератор обрабатывает заявки за 24 часа.

### Задачи

1. **Middleware для ролей**
   - `RoleMiddleware("moderator", "super_admin")` — применяется ко всем /moderator/* роутам
   - `RoleMiddleware("super_admin")` — для /admin/* роутов

2. **Очередь верификации**
   - Список заявок со статусом `pending` или `docs_requested`, отсортированных по дате
   - Каждая заявка имеет поле `deadline` = created_at + 24 часа
   - Если до дедлайна < 2 часов — помечается `is_urgent: true`

3. **Обработка заявок**
   - Одобрить → status = approved, verified_at = now()
   - Отклонить → status = rejected, rejection_reason обязателен
   - Запросить документы → status = docs_requested, пользователь получает уведомление

4. **Уведомление пользователю**
   - При любом изменении статуса — создавать запись в таблице `notifications`

### API Endpoints

| Метод | Путь | Auth | Описание |
|-------|------|------|----------|
| GET | `/api/v1/moderator/queue` | Moderator | Очередь заявок |
| GET | `/api/v1/moderator/queue/:id` | Moderator | Детали заявки |
| POST | `/api/v1/moderator/queue/:id/approve` | Moderator | Одобрить |
| POST | `/api/v1/moderator/queue/:id/reject` | Moderator | Отклонить |
| POST | `/api/v1/moderator/queue/:id/request-docs` | Moderator | Запросить документы |
| GET | `/api/v1/moderator/history` | Moderator | История своих решений |

**GET /moderator/queue**
- Query params: `status=pending` (по умолчанию)
- Возвращает: id, company name, inn, created_at, deadline, is_urgent, user_name, user_phone

**POST /moderator/queue/:id/reject**
- Body: `{ "reason": "Документы не читаемы, загрузите скан лучшего качества" }`
- reason обязателен

### Бизнес-правила
- Модератор видит ТОЛЬКО очередь верификации. Никаких платежей, статистики, пользователей.
- После одобрения компании — пользователь может публиковать грузы и склады без дополнительной модерации
- Один модератор не может обработать заявку дважды (только первый успешный action)

---

## ДЕНЬ 5 — Объявления о грузах

**Цель:** Верифицированные компании публикуют грузы. Все работает с фильтрами.

### Задачи

1. **Проверка перед публикацией**
   - Пользователь должен иметь хотя бы одну verified компанию
   - Проверка бесплатного объявления: если у пользователя нет ни одного активного бесплатного → первое объявление бесплатно. Иначе — нужна оплата (проверить `pricing_config`)

2. **Создание объявления**
   - Полная форма из ТЗ (все поля раздела 4.2)
   - Тип перевозки (domestic/international) определяет набор обязательных полей
   - Автоматически рассчитать `expires_at` = created_at + duration_hours

3. **Фоновая задача — истечение объявлений**
   - Каждые 30 минут проверять: expires_at < now() AND status = active
   - Переводить в status = archived
   - Уведомить пользователя

4. **Шаблоны**
   - Сохранить текущее объявление как шаблон (is_template = true, template_name = ...)
   - Создать новое объявление из шаблона — копирует все поля, сбрасывает dates/status

### API Endpoints

| Метод | Путь | Auth | Описание |
|-------|------|------|----------|
| POST | `/api/v1/listings/cargo` | JWT + Verified | Создать объявление |
| GET | `/api/v1/listings/cargo` | Нет | Список с фильтрами |
| GET | `/api/v1/listings/cargo/:id` | Нет | Карточка объявления |
| PUT | `/api/v1/listings/cargo/:id` | JWT + Owner | Редактировать |
| PATCH | `/api/v1/listings/cargo/:id/status` | JWT + Owner | Изменить статус |
| POST | `/api/v1/listings/cargo/:id/duplicate` | JWT + Verified | Дублировать |
| DELETE | `/api/v1/listings/cargo/:id` | JWT + Owner | Удалить |
| GET | `/api/v1/listings/cargo/:id/stats` | JWT + Owner | Статистика объявления |
| GET | `/api/v1/listings/cargo/templates` | JWT | Мои шаблоны |
| POST | `/api/v1/listings/cargo/templates` | JWT + Verified | Сохранить как шаблон |
| POST | `/api/v1/listings/cargo/from-template/:id` | JWT + Verified | Создать из шаблона |

**GET /listings/cargo — Query параметры фильтров:**
```
from_city, to_city, type (domestic/international),
loading_type (ftl/ltl/partial), cargo_type,
body_types (массив), weight_min, weight_max,
date_from, verified_only (bool),
sort (newest/date/weight),
page, per_page
```

**GET /listings/cargo/:id — логика отображения контактов:**
- Телефон, WhatsApp, Telegram: если у пользователя нет opened contact → показывать `null` (фронтенд применяет blur)
- Если есть opened contact (проверить contact_views) → показывать реальные данные

**GET /listings/cargo/:id/stats**
- Возвращает: views_count, contacts_bought_count, список пользователей в избранном (name, company, date)

### Бизнес-правила
- При загрузке списка — boosted объявления всегда вверху (сортировка: is_boosted DESC, created_at DESC)
- Автор объявления видит телефон без токена
- После истечения объявления — создать уведомление пользователю
- При дублировании — все поля копируются, кроме: status (active), expires_at (пересчитать), dates погрузки/выгрузки (null)

---

## ДЕНЬ 6 — Автопарк и транспорт

**Цель:** Перевозчики ведут автопарк и публикуют свободные машины.

### Задачи

1. **Автопарк (fleet)** — личный кабинет, не виден публично
   - CRUD машин
   - Загрузка фото и документов через /upload

2. **Публикация транспорта**
   - Из машины автопарка создаётся transport_listing
   - Одна машина — одно активное объявление одновременно
   - Та же логика бесплатного объявления, что и для грузов

3. **Список транспорта с фильтрами**
   - Все поля раздела 5.3 из ТЗ
   - Специальный фильтр «Китайский маршрут»: показывать только тягачи/грузовики с 5, 6 или 8 осями суммарно

### API Endpoints

| Метод | Путь | Auth | Описание |
|-------|------|------|----------|
| POST | `/api/v1/fleet` | JWT | Добавить машину |
| GET | `/api/v1/fleet` | JWT | Мой автопарк |
| GET | `/api/v1/fleet/:id` | JWT | Одна машина |
| PUT | `/api/v1/fleet/:id` | JWT + Owner | Редактировать |
| DELETE | `/api/v1/fleet/:id` | JWT + Owner | Удалить |
| POST | `/api/v1/fleet/:id/publish` | JWT | Опубликовать транспорт |
| GET | `/api/v1/listings/transport` | Нет | Список транспорта |
| GET | `/api/v1/listings/transport/:id` | Нет | Карточка |
| PUT | `/api/v1/listings/transport/:id` | JWT + Owner | Редактировать объявление |
| PATCH | `/api/v1/listings/transport/:id/status` | JWT + Owner | Изменить статус |
| DELETE | `/api/v1/listings/transport/:id` | JWT + Owner | Удалить |

**GET /listings/transport — Query параметры:**
```
from_city, to_city, body_type, total_axles_min, total_axles_max,
china_route (bool — фильтр 5/6/8 осей),
volume_min, volume_max,
load_capacity_min, load_capacity_max,
has_tir, has_cmr, has_adr,
allowed_country, ready_date,
sort, page, per_page
```

### Бизнес-правила
- Удаление машины из автопарка: если есть активное объявление → сначала архивировать объявление, потом удалять машину
- При удалении транспортного объявления — машина остаётся в автопарке
- Гос. номер машины скрыт для неавторизованных пользователей

---

## ДЕНЬ 7 — Склады

**Цель:** Три типа складов с разными полями и картой.

### Задачи

1. **CRUD складов**
   - Только от верифицированных компаний
   - Каждый тип (regular/cold/customs) имеет свои обязательные поля
   - Геометка: lat/lng сохраняются в БД

2. **Валидация по типу**
   - `cold`: temp_min, temp_max, cold_chamber_types — обязательны
   - `customs`: customs_license_number, customs_license_issued, customs_license_expires — обязательны
   - `regular`: только общие поля

3. **Прогресс-бар занятости**
   - В ответе вычислять: `occupancy_percent = ((area_total_m2 - area_free_m2) / area_total_m2) * 100`

### API Endpoints

| Метод | Путь | Auth | Описание |
|-------|------|------|----------|
| POST | `/api/v1/warehouses` | JWT + Verified | Добавить склад |
| GET | `/api/v1/warehouses` | Нет | Список складов |
| GET | `/api/v1/warehouses/:id` | Нет | Карточка |
| PUT | `/api/v1/warehouses/:id` | JWT + Owner | Редактировать |
| PATCH | `/api/v1/warehouses/:id/status` | JWT + Owner | Архивировать |
| DELETE | `/api/v1/warehouses/:id` | JWT + Owner | Удалить |

**GET /warehouses — Query параметры:**
```
region, warehouse_type (regular/cold/customs),
area_min, area_max,
temp_min, temp_max (только для cold),
services (массив: crossdock/railway/parking_trucks),
sort, page, per_page
```

### Бизнес-правила
- Свободная площадь не может быть больше общей площади
- При фильтре по температурному режиму — учитывать что cold склад может хранить груз если temp попадает в диапазон [temp_min, temp_max]

---

## ДЕНЬ 8 — Токены и просмотр контактов

**Цель:** Основная монетизация. Пользователь платит токен чтобы увидеть телефон.

### Задачи

1. **Логика просмотра контакта — 5 сценариев**

| Сценарий | Действие |
|---------|---------|
| Гость (не авторизован) | Вернуть 401, фронт редиректит на вход |
| Авторизован, нет токенов, нет подписки | Вернуть 402 с данными тарифов (пакеты + подписки) |
| Есть активная подписка | Показать контакт, токен НЕ списывать. Записать в contact_views (tokens_spent = 0) |
| Уже смотрел (запись в contact_views за последние 30 дней) | Показать контакт, токен НЕ списывать |
| Есть токены, нет подписки, первый просмотр | Списать 1 токен, показать контакт |

2. **Атомарность операции**
   - Списание токена и запись contact_views — в одной транзакции PostgreSQL
   - Если транзакция упала — контакт не показывать

3. **История контактов**
   - Список всех просмотренных контактов за 30 дней
   - С деталями объявления (краткая карточка)

### API Endpoints

| Метод | Путь | Auth | Описание |
|-------|------|------|----------|
| POST | `/api/v1/contacts/view` | JWT | Открыть контакт |
| GET | `/api/v1/contacts/history` | JWT | История просмотров |
| GET | `/api/v1/users/me/tokens` | JWT | Баланс и история токенов |

**POST /contacts/view**
- Body: `{ "listing_type": "cargo", "listing_id": "uuid" }`
- Response (успех): `{ phone, phone_extra, whatsapp, telegram, email, company_name, city, country }`
- Response (402): `{ "error": { "code": "INSUFFICIENT_TOKENS" }, "data": { "packages": [...], "subscriptions": [...] } }`

**GET /contacts/history**
- Возвращает записи из contact_views за последние 30 дней
- С join на объявление: краткое описание (от/до, груз/транспорт/склад)

### Бизнес-правила
- Просмотр собственного объявления — токен не списывать, просто вернуть данные
- contacts_bought_count в объявлении инкрементируется только когда реально списан токен (не при подписке и не при повторном просмотре)
- Уведомление автору объявления: при каждой покупке контакта → создать notification типа `contact_purchased` с именем/компанией покупателя

---

## ДЕНЬ 9 — Монетизация (Rahmat, подписка, Boost)

**Цель:** Полный платёжный цикл. Rahmat webhook активирует токены/подписки автоматически.

### Задачи

1. **Получение тарифов**
   - Все активные тарифы берутся из `pricing_config`
   - Никаких хардкодированных цен в коде

2. **Создание платежа**
   - Создать payment_order со статусом `pending`
   - Передать данные в Rahmat API (сумма, merchant_id, callback URL, metadata)
   - Rahmat возвращает URL платёжной страницы → вернуть на фронт

3. **Webhook от Rahmat**
   - Принять POST на `/api/v1/payments/webhook`
   - Проверить подпись (HMAC или ключ из Rahmat)
   - Обновить payment_order.status = paid
   - В зависимости от `payment_type`:
     - `tokens` → начислить токены (update token_balance + insert token_transaction)
     - `subscription` → создать запись в subscriptions с датами
     - `listing` → пометить объявление как is_paid = true (создать если ещё нет)
     - `boost` → обновить listing: is_boosted = true, boost_expires_at = now() + duration_days

4. **Boost истекает**
   - Фоновая задача каждый час: boost_expires_at < now() → is_boosted = false

5. **Ручное пополнение (наличные)**
   - Только супер-админ
   - Создаёт token_transaction type=credit, reason=manual

6. **Управление ценами (супер-админ)**
   - Читать и обновлять pricing_config
   - Лог изменений: кто, когда, что изменил (поле updated_by + updated_at)

### API Endpoints

| Метод | Путь | Auth | Описание |
|-------|------|------|----------|
| GET | `/api/v1/payments/packages` | Нет | Все тарифы |
| POST | `/api/v1/payments/create` | JWT | Создать платёж → URL Rahmat |
| POST | `/api/v1/payments/webhook` | Нет (проверить подпись) | Webhook от Rahmat |
| GET | `/api/v1/payments/history` | JWT | История платежей пользователя |
| GET | `/api/v1/subscriptions/active` | JWT | Активная подписка |
| POST | `/api/v1/listings/:type/:id/boost` | JWT + Owner | Запустить Boost |
| POST | `/api/v1/admin/payments/topup` | SuperAdmin | Ручное пополнение токенов |
| GET | `/api/v1/admin/pricing` | SuperAdmin | Все тарифы |
| PUT | `/api/v1/admin/pricing/:key` | SuperAdmin | Обновить тариф |

**POST /payments/create**
- Body: `{ "pricing_key": "tokens_basic", "currency": "UZS" }`
- Создать payment_order
- Отправить запрос в Rahmat API
- Вернуть: `{ "payment_url": "https://rahmat.uz/pay/...", "order_id": "uuid" }`

**POST /payments/webhook**
- Обработка идемпотентна: если payment_order уже paid → вернуть 200, не обрабатывать повторно
- При ошибке обработки — вернуть 500 (Rahmat будет повторять)

**Подписка: уведомление за 3 дня**
- Фоновая задача раз в сутки: subscriptions WHERE expires_at < now() + 3 days AND is_active = true
- Создать notification типа `subscription_expiring`

### Бизнес-правила
- Нельзя купить boost для чужого объявления
- Если у объявления уже есть активный boost — продлевать, не перезаписывать
- При покупке подписки — существующая не отменяется, новая добавляет дни сверху (или стартует после окончания текущей — уточнить)
- payment_orders.status меняется только webhook-ом, не вручную через API (кроме супер-админа)

---

## ДЕНЬ 10 — Поиск, маршруты, избранное, уведомления, Супер-админ

**Цель:** Финальные модули. Полноценная панель супер-админа. Postman-коллекция готова.

### Задачи

1. **Единый поиск**
   - Один endpoint принимает тип (cargo/transport/warehouse) и параметры
   - Автодополнение городов: отдельный endpoint для подсказок

2. **Сохранённые маршруты**
   - Перевозчик сохраняет пару from/to
   - При создании нового cargo_listing → проверить все saved_routes, найти совпадения → создать notifications

3. **Избранное**
   - Добавить/убрать объявление
   - В ответе списка избранного — проверять, активно ли объявление ещё

4. **Уведомления**
   - Получить список (с пагинацией, сначала непрочитанные)
   - Пометить прочитанным (одно или все)
   - Счётчик непрочитанных: `GET /notifications/unread-count`

5. **Супер-админ панель**
   - Все разделы из ТЗ раздела 13.2

### API Endpoints — Поиск

| Метод | Путь | Auth | Описание |
|-------|------|------|----------|
| GET | `/api/v1/search` | Нет | Общий поиск |
| GET | `/api/v1/search/cities` | Нет | Автодополнение городов |

**GET /search**
- Query: `type=cargo&from_city=Ташкент&to_city=Москва&...` (все фильтры из разделов 9.2–9.4)
- Использует те же фильтры что и /listings/cargo, /listings/transport, /warehouses

**GET /search/cities**
- Query: `q=Таш` → возвращает список: `["Ташкент", "Ташкентская область"]`
- Источник: статичный список городов СНГ в коде (JSON-файл) или таблица в БД

---

### API Endpoints — Маршруты

| Метод | Путь | Auth | Описание |
|-------|------|------|----------|
| POST | `/api/v1/routes` | JWT | Сохранить маршрут |
| GET | `/api/v1/routes` | JWT | Мои маршруты |
| DELETE | `/api/v1/routes/:id` | JWT + Owner | Удалить |
| PATCH | `/api/v1/routes/:id/notifications` | JWT + Owner | Вкл/выкл уведомления |

### Бизнес-правила маршрутов
- При создании груза (POST /listings/cargo) — в конце транзакции запускать поиск совпадений по saved_routes
- Совпадение: from_city объявления ILIKE from_city маршрута (и аналогично to_city)
- Уведомление создаётся только если notifications_enabled = true для маршрута

---

### API Endpoints — Избранное

| Метод | Путь | Auth | Описание |
|-------|------|------|----------|
| POST | `/api/v1/favorites` | JWT | Добавить в избранное |
| DELETE | `/api/v1/favorites` | JWT | Убрать из избранного |
| GET | `/api/v1/favorites` | JWT | Мои избранные |

**POST /favorites**
- Body: `{ "listing_type": "cargo", "listing_id": "uuid" }`
- Если уже в избранном — вернуть 200 (идемпотентно)

**GET /favorites**
- Возвращает объявления с флагом `is_expired: true` если объявление архивировано

---

### API Endpoints — Уведомления

| Метод | Путь | Auth | Описание |
|-------|------|------|----------|
| GET | `/api/v1/notifications` | JWT | Список |
| GET | `/api/v1/notifications/unread-count` | JWT | Счётчик непрочитанных |
| PATCH | `/api/v1/notifications/:id/read` | JWT | Прочитать одно |
| PATCH | `/api/v1/notifications/read-all` | JWT | Прочитать все |

---

### API Endpoints — Супер-Админ

| Метод | Путь | Описание |
|-------|------|----------|
| GET | `/api/v1/admin/dashboard` | Метрики: пользователи, объявления, выручка за период |
| GET | `/api/v1/admin/users` | Список пользователей (поиск, фильтры) |
| GET | `/api/v1/admin/users/:id` | Профиль пользователя (полный) |
| PATCH | `/api/v1/admin/users/:id/block` | Заблокировать / разблокировать |
| POST | `/api/v1/admin/users/:id/topup` | Начислить токены вручную |
| GET | `/api/v1/admin/listings` | Все объявления всех типов |
| DELETE | `/api/v1/admin/listings/:type/:id` | Удалить объявление |
| PATCH | `/api/v1/admin/listings/:type/:id/block` | Заблокировать объявление |
| GET | `/api/v1/admin/companies` | Все компании |
| GET | `/api/v1/admin/payments` | Все платежи |
| GET | `/api/v1/admin/pricing` | Все тарифы |
| PUT | `/api/v1/admin/pricing/:key` | Обновить тариф |
| POST | `/api/v1/admin/moderators` | Создать учётную запись модератора |
| DELETE | `/api/v1/admin/moderators/:id` | Удалить модератора |

**GET /admin/dashboard**
- Query: `period=7d / 30d / 90d`
- Возвращает: новые пользователи, новые объявления, выручка (UZS/USD), купленные контакты, активные подписки, срочные верификации (deadline < 2h)

---

### Финальный Postman

- Все 60+ endpoint-ов в одной коллекции
- Папки по дням/модулям: Auth, Users, Companies, Moderator, Cargo, Fleet, Transport, Warehouses, Contacts, Payments, Search, Routes, Favorites, Notifications, Admin
- Переменные: `base_url`, `access_token`, `admin_token`, `moderator_token`
- Pre-request Script для автообновления access_token через refresh
- Тесты на критичных запросах: verify-otp сохраняет токены, contact/view проверяет что contacts_bought_count вырос

---

# РАЗДЕЛ 5 — ФОНОВЫЕ ЗАДАЧИ

Реализовать как goroutine с ticker внутри приложения (или через pg_cron):

| Задача | Интервал | Логика |
|--------|---------|--------|
| Истечение объявлений | 30 минут | cargo/transport: expires_at < now() AND active → archived + notification |
| Истечение Boost | 1 час | boost_expires_at < now() AND is_boosted → is_boosted = false |
| Уведомление подписки | 1 раз в сутки в 09:00 | expires_at < now() + 3 days → notification sub_expiring |
| OTP очистка | 1 раз в сутки | Удалять otp_codes старше 24 часов |

---

# РАЗДЕЛ 6 — БЕЗОПАСНОСТЬ

| Правило | Реализация |
|---------|-----------|
| Только свои объявления | При PUT/DELETE проверять user_id = listing.user_id |
| ИНН через модератора | ИНН в ответе /companies скрыт для обычных пользователей |
| Телефон скрыт | В карточке объявления phone = null если нет contact_view |
| Гос. номер скрыт | fleet_vehicles.plate_number не возвращать неавторизованным |
| Webhook подпись | Проверять HMAC-подпись от Rahmat (документация Rahmat) |
| Rate limit OTP | Максимум 3 запроса OTP в час с одного IP / телефона |
| Блокировка | Заблокированный пользователь получает 403 на любой запрос |