# KARVON
## Stack: Go + Gin · PostgreSQL · Postman · Next.js (фронт) · OpenStreetMap (своя карта)
## MVP v3.2 · Июнь 2026
### Scope: Регистрация · Компании · **Товары (маркетплейс)** · Склады · Монетизация
> ⚠️ Транспорт и автопарк вынесены из текущего ТЗ (вернём отдельным этапом).

**Изменения v3.2:**
- **Грузы → карточка товара (маркетплейс):** наименование, категория, количество в наличии + мин. заказ, делимость, упаковка, **цена за единицу видна без покупки контакта**, фото/видео/сертификаты, статус «Нет в наличии», постоянная карточка (без таймера истечения).
- **Тарифы полностью динамические:** супер-админ через панель управляет тарифами/токенами/суммами (вкл. бонус за регистрацию `tokens_registration`); CRUD тарифов. Пакеты: Мини 5 / Базовый 20 / Стандарт 50 / Максимум 100 (значения редактируемы).
- **Своя карта:** self-hosted tileserver (OpenStreetMap), фронт берёт адрес из `GET /api/v1/config`.
- **Дашборд:** лента событий (кто купил контакт / добавил в избранное) — `GET /users/me/events`.
- **Склады:** контактное лицо обязательно, поддержка видео, цветовая кодировка типов в ответе (`color`).
- **Избранное:** автор видит кто добавил (имя + компания), без анонимизации.
- **Оплата:** Rahmat scaffold готов к подключению (заказ → webhook → активация токенов/подписки/Boost).

---

# РАЗДЕЛ 1 — АРХИТЕКТУРА

## Стек и библиотеки

| Слой | Технология |
|------|-----------|
| HTTP-фреймворк | Gin |
| База данных | PostgreSQL 15+ |
| ORM / query builder | GORM (AutoMigrate) + sqlx для сложных выборок |
| Аутентификация | JWT (access + refresh tokens) |
| OTP | Telegram Gateway API + WhatsApp (Baileys микросервис) |
| Платежи | Rahmat (webhook) |
| Файлы | Локальный storage (S3-совместимый интерфейс) |
| Конфиг | .env через godotenv |
| Логирование | zerolog |
| Контейнеры | Docker + Docker Compose |
| API-документация | Postman Collection |
| i18n | ru / uz / en (поле `message` во всех ответах) |

---

## Структура проекта

```
/cmd/api          — точка входа (main.go)
/internal
  /config         — загрузка .env
  /handler        — HTTP-хэндлеры (по модулям)
  /service        — бизнес-логика
  /repository     — запросы к БД
  /middleware     — JWT, роли, lang, rate limit
  /model          — структуры данных (Go structs / GORM models)
  /dto            — Request / Response объекты
/pkg
  /jwt            — генерация/валидация токенов
  /notifier       — Telegram / WhatsApp OTP
  /rahmat         — клиент Rahmat платежи
  /storage        — работа с файлами
  /i18n           — переводы сообщений
/docs/postman     — Postman-коллекция
/docker-compose.yml
/.env.example
```

---

## Переменные окружения (.env)

```
APP_PORT=8082
DB_URL=postgres://user:pass@localhost:5432/dbname?sslmode=disable
JWT_SECRET=...
JWT_ACCESS_TTL=15m
JWT_REFRESH_TTL=720h

TELEGRAM_GATEWAY_BASE_URL=https://gatewayapi.telegram.org
TELEGRAM_GATEWAY_TOKEN=...
TELEGRAM_GATEWAY_BYPASS=false

WHATSAPP_OTP_BASE_URL=http://whatsapp:3211
WHATSAPP_OTP_TOKEN=...
WHATSAPP_OTP_BYPASS=false

RAHMAT_MERCHANT_ID=...
RAHMAT_SECRET_KEY=...
STORAGE_TYPE=local
STORAGE_PATH=./uploads
```

---

## Формат всех API-ответов

Все ответы (успех и ошибка) ВСЕГДА содержат переведённое поле `message` на языке запроса (ru/uz/en).

**Успех:**
```json
{
  "success": true,
  "message": "Готово",
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

Язык: `?lang=ru|uz|en` или заголовок `Accept-Language`, по умолчанию `ru`.

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
| `INN_NOT_VERIFIED` | ИНН не прошёл автопроверку по реестру |
| `COMPANY_NOT_VERIFIED` | Компания не верифицирована |
| `INSUFFICIENT_TOKENS` | Нет токенов |
| `FREE_LISTING_USED` | Бесплатное объявление уже использовано |
| `PAYMENT_FAILED` | Ошибка платежа |
| `ALREADY_EXISTS` | Дубликат (например, компания с таким ИНН) |

---

# РАЗДЕЛ 2 — СХЕМА БАЗЫ ДАННЫХ

---

### users
Основной аккаунт. Одна запись на телефон.

| Колонка | Тип | Описание |
|---------|-----|----------|
| id | UUID PK | gen_random_uuid() |
| phone | VARCHAR(20) UNIQUE NOT NULL | Номер телефона |
| name | VARCHAR(100) | Имя / название |
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
Юрлица, привязанные к пользователю. Один user → много companies. Каждая верифицируется отдельно.

| Колонка | Тип | Обяз? | Описание |
|---------|-----|-------|----------|
| id | UUID PK | | |
| user_id | UUID FK → users | | Владелец |
| country | VARCHAR(100) NOT NULL | ✓ | Страна регистрации |
| org_type | ENUM: ooo / ao / ip / ltd / gmbh / co_ltd | ✓ | Тип организации |
| name | VARCHAR(200) NOT NULL | ✓ | Полное юридическое название |
| inn | VARCHAR(30) NOT NULL | ✓ | ИНН / БИН / VAT / USCC |
| inn_verified | BOOLEAN DEFAULT false | | Прошёл автопроверку по реестру страны |
| inn_verified_at | TIMESTAMPTZ | | Когда подтверждён ИНН |
| phone | VARCHAR(20) NOT NULL | ✓ | Телефон компании |
| email | VARCHAR(100) NOT NULL | ✓ | Email компании |
| city | VARCHAR(100) NOT NULL | ✓ | Город / район |
| region | VARCHAR(100) | — | Область / провинция |
| street | VARCHAR(200) NOT NULL | ✓ | Улица, дом, офис |
| postal_code | VARCHAR(20) | — | Почтовый индекс |
| reg_doc_url | TEXT NOT NULL | ✓ | Свидетельство о регистрации (PDF/JPG, до 10 МБ) |
| status | ENUM: pending / approved / rejected / docs_requested | | DEFAULT pending |
| rejection_reason | TEXT | | Причина отклонения (обязательна при reject) |
| docs_request_note | TEXT | | Текст запроса документов модератором |
| moderator_id | UUID FK → users | | Кто обрабатывал |
| verified_at | TIMESTAMPTZ | | Дата одобрения |
| created_at | TIMESTAMPTZ | | |
| updated_at | TIMESTAMPTZ | | |

**Типы организации (org_type):** ООО (`ooo`), АО (`ao`), ИП (`ip`), Ltd (`ltd`), GmbH (`gmbh`), Co. Ltd (`co_ltd`).

**Соответствие статусов UI:** `pending` = «На проверке», `docs_requested` = «Требует документов», `approved` = «Проверен», `rejected` = «Отклонено».

---

### cargo_listings — КАРТОЧКА ТОВАРА (маркетплейс-модель)
Грузовладелец публикует карточку товара (как в маркетплейсе): наименование, количество, цена, откуда забрать. Перевозчик находит товар и покупает контакт. Карточка **постоянная** (не истекает по таймеру). Только верифицированные компании. Цена видна в листинге **без** покупки контакта.

| Колонка | Тип | Описание |
|---------|-----|----------|
| id | UUID PK | |
| company_id | UUID FK → companies | |
| user_id | UUID FK → users | Публикатор |
| cargo_name | VARCHAR(200) NOT NULL | Наименование товара |
| category | ENUM cargo_category_enum | stroymat/food/textile/metal/chemical/wood/electronics/other |
| description | TEXT | Описание, состав, характеристики |
| quantity_available | DECIMAL | Всего в наличии |
| quantity_unit | ENUM: ton/places/pallet/m3 | Ед. измерения |
| min_order | DECIMAL | Минимальный заказ |
| min_order_unit | ENUM: ton/places/pallet/m3 | |
| divisibility | ENUM: ftl/ltl/dogruz | Делимость |
| packaging | ENUM: bulk/pallets/bags/barrels/rolls/boxes/liquid/oversized | Упаковка |
| weight_per_place | DECIMAL | Вес одного места/паллета |
| length_m / width_m / height_m | DECIMAL | Габариты |
| rate_mode (price_mode) | ENUM: negotiable/fixed/announcement | Режим цены |
| rate_amount (price_per_unit) | DECIMAL | Цена за единицу |
| rate_currency (currency) | ENUM: UZS/USD | |
| vat_mode | ENUM: yes/no/unspecified | Включает НДС |
| payment_terms | VARCHAR(50) | prepaid/on_fact/deferred |
| payment_method | VARCHAR(50) | cash/bank/agreement |
| from_country / from_city | VARCHAR(100) | Откуда забрать |
| pickup_address | TEXT | Адрес склада/завода |
| lat / lng | DECIMAL | Геометка (OpenStreetMap) |
| loading_methods | TEXT[] | self/loader/crane |
| working_hours | JSONB | Режим работы по дням |
| is_adr | BOOLEAN | Опасный груз |
| adr_class | INT (1–9) | Класс ADR |
| has_temp_regime | BOOLEAN | |
| temp_min / temp_max | DECIMAL | Температурный режим |
| body_types (required_body_types) | TEXT[] | Требования к кузову: tent/ref/container/cistern/carcarrier |
| incoterms | VARCHAR(20) | Базис поставки (для международных) |
| delivery_geography | TEXT[] | uz/cis/world |
| permits | TEXT[] | TIR/CMR/phyto/CITES/EKMT |
| notes (comment_for_carrier) | TEXT | Комментарий перевозчику |
| in_stock | BOOLEAN DEFAULT true | false = «Нет в наличии» |
| status | ENUM: active / archived | DEFAULT active (`completed` не используется) |
| is_template / template_name | | Шаблоны постоянных товаров |
| is_boosted / boost_expires_at | | Boost (продвижение) |
| is_paid | BOOLEAN | Платное/бесплатное объявление |
| views_count / contacts_bought_count | INT | Статистика |
| created_at / updated_at | TIMESTAMPTZ | |

**Медиа** (фото/видео/сертификаты, до 10) хранятся в `listing_media` (`entity_type='cargo'`). Контакты в карточке скрыты — открываются через `POST /contacts/view`.

> Поле `type` (domestic/international) и `duration_hours` сохранены в БД как legacy NOT NULL и заполняются дефолтами; в API не используются (карточка товара постоянная, без маршрута назначения).

---

### warehouse_listings
Карточки складов. Только от верифицированных компаний.

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
Фото и документы для объявлений.

| Колонка | Тип | Описание |
|---------|-----|----------|
| id | UUID PK | |
| entity_type | ENUM: cargo / warehouse | |
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
| listing_type | ENUM: cargo / warehouse | |
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
Сохранённые маршруты (отслеживание новых грузов).

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
| listing_type | ENUM: cargo / warehouse | |
| listing_id | UUID | |
| created_at | TIMESTAMPTZ | |

Уникальный индекс: (user_id, listing_type, listing_id)

---

### notifications
Внутренние уведомления платформы.

| Колонка | Тип | Описание |
|---------|-----|----------|
| id | UUID PK | |
| user_id | UUID FK → users | Получатель |
| type | VARCHAR(50) | new_cargo_on_route / contact_purchased / sub_expiring / verification_approved / verification_rejected / docs_requested / etc. |
| title | VARCHAR(200) | |
| body | TEXT | |
| is_read | BOOLEAN DEFAULT false | |
| meta | JSONB | Дополнительные данные (listing_id, company_id и т.д.) |
| created_at | TIMESTAMPTZ | |

---

# РАЗДЕЛ 3 — MIDDLEWARE

| Middleware | Назначение |
|-----------|-----------|
| `AuthMiddleware` | Проверяет JWT, кладёт user в context |
| `RoleMiddleware(roles...)` | Проверяет role: user / moderator / super_admin |
| `CompanyVerifiedMiddleware` | Проверяет, есть ли у user хоть одна verified company |
| `LangMiddleware` | Определяет язык запроса (query/header), кладёт в context |
| `RateLimitMiddleware` | Ограничение запросов (особенно для /auth/send-otp) |
| `CORSMiddleware` | Разрешённые origins |
| `LoggerMiddleware` | Логирование запросов |

---

# РАЗДЕЛ 4 — ПЛАН РАЗРАБОТКИ (9 ДНЕЙ)

> Транспорт/автопарк исключены. Поток: Старт → Auth → Профиль+Компании → Модерация → Грузы → Склады → Токены/Контакты → Платежи → Поиск/Маршруты/Избранное/Уведомления/Админ.

---

## ДЕНЬ 1 — Инициализация проекта и база данных

**Цель:** Рабочий скелет, все таблицы созданы, проект стартует.

### Задачи

1. **Go-проект** — `go mod init karvon`; зависимости: gin, gorm, pgx/lib-pq, godotenv, zerolog, uuid.
2. **Docker Compose** — PostgreSQL 15, приложение, WhatsApp-микросервис; порты, volumes.
3. **Конфигурация** — загрузка `.env`, структура Config с валидацией обязательных полей (fail fast).
4. **Миграции (GORM AutoMigrate)** — порядок: users → otp_codes → refresh_tokens → companies → cargo_listings → cargo_waypoints → warehouse_listings → listing_media → contact_views → token_transactions → payment_orders → subscriptions → pricing_config → saved_routes → favorites → notifications. Seed `pricing_config` всеми тарифами с нулевыми ценами.
5. **HTTP-сервер** — Gin на APP_PORT; `GET /health` → `{"status":"ok"}`; ping PostgreSQL.
6. **Postman** — Collection `KARVON API v1`; переменные `base_url`, `access_token`, `refresh_token`; запрос `GET /health`.

### Бизнес-правила
- БД недоступна при старте → завершение с ошибкой (fail fast).
- Миграции запускаются автоматически при старте.

---

## ДЕНЬ 2 — Регистрация и аутентификация (OTP + JWT)

**Цель:** Регистрация в три шага: телефон → OTP → имя. Вход и регистрация — единый поток.

### Регистрация — три шага

| Шаг | Действие | Детали |
|-----|----------|--------|
| 1 — Телефон | Ввод номера | Код страны + номер. Поддержка УЗ (+998), СНГ, ЕС, Китая. |
| 2 — OTP код | Ввод 6-значного кода | Отправка через **Telegram Gateway / WhatsApp**. Действует 5 минут. Повторная отправка — таймер 60 сек. |
| 3 — Имя | Ввод имени или названия | Отображается в профиле. После — аккаунт создан, начислено **5 токенов** автоматически. |

Ролей при регистрации нет. Один аккаунт — все возможности. Верификация компании — отдельный шаг в профиле.

### Задачи

1. **Отправка OTP (Telegram/WhatsApp)** — выбор канала; код извлекается из шаблона сообщения; режим bypass для отладки (логирует код, не отправляет).
2. **Генерация OTP** — 6-значный код; запись в `otp_codes`, TTL 5 минут; старые коды телефона → used=true.
3. **Rate limit** — макс. 3 активных запроса OTP в час с телефона; иначе `OTP_MAX_ATTEMPTS` + сколько секунд ждать.
4. **JWT** — access (user_id, role, 15 мин), refresh (user_id, 30 дней, хранится в `refresh_tokens`); refresh выдаёт новую пару.
5. **AuthMiddleware** — читает Bearer, валидирует, достаёт user; заблокированный → 403.

### API Endpoints

| Метод | Путь | Auth | Описание |
|-------|------|------|----------|
| POST | `/api/v1/auth/send-otp` | Нет | Отправить OTP |
| POST | `/api/v1/auth/verify-otp` | Нет | Проверить OTP, получить токены |
| POST | `/api/v1/auth/refresh` | Нет | Обновить access token |
| POST | `/api/v1/auth/logout` | JWT | Инвалидировать refresh token |
| POST | `/api/v1/auth/complete-registration` | JWT | Шаг 3: ввод имени |

**POST /auth/send-otp** — `{ "phone": "+998901234567" }` → `{ "expires_in": 300 }`. Если пользователя нет — создаётся запись с пустым именем.

**POST /auth/verify-otp** — `{ "phone", "code" }` → `{ access_token, refresh_token, is_new_user }`. Новому пользователю начислить 5 токенов. Неверный код → attempts++, при ≥5 код блокируется.

**POST /auth/complete-registration** — `{ "name": "Алишер" }`. Обновляет users.name. Доступен только новым пользователям без имени.

### Бизнес-правила
- Телефон нормализовать (убрать пробелы, скобки, дефисы).
- OTP валиден максимум 5 минут; после успеха → used=true.
- Существующий пользователь — это вход; токены выдаются в обоих случаях.

---

## ДЕНЬ 3 — Профиль и компании (верификация)

**Цель:** Пользователь управляет профилем и добавляет юрлица с полной формой верификации.

### Точка входа
Профиль → «Добавить компанию». Один пользователь может добавить несколько компаний; каждая верифицируется отдельно; переключение между компаниями — в личном кабинете.

### Форма верификации — поля

| Группа | Поле | Обяз? |
|--------|------|-------|
| Компания | Страна регистрации | ✓ |
| Компания | Тип организации (ООО, АО, ИП, Ltd, GmbH, Co. Ltd) | ✓ |
| Компания | Полное юридическое название | ✓ |
| Компания | ИНН / БИН / VAT / USCC — автопроверка по реестру страны | ✓ |
| Контакты | Телефон компании | ✓ |
| Контакты | Email компании | ✓ |
| Адрес | Город / район | ✓ |
| Адрес | Область / провинция | — |
| Адрес | Улица, дом, офис | ✓ |
| Адрес | Почтовый индекс | — |
| Документ | Свидетельство о регистрации (PDF/JPG, до 10 МБ) | ✓ |

Кнопка «Отправить на проверку» разблокируется только после: **ИНН подтверждён автопроверкой** + заполнены все обязательные поля + загружен документ.

### Автопроверка ИНН по стране

| Страна | Реестр |
|--------|--------|
| Узбекистан | tax.gov.uz |
| Россия | egrul.nalog.ru |
| Казахстан | egov.kz |
| Европа (ЕС) | VIES (ec.europa.eu) |
| Китай | gsxt.gov.cn |

Реализация: отдельный endpoint автопроверки по стране и номеру; результат пишется в `inn_verified` + `inn_verified_at`. Если реестр недоступен — компания подаётся с `inn_verified=false`, финальное решение за модератором.

### Задачи

1. **Профиль** — get/update личных данных: name, email, whatsapp, telegram, city, country.
2. **Загрузка файлов** — multipart/form-data; PDF и изображения, до 10 МБ; возвращает URL.
3. **Автопроверка ИНН** — endpoint проверки по реестру страны перед отправкой формы.
4. **Компании** — создание/чтение/обновление; один user → несколько компаний; одно бесплатное объявление на весь аккаунт (по user_id).

### API Endpoints

| Метод | Путь | Auth | Описание |
|-------|------|------|----------|
| GET | `/api/v1/users/me` | JWT | Мой профиль |
| PUT | `/api/v1/users/me` | JWT | Обновить профиль |
| GET | `/api/v1/users/me/stats` | JWT | Статистика (токены, подписка, активные объявления) |
| POST | `/api/v1/companies/verify-inn` | JWT | Автопроверка ИНН по стране |
| POST | `/api/v1/companies` | JWT | Добавить компанию (отправить на проверку) |
| GET | `/api/v1/companies` | JWT | Мои компании |
| GET | `/api/v1/companies/:id` | JWT | Одна компания |
| PUT | `/api/v1/companies/:id` | JWT | Обновить данные компании |
| POST | `/api/v1/upload` | JWT | Загрузить файл → вернуть URL |

**POST /companies/verify-inn** — `{ "country": "UZ", "inn": "123456789" }` → `{ "verified": true, "name": "...", "registry": "tax.gov.uz" }` либо `INN_NOT_VERIFIED`.

**POST /companies** — `{ country, org_type, name, inn, phone, email, city, region?, street, postal_code?, reg_doc_url }`. Требуется `inn_verified=true`. Статус по умолчанию `pending`. Создаётся уведомление модераторам.

**GET /companies** — все компании пользователя: name, inn, status, rejection_reason, docs_request_note, verified_at.

### Бизнес-правила
- ИНН уникален на платформе: если такой ИНН уже у другой verified-компании → `ALREADY_EXISTS`.
- Нельзя редактировать компанию в статусе `pending`.
- В статусе `docs_requested` — пользователь может обновить документ и поля и повторно отправить.
- ИНН в ответе `/companies` скрыт для чужих пользователей.

---

## ДЕНЬ 4 — Верификация компаний (Модератор)

**Цель:** Модератор обрабатывает заявки за 24 часа.

### Статусы заявки

| Статус | Описание |
|--------|----------|
| На проверке (`pending`) | Заявка подана, ожидает решения. |
| Требует документов (`docs_requested`) | Модератор запросил доп. документы; пользователь получает уведомление. |
| Проверен (`approved`) | Одобрено. Бейдж «✓ Проверен» + право публиковать грузы и склады. |
| Отклонено (`rejected`) | Отказ с обязательной причиной; пользователь видит причину и может подать повторно. |

### Действия модератора

- **Одобрить** → `approved`, `verified_at=now()`, пользователь получает **push + email**.
- **Отклонить** → `rejected`, `rejection_reason` обязателен (текст, есть шаблоны), **push + email**.
- **Запросить документы** → `docs_requested`, `docs_request_note` (текст + шаблоны), **push + email**.

### Задачи

1. **Ролевой middleware** — `RoleMiddleware("moderator","super_admin")` на /moderator/*.
2. **Очередь верификации** — заявки `pending`/`docs_requested`, сортировка по дате; `deadline=created_at+24ч`; `is_urgent=true` если до дедлайна <2ч.
3. **Обработка** — approve / reject (reason обяз.) / request-docs (note обяз.).
4. **Уведомления** — при любом изменении статуса: запись в `notifications` + push + email.

### API Endpoints

| Метод | Путь | Auth | Описание |
|-------|------|------|----------|
| GET | `/api/v1/moderator/queue` | Moderator | Очередь заявок |
| GET | `/api/v1/moderator/queue/:id` | Moderator | Детали заявки |
| POST | `/api/v1/moderator/queue/:id/approve` | Moderator | Одобрить |
| POST | `/api/v1/moderator/queue/:id/reject` | Moderator | Отклонить (reason обязателен) |
| POST | `/api/v1/moderator/queue/:id/request-docs` | Moderator | Запросить документы (note обязателен) |
| GET | `/api/v1/moderator/history` | Moderator | История своих решений |

**GET /moderator/queue** — `status=pending` (по умолч.); возвращает id, company name, inn, org_type, country, created_at, deadline, is_urgent, user_name, user_phone.

**POST /moderator/queue/:id/reject** — `{ "reason": "Документы не читаемы..." }` (обязателен).

**POST /moderator/queue/:id/request-docs** — `{ "note": "Загрузите свидетельство лучшего качества" }` (обязателен).

### Бизнес-правила
- Модератор видит ТОЛЬКО очередь верификации (никаких платежей/статистики/пользователей).
- После одобрения компании — публикация грузов и складов без доп. модерации.
- Один модератор не может обработать заявку дважды (только первый успешный action).

---

## ДЕНЬ 5 — Объявления о грузах

**Цель:** Верифицированные компании публикуют грузы. Работают фильтры.

**Модель:** карточка ТОВАРА (см. схему `cargo_listings`). Карточка постоянная — без таймера истечения; «снимается» вручную (status=archived) или помечается «Нет в наличии» (in_stock=false).

### Задачи

1. **Проверка перед публикацией** — у пользователя ≥1 verified компания.
2. **Создание** — форма товара (категория, количество+ед., мин. заказ, делимость, упаковка, цена за ед., геометка, ADR/класс, темп. режим, требования к кузову, медиа). Медиа (фото/видео/сертификаты) — в `listing_media`.
3. **Цена видна в листинге** без покупки контакта; скрыты только контакты компании.
4. **Шаблоны / дублирование** — для постоянных товаров.

### API Endpoints

| Метод | Путь | Auth | Описание |
|-------|------|------|----------|
| POST | `/api/v1/listings/cargo` | JWT + Verified | Создать карточку товара |
| GET | `/api/v1/listings/cargo` | Нет | Листинг с фильтрами |
| GET | `/api/v1/listings/cargo/:id` | Нет | Карточка |
| PUT | `/api/v1/listings/cargo/:id` | JWT + Owner | Редактировать |
| PATCH | `/api/v1/listings/cargo/:id/status` | JWT + Owner | Статус (active/archived) и/или in_stock |
| POST | `/api/v1/listings/cargo/:id/duplicate` | JWT + Verified | Дублировать |
| DELETE | `/api/v1/listings/cargo/:id` | JWT + Owner | Удалить |
| GET | `/api/v1/listings/cargo/:id/stats` | JWT + Owner | Статистика |
| GET | `/api/v1/listings/cargo/templates` | JWT | Мои шаблоны |
| POST | `/api/v1/listings/cargo/:id/template` | JWT + Owner | Сохранить как шаблон |
| POST | `/api/v1/listings/cargo/from-template/:id` | JWT + Verified | Создать из шаблона |

**GET /listings/cargo — фильтры:**
```
category, from_city, from_country, divisibility (ftl/ltl/dogruz),
packaging, min_order_max, qty_min, qty_max, price_min, price_max,
temp_regime (bool), adr (bool), verified_only (bool),
sort (newest/price_asc/price_desc/quantity), page, per_page
```

**GET /listings/cargo/:id:** контакты компании скрыты (открываются через `POST /contacts/view`); цена за единицу и объём видны сразу.

**GET /listings/cargo/:id/stats:** views_count, contacts_bought_count, список добавивших в избранное (имя + компания + дата).

### Бизнес-правила
- Boosted карточки вверху (is_boosted DESC, created_at DESC).
- Карточка постоянная; «Нет в наличии» = in_stock=false (видна, но помечена).
- При дублировании копируются все поля и медиа; сбрасываются статистика/boost.

---

## ДЕНЬ 6 — Склады

**Цель:** Три типа складов с разными полями и картой.

### Задачи

1. **CRUD складов** — только verified компании; каждый тип (regular/cold/customs) имеет свои обязательные поля; геометка lat/lng.
2. **Валидация по типу** — `cold`: temp_min/temp_max/cold_chamber_types обязательны; `customs`: customs_license_* обязательны; `regular`: только общие.
3. **Прогресс-бар занятости** — `occupancy_percent = ((area_total_m2 - area_free_m2)/area_total_m2)*100`.

### API Endpoints

| Метод | Путь | Auth | Описание |
|-------|------|------|----------|
| POST | `/api/v1/warehouses` | JWT + Verified | Добавить склад |
| GET | `/api/v1/warehouses` | Нет | Список складов |
| GET | `/api/v1/warehouses/:id` | Нет | Карточка |
| PUT | `/api/v1/warehouses/:id` | JWT + Owner | Редактировать |
| PATCH | `/api/v1/warehouses/:id/status` | JWT + Owner | Архивировать |
| DELETE | `/api/v1/warehouses/:id` | JWT + Owner | Удалить |

**GET /warehouses — фильтры:**
```
region, warehouse_type (regular/cold/customs),
area_min, area_max, temp_min, temp_max (только cold),
services (массив: crossdock/railway/parking_trucks),
sort, page, per_page
```

### Бизнес-правила
- Свободная площадь ≤ общей площади.
- При фильтре по температуре — cold склад подходит если temp в [temp_min, temp_max].
- Контакты склада скрыты так же, как у грузов (через токен/подписку).

---

## ДЕНЬ 7 — Токены и просмотр контактов

**Цель:** Основная монетизация. Пользователь платит токен чтобы увидеть телефон.

### Логика просмотра контакта — сценарии

| Сценарий | Действие |
|----------|----------|
| Гость (не авторизован) | 401, фронт редиректит на вход |
| Авторизован, нет токенов, нет подписки | 402 + данные тарифов (пакеты + подписки) |
| Есть активная подписка | Показать контакт, токен НЕ списывать; contact_views (tokens_spent=0) |
| Уже смотрел (за 30 дней) | Показать контакт, токен НЕ списывать |
| Есть токены, нет подписки, первый просмотр | Списать 1 токен, показать контакт |
| Своё объявление | Показать контакт, токен НЕ списывать |

### Задачи

1. **Атомарность** — списание токена и запись contact_views в одной транзакции; при сбое контакт не показывать.
2. **История контактов** — список за 30 дней с краткой карточкой объявления.

### API Endpoints

| Метод | Путь | Auth | Описание |
|-------|------|------|----------|
| POST | `/api/v1/contacts/view` | JWT | Открыть контакт |
| GET | `/api/v1/contacts/history` | JWT | История просмотров |
| GET | `/api/v1/users/me/tokens` | JWT | Баланс и история токенов |

**POST /contacts/view** — `{ "listing_type": "cargo|warehouse", "listing_id": "uuid" }`.
- Успех: `{ phone, phone_extra, whatsapp, telegram, email, company_name, city, country }`.
- 402: `{ "error": {"code":"INSUFFICIENT_TOKENS"}, "data": {"packages":[...], "subscriptions":[...]} }`.

### Бизнес-правила
- `contacts_bought_count` инкрементируется только при реальном списании токена (не при подписке/повторе).
- Уведомление автору при покупке контакта: notification типа `contact_purchased` с именем/компанией покупателя.

---

## ДЕНЬ 8 — Монетизация (Rahmat, подписка, Boost)

**Цель:** Полный платёжный цикл. Webhook Rahmat активирует токены/подписки автоматически.

### Задачи

1. **Тарифы** — все активные из `pricing_config`; никаких хардкодов.
2. **Создание платежа** — payment_order `pending`; запрос в Rahmat (сумма, merchant_id, callback, metadata); вернуть URL платёжной страницы.
3. **Webhook** — POST `/payments/webhook`; проверка подписи; status=paid; по `payment_type`:
   - `tokens` → начислить токены (+token_transaction);
   - `subscription` → создать subscription с датами;
   - `listing` → пометить объявление is_paid=true;
   - `boost` → is_boosted=true, boost_expires_at=now()+duration.
4. **Boost истекает** — фон каждый час: boost_expires_at<now() → is_boosted=false.
5. **Ручное пополнение (наличные)** — только супер-админ; token_transaction type=credit, reason=manual.
6. **Управление ценами** — супер-админ читает/обновляет pricing_config (лог: updated_by, updated_at).

### API Endpoints

| Метод | Путь | Auth | Описание |
|-------|------|------|----------|
| GET | `/api/v1/payments/packages` | Нет | Все тарифы |
| POST | `/api/v1/payments/create` | JWT | Создать платёж → URL Rahmat |
| POST | `/api/v1/payments/webhook` | Нет (подпись) | Webhook от Rahmat |
| GET | `/api/v1/payments/history` | JWT | История платежей |
| GET | `/api/v1/subscriptions/active` | JWT | Активная подписка |
| POST | `/api/v1/listings/:type/:id/boost` | JWT + Owner | Запустить Boost (type: cargo/warehouse) |
| POST | `/api/v1/admin/payments/topup` | SuperAdmin | Ручное пополнение токенов |
| GET | `/api/v1/admin/pricing` | SuperAdmin | Все тарифы |
| PUT | `/api/v1/admin/pricing/:key` | SuperAdmin | Обновить тариф |

**POST /payments/create** — `{ "pricing_key": "tokens_basic", "currency": "UZS" }` → `{ payment_url, order_id }`.

**POST /payments/webhook** — идемпотентен (paid → 200 без повторной обработки); при ошибке → 500 (Rahmat повторит).

**Подписка: уведомление за 3 дня** — фон раз в сутки: expires_at<now()+3d AND is_active → notification `sub_expiring`.

### Бизнес-правила
- Нельзя купить boost для чужого объявления.
- Активный boost — продлевать, не перезаписывать.
- payment_orders.status меняется только webhook-ом (кроме супер-админа).

---

## ДЕНЬ 9 — Поиск, маршруты, избранное, уведомления, Супер-админ

**Цель:** Финальные модули и панель супер-админа. Postman готов.

### Поиск

| Метод | Путь | Auth | Описание |
|-------|------|------|----------|
| GET | `/api/v1/search` | Нет | Общий поиск (type=cargo/warehouse) |
| GET | `/api/v1/search/cities` | Нет | Автодополнение городов |

**GET /search** — `type=cargo|warehouse` + соответствующие фильтры (из Дней 5/6).
**GET /search/cities** — `q=Таш` → `["Ташкент","Ташкентская область"]` (статичный список СНГ/ЕС/Китай).

### Маршруты

| Метод | Путь | Auth | Описание |
|-------|------|------|----------|
| POST | `/api/v1/routes` | JWT | Сохранить маршрут |
| GET | `/api/v1/routes` | JWT | Мои маршруты |
| DELETE | `/api/v1/routes/:id` | JWT + Owner | Удалить |
| PATCH | `/api/v1/routes/:id/notifications` | JWT + Owner | Вкл/выкл уведомления |

- При создании груза — в конце транзакции искать совпадения по saved_routes (from_city/to_city ILIKE) → notifications (если notifications_enabled).

### Избранное

| Метод | Путь | Auth | Описание |
|-------|------|------|----------|
| POST | `/api/v1/favorites` | JWT | Добавить (listing_type: cargo/warehouse) |
| DELETE | `/api/v1/favorites` | JWT | Убрать |
| GET | `/api/v1/favorites` | JWT | Мои избранные |

- POST идемпотентен; GET помечает `is_expired:true` для архивных объявлений.

### Уведомления

| Метод | Путь | Auth | Описание |
|-------|------|------|----------|
| GET | `/api/v1/notifications` | JWT | Список (пагинация, сначала непрочитанные) |
| GET | `/api/v1/notifications/unread-count` | JWT | Счётчик непрочитанных |
| PATCH | `/api/v1/notifications/:id/read` | JWT | Прочитать одно |
| PATCH | `/api/v1/notifications/read-all` | JWT | Прочитать все |

### Супер-админ

| Метод | Путь | Описание |
|-------|------|----------|
| GET | `/api/v1/admin/dashboard` | Метрики: пользователи, объявления, выручка за период |
| GET | `/api/v1/admin/users` | Список пользователей |
| GET | `/api/v1/admin/users/:id` | Профиль пользователя |
| PATCH | `/api/v1/admin/users/:id/block` | Блокировать / разблокировать |
| POST | `/api/v1/admin/users/:id/topup` | Начислить токены вручную |
| GET | `/api/v1/admin/listings` | Все объявления (cargo/warehouse) |
| DELETE | `/api/v1/admin/listings/:type/:id` | Удалить объявление |
| PATCH | `/api/v1/admin/listings/:type/:id/block` | Заблокировать объявление |
| GET | `/api/v1/admin/companies` | Все компании |
| GET | `/api/v1/admin/payments` | Все платежи |
| GET | `/api/v1/admin/pricing` | Все тарифы |
| PUT | `/api/v1/admin/pricing/:key` | Обновить тариф |
| POST | `/api/v1/admin/moderators` | Создать модератора |
| DELETE | `/api/v1/admin/moderators/:id` | Удалить модератора |

**GET /admin/dashboard** — `period=7d/30d/90d`: новые пользователи, новые объявления, выручка (UZS/USD), купленные контакты, активные подписки, срочные верификации (deadline<2h).

### Финальный Postman
- Все endpoint-ы в одной коллекции; папки: Auth, Users, Companies, Moderator, Cargo, Warehouses, Contacts, Payments, Search, Routes, Favorites, Notifications, Admin.
- Переменные: `base_url`, `access_token`, `admin_token`, `moderator_token`.
- Pre-request Script для автообновления access_token через refresh.
- Тесты: verify-otp сохраняет токены; contact/view проверяет рост contacts_bought_count.

---

# РАЗДЕЛ 5 — ФОНОВЫЕ ЗАДАЧИ

Реализовать как goroutine с ticker внутри приложения (или pg_cron):

| Задача | Интервал | Логика |
|--------|---------|--------|
| Истечение объявлений | 30 минут | cargo: expires_at<now() AND active → archived + notification |
| Истечение Boost | 1 час | boost_expires_at<now() AND is_boosted → is_boosted=false |
| Уведомление подписки | 1 раз в сутки в 09:00 | expires_at<now()+3 days → notification sub_expiring |
| OTP очистка | 1 раз в сутки | Удалять otp_codes старше 24 часов |

---

# РАЗДЕЛ 6 — БЕЗОПАСНОСТЬ

| Правило | Реализация |
|---------|-----------|
| Только свои объявления | При PUT/DELETE проверять user_id = listing.user_id |
| ИНН через модератора | ИНН в ответе /companies скрыт для чужих пользователей |
| Телефон скрыт | В карточке объявления phone=null если нет contact_view |
| Webhook подпись | Проверять HMAC-подпись от Rahmat |
| Rate limit OTP | Максимум 3 запроса OTP в час с телефона |
| Блокировка | Заблокированный пользователь получает 403 на любой запрос |
