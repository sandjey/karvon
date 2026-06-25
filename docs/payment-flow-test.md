# KARVON — Payment Flow Documentation

**Base URL:** `https://api.centraltrademarket.com/api/v1`  
**Payment Gateway:** Multicard / Rahmat (UZS only)  
**Test OTP bypass:** `136092` (any phone)  
**Test phone:** `+998994878461`

---

## Шаг 0 — Аутентификация

Все платёжные эндпоинты (кроме `GET /payments/packages` и `POST /payments/webhook`) требуют Bearer токен.

### 1. Отправить OTP

```bash
curl -X POST https://api.centraltrademarket.com/api/v1/auth/send-otp \
  -H "Content-Type: application/json" \
  -d '{"phone":"+998994878461"}'
```

**Ответ:**
```json
{
  "success": true,
  "message": "OTP код отправлен"
}
```

### 2. Подтвердить OTP → получить токен

```bash
curl -X POST https://api.centraltrademarket.com/api/v1/auth/verify-otp \
  -H "Content-Type: application/json" \
  -d '{"phone":"+998994878461","code":"136092"}'
```

**Ответ:**
```json
{
  "success": true,
  "message": "Успешная авторизация",
  "data": {
    "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "token_type": "Bearer",
    "expires_in": 86400
  }
}
```

Сохранить `access_token` для дальнейших запросов:
```bash
TOKEN="eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
```

---

## Шаг 1 — Список тарифов (публичный, без авторизации)

```bash
curl https://api.centraltrademarket.com/api/v1/payments/packages
```

**Ответ:**
```json
{
  "success": true,
  "data": {
    "packages": [
      {"key": "tokens_mini",          "label": "Мини пакет (10 токенов)",       "price_uzs": 15000,  "price_usd": 1.5, "tokens_amount": 10,  "duration_days": 0},
      {"key": "tokens_standard",      "label": "Стандарт пакет (60 токенов)",   "price_uzs": 100000, "price_usd": 9,   "tokens_amount": 60,  "duration_days": 0},
      {"key": "tokens_premium",       "label": "Премиум пакет (50 токенов)",    "price_uzs": 150000, "price_usd": 12,  "tokens_amount": 50,  "duration_days": 0},
      {"key": "tokens_max",           "label": "Макс пакет (150 токенов)",      "price_uzs": 180000, "price_usd": 16,  "tokens_amount": 150, "duration_days": 0},
      {"key": "tokens_registration",  "label": "Бонус при регистрации",         "price_uzs": 0,      "price_usd": 0,   "tokens_amount": 5,   "duration_days": 0}
    ],
    "subscriptions": [
      {"key": "sub_week",       "label": "Подписка на неделю",      "price_uzs": 50000,   "price_usd": 5,  "duration_days": 7},
      {"key": "sub_month",      "label": "Подписка на месяц",       "price_uzs": 150000,  "price_usd": 13, "duration_days": 31, "tokens_amount": 100},
      {"key": "sub_pro_month",  "label": "Про подписка — 1 месяц",  "price_uzs": 280000,  "price_usd": 22, "duration_days": 30},
      {"key": "sub_year",       "label": "Подписка на год",         "price_uzs": 1000000, "price_usd": 90, "duration_days": 365}
    ],
    "listings": [
      {"key": "listing_cargo",           "label": "Объявление груза",               "price_uzs": 30000, "price_usd": 2.5, "duration_days": 30},
      {"key": "listing_paid",            "label": "Платное объявление (груз)",      "price_uzs": 30000, "price_usd": 3,   "duration_days": 0},
      {"key": "listing_warehouse",       "label": "Платное объявление (склад)",     "price_uzs": 30000, "price_usd": 3,   "duration_days": 0},
      {"key": "listing_warehouse_fridge","label": "Объявление склада (холодильник)","price_uzs": 50000, "price_usd": 4,   "duration_days": 30}
    ],
    "boosts": [
      {"key": "boost_1day", "label": "Буст — 1 день (груз / склад)",  "price_uzs": 20000,  "price_usd": 2,   "duration_days": 1},
      {"key": "boost_3day", "label": "Буст — 3 дня (груз / склад)",   "price_uzs": 50000,  "price_usd": 4.5, "duration_days": 3},
      {"key": "boost_7day", "label": "Буст — 7 дней (груз / склад)",  "price_uzs": 100000, "price_usd": 9,   "duration_days": 7}
    ]
  }
}
```

> **Заметка:** `tokens_registration` имеет `price_uzs=0` — его нельзя оплатить через Multicard (вернёт 400 TARIFF_NOT_PAYABLE).

---

## Поток A — Покупка токенов

**Endpoint:** `POST /payments/create`  
**Auth:** Bearer токен  
**payment_type:** `"tokens"`

### A1 — Мини пакет (10 токенов / 15 000 UZS)

```bash
curl -X POST https://api.centraltrademarket.com/api/v1/payments/create \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "payment_type": "tokens",
    "pricing_key": "tokens_mini",
    "currency": "UZS",
    "return_url": "https://karvon.uz/payment/success"
  }'
```

**Ответ (201):**
```json
{
  "success": true,
  "message": "Платёж создан",
  "data": {
    "order_id": "d63eda12-b270-4ff6-aa52-1c9e939de5cf",
    "payment_url": "https://dev-app.rhmt.uz/invoice/d3511d38-10f6-4246-8285-11eba2bf94b6",
    "amount": 15000,
    "currency": "UZS"
  }
}
```

### A2 — Стандарт пакет (60 токенов / 100 000 UZS)

```bash
curl -X POST https://api.centraltrademarket.com/api/v1/payments/create \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "payment_type": "tokens",
    "pricing_key": "tokens_standard",
    "currency": "UZS",
    "return_url": "https://karvon.uz/payment/success"
  }'
```

**Ответ (201):**
```json
{
  "success": true,
  "message": "Платёж создан",
  "data": {
    "order_id": "3042a36d-dced-479f-b77f-314c7c278397",
    "payment_url": "https://dev-app.rhmt.uz/invoice/d2b6cebb-e03f-4284-81b6-c7f9abc7e370",
    "amount": 100000,
    "currency": "UZS"
  }
}
```

### A3 — Премиум пакет (50 токенов / 150 000 UZS)

```bash
curl -X POST https://api.centraltrademarket.com/api/v1/payments/create \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "payment_type": "tokens",
    "pricing_key": "tokens_premium",
    "currency": "UZS",
    "return_url": "https://karvon.uz/payment/success"
  }'
```

**Ответ (201):**
```json
{
  "success": true,
  "message": "Платёж создан",
  "data": {
    "order_id": "...",
    "payment_url": "https://dev-app.rhmt.uz/invoice/...",
    "amount": 150000,
    "currency": "UZS"
  }
}
```

### A4 — Макс пакет (150 токенов / 180 000 UZS)

```bash
curl -X POST https://api.centraltrademarket.com/api/v1/payments/create \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "payment_type": "tokens",
    "pricing_key": "tokens_max",
    "currency": "UZS",
    "return_url": "https://karvon.uz/payment/success"
  }'
```

**Ответ (201):**
```json
{
  "success": true,
  "message": "Платёж создан",
  "data": {
    "order_id": "...",
    "payment_url": "https://dev-app.rhmt.uz/invoice/...",
    "amount": 180000,
    "currency": "UZS"
  }
}
```

**Что происходит после оплаты:**  
Multicard отправляет webhook `POST /payments/webhook` со статусом `"success"`.  
Сервер зачисляет токены на баланс пользователя (`token_balance += tokens_amount`).

---

## Поток B — Подписки

**Endpoint:** `POST /payments/create`  
**payment_type:** `"subscription"`

### B1 — Неделя (50 000 UZS / 7 дней)

```bash
curl -X POST https://api.centraltrademarket.com/api/v1/payments/create \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "payment_type": "subscription",
    "pricing_key": "sub_week",
    "currency": "UZS",
    "return_url": "https://karvon.uz/payment/success"
  }'
```

**Ответ (201):**
```json
{
  "success": true,
  "message": "Платёж создан",
  "data": {
    "order_id": "d2288179-fccc-4e92-959d-5765ac8d4b12",
    "payment_url": "https://dev-app.rhmt.uz/invoice/32b8a9e9-5010-4acf-a6d8-487fc2944e0a",
    "amount": 50000,
    "currency": "UZS"
  }
}
```

### B2 — Месяц (150 000 UZS / 31 день + 100 токенов)

```bash
curl -X POST https://api.centraltrademarket.com/api/v1/payments/create \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "payment_type": "subscription",
    "pricing_key": "sub_month",
    "currency": "UZS",
    "return_url": "https://karvon.uz/payment/success"
  }'
```

**Ответ (201):**
```json
{
  "success": true,
  "message": "Платёж создан",
  "data": {
    "order_id": "d09f970b-9c2f-445b-acc8-cbafe61cb3e3",
    "payment_url": "https://dev-app.rhmt.uz/invoice/cd97f367-fb88-493d-acb9-037ab3b1d777",
    "amount": 150000,
    "currency": "UZS"
  }
}
```

### B3 — Про месяц (280 000 UZS / 30 дней)

```bash
curl -X POST https://api.centraltrademarket.com/api/v1/payments/create \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "payment_type": "subscription",
    "pricing_key": "sub_pro_month",
    "currency": "UZS",
    "return_url": "https://karvon.uz/payment/success"
  }'
```

**Ответ (201):**
```json
{
  "success": true,
  "message": "Платёж создан",
  "data": {
    "order_id": "ed1da8da-d3a6-49be-8196-1b306f706a44",
    "payment_url": "https://dev-app.rhmt.uz/invoice/3db5864f-8e04-44ba-aa8f-1b3741327a82",
    "amount": 280000,
    "currency": "UZS"
  }
}
```

### B4 — Год (1 000 000 UZS / 365 дней)

```bash
curl -X POST https://api.centraltrademarket.com/api/v1/payments/create \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "payment_type": "subscription",
    "pricing_key": "sub_year",
    "currency": "UZS",
    "return_url": "https://karvon.uz/payment/success"
  }'
```

**Ответ (201):**
```json
{
  "success": true,
  "message": "Платёж создан",
  "data": {
    "order_id": "06020a6d-92f3-472c-8a41-5a51a285f292",
    "payment_url": "https://dev-app.rhmt.uz/invoice/ae2f6307-64f3-4993-ba52-329f907c92b6",
    "amount": 1000000,
    "currency": "UZS"
  }
}
```

**Что происходит после оплаты:**  
Создаётся запись `subscriptions` с полями `starts_at`, `expires_at`, `is_active=true`.

---

## Поток C — Платное объявление груза

**Endpoint:** `POST /payments/create`  
**payment_type:** `"listing"`  
**listing_type:** `"cargo"`  
**listing_id:** UUID существующего объявления груза

```bash
CARGO_ID="8ee0a8e1-d36a-4bc6-9bd6-9c7632c472bb"

curl -X POST https://api.centraltrademarket.com/api/v1/payments/create \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d "{
    \"payment_type\": \"listing\",
    \"pricing_key\": \"listing_cargo\",
    \"listing_type\": \"cargo\",
    \"listing_id\": \"$CARGO_ID\",
    \"currency\": \"UZS\",
    \"return_url\": \"https://karvon.uz/payment/success\"
  }"
```

**Ответ (201):**
```json
{
  "success": true,
  "message": "Платёж создан",
  "data": {
    "order_id": "9263d77b-ab1a-42a5-811c-a973d4d8a092",
    "payment_url": "https://dev-app.rhmt.uz/invoice/6d658487-be2d-4533-be1b-394e1e73fca2",
    "amount": 30000,
    "currency": "UZS"
  }
}
```

**`item_id` в базе:** `cargo|<listing_id>|listing_cargo`

**Что происходит после оплаты:**  
`cargo.is_paid = true`, `cargo.status = "active"`.

---

## Поток D — Платное объявление склада

**Endpoint:** `POST /payments/create`  
**payment_type:** `"listing"`  
**listing_type:** `"warehouse"`

```bash
WH_ID="05d64922-75ac-41b6-be64-fec5d5091540"

curl -X POST https://api.centraltrademarket.com/api/v1/payments/create \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d "{
    \"payment_type\": \"listing\",
    \"pricing_key\": \"listing_warehouse\",
    \"listing_type\": \"warehouse\",
    \"listing_id\": \"$WH_ID\",
    \"currency\": \"UZS\",
    \"return_url\": \"https://karvon.uz/payment/success\"
  }"
```

**Ответ (201):**
```json
{
  "success": true,
  "message": "Платёж создан",
  "data": {
    "order_id": "4c1680af-b527-4948-b986-752cd86187cb",
    "payment_url": "https://dev-app.rhmt.uz/invoice/7b21ee95-4961-45e4-a704-73ed7adfb138",
    "amount": 30000,
    "currency": "UZS"
  }
}
```

**`item_id` в базе:** `warehouse|<listing_id>|listing_warehouse`

Для холодильного склада использовать `"pricing_key": "listing_warehouse_fridge"` (50 000 UZS).

**Что происходит после оплаты:**  
`warehouse.is_paid = true` (вызывается `warehouse.MarkPaid`).

---

## Поток E — Буст объявления

**Endpoint:** `POST /listings/:type/:id/boost`  
**Auth:** Bearer токен — ТОЛЬКО владелец объявления  
**:type** = `cargo` или `warehouse`

### E1 — Буст груза на 1 день (20 000 UZS)

```bash
CARGO_ID="<ваш_cargo_id>"

curl -X POST "https://api.centraltrademarket.com/api/v1/listings/cargo/$CARGO_ID/boost" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "pricing_key": "boost_1day",
    "currency": "UZS",
    "return_url": "https://karvon.uz/payment/success"
  }'
```

**Ответ (201):**
```json
{
  "success": true,
  "message": "Платёж создан",
  "data": {
    "order_id": "...",
    "payment_url": "https://dev-app.rhmt.uz/invoice/...",
    "amount": 20000,
    "currency": "UZS"
  }
}
```

### E2 — Буст склада на 3 дня (50 000 UZS)

```bash
WH_ID="<ваш_warehouse_id>"

curl -X POST "https://api.centraltrademarket.com/api/v1/listings/warehouse/$WH_ID/boost" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "pricing_key": "boost_3day",
    "currency": "UZS",
    "return_url": "https://karvon.uz/payment/success"
  }'
```

**Ответ (201):**
```json
{
  "success": true,
  "message": "Платёж создан",
  "data": {
    "order_id": "...",
    "payment_url": "https://dev-app.rhmt.uz/invoice/...",
    "amount": 50000,
    "currency": "UZS"
  }
}
```

### E3 — Буст на 7 дней (100 000 UZS)

```bash
curl -X POST "https://api.centraltrademarket.com/api/v1/listings/cargo/$CARGO_ID/boost" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "pricing_key": "boost_7day",
    "currency": "UZS",
    "return_url": "https://karvon.uz/payment/success"
  }'
```

**`item_id` в базе:** `cargo|<listing_id>|boost_1day`

**Что происходит после оплаты:**  
`cargo.is_boosted = true`, `cargo.boost_expires_at = now + duration_days`.

---

## Поток F — Статус платежа (поллинг)

Клиент может проверять статус после редиректа с платёжной страницы.

```bash
ORDER_ID="d63eda12-b270-4ff6-aa52-1c9e939de5cf"

curl "https://api.centraltrademarket.com/api/v1/payments/$ORDER_ID" \
  -H "Authorization: Bearer $TOKEN"
```

**Ответ (200) — в ожидании:**
```json
{
  "success": true,
  "data": {
    "id": "d63eda12-b270-4ff6-aa52-1c9e939de5cf",
    "user_id": "16c9286c-f968-489a-b9b6-953673e7cafd",
    "payment_type": "tokens",
    "item_id": "tokens_mini",
    "amount": 15000,
    "currency": "UZS",
    "status": "pending",
    "rahmat_order_id": "d3511d38-10f6-4246-8285-11eba2bf94b6",
    "payment_method": null,
    "paid_at": null,
    "created_at": "2026-06-19T12:31:30.081006Z"
  }
}
```

**Ответ после успешной оплаты:**
```json
{
  "success": true,
  "data": {
    "id": "...",
    "status": "paid",
    "payment_method": "multicard",
    "paid_at": "2026-06-19T13:00:00Z"
  }
}
```

**Возможные статусы:**

| status     | Значение                                |
|------------|-----------------------------------------|
| `pending`  | Ожидает оплаты (пользователь не платил) |
| `paid`     | Успешно оплачен, эффект применён        |
| `reverted` | Платёж отменён / возврат               |

---

## Поток G — История платежей

```bash
curl "https://api.centraltrademarket.com/api/v1/payments/history?page=1&per_page=10" \
  -H "Authorization: Bearer $TOKEN"
```

**Ответ (200):**
```json
{
  "success": true,
  "data": [
    {
      "id": "06020a6d-92f3-472c-8a41-5a51a285f292",
      "payment_type": "subscription",
      "item_id": "sub_year",
      "amount": 1000000,
      "currency": "UZS",
      "status": "pending",
      "created_at": "2026-06-19T12:34:27.426207Z"
    },
    {
      "id": "4c1680af-b527-4948-b986-752cd86187cb",
      "payment_type": "listing",
      "item_id": "warehouse|05d64922-75ac-41b6-be64-fec5d5091540|listing_warehouse",
      "amount": 30000,
      "currency": "UZS",
      "status": "pending",
      "created_at": "2026-06-19T12:34:26.518007Z"
    },
    {
      "id": "9263d77b-ab1a-42a5-811c-a973d4d8a092",
      "payment_type": "listing",
      "item_id": "cargo|8ee0a8e1-d36a-4bc6-9bd6-9c7632c472bb|listing_cargo",
      "amount": 30000,
      "currency": "UZS",
      "status": "pending",
      "created_at": "2026-06-19T12:34:26.057903Z"
    }
  ],
  "meta": {
    "total": 49,
    "page": 1,
    "per_page": 10
  }
}
```

---

## Поток H — Активная подписка

```bash
curl "https://api.centraltrademarket.com/api/v1/subscriptions/active" \
  -H "Authorization: Bearer $TOKEN"
```

**Ответ (200) — нет подписки:**
```json
{
  "success": true,
  "data": null
}
```

**Ответ (200) — есть активная подписка:**
```json
{
  "success": true,
  "data": {
    "id": "...",
    "user_id": "16c9286c-f968-489a-b9b6-953673e7cafd",
    "plan": "month",
    "starts_at": "2026-06-19T13:00:00Z",
    "expires_at": "2026-07-20T13:00:00Z",
    "is_active": true,
    "payment_order_id": "d09f970b-9c2f-445b-acc8-cbafe61cb3e3"
  }
}
```

---

## Webhook от Multicard (внутренний)

Multicard автоматически вызывает этот эндпоинт после оплаты.

**Endpoint:** `POST /payments/webhook` (публичный, без авторизации)

```json
{
  "store_invoice_id": "<order_id>",
  "status": "success",
  "uuid": "multicard-uuid",
  "phone": "+998991234567",
  "payment_amount": 1500000,
  "total_amount": 1500000,
  "commission_amount": 0,
  "ps": "multicard",
  "receipt_url": "https://..."
}
```

Сервер находит заказ по `store_invoice_id` и применяет эффект (`apply(order)`).

**Обрабатываемые статусы:** `"success"` (зачислить) и `"revert"` (откатить).  
Остальные статусы (`draft`, `progress`, `billing`, `error`) — квитируются `200 {"ok":true}` без обработки.

---

## Ошибки

### 400 — TARIFF_NOT_PAYABLE (тариф бесплатный, price_uzs=0)

```bash
curl -X POST https://api.centraltrademarket.com/api/v1/payments/create \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"payment_type":"tokens","pricing_key":"tokens_registration","currency":"UZS"}'
```

```json
{
  "success": false,
  "error": {
    "code": "TARIFF_NOT_PAYABLE",
    "message": "Этот тариф нельзя оплатить (цена 0 или тариф бесплатный)"
  }
}
```

### 400 — VALIDATION_ERROR (невалидный JSON или отсутствует поле)

```json
{
  "success": false,
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Ошибка валидации данных"
  }
}
```

### 401 — UNAUTHORIZED (нет токена или истёк)

```json
{
  "success": false,
  "error": {
    "code": "UNAUTHORIZED",
    "message": "Не авторизован"
  }
}
```

### 403 — FORBIDDEN (попытка буста чужого объявления)

```bash
curl -X POST "https://api.centraltrademarket.com/api/v1/listings/cargo/<чужой_id>/boost" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"pricing_key":"boost_1day","currency":"UZS"}'
```

```json
{
  "success": false,
  "error": {
    "code": "FORBIDDEN",
    "message": "Нет прав"
  }
}
```

### 404 — NOT_FOUND (несуществующий pricing_key или listing_id)

```bash
curl -X POST https://api.centraltrademarket.com/api/v1/payments/create \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"payment_type":"tokens","pricing_key":"tokens_nonexistent","currency":"UZS"}'
```

```json
{
  "success": false,
  "error": {
    "code": "NOT_FOUND",
    "message": "Не найдено"
  }
}
```

---

## Структура тела запроса POST /payments/create

| Поле           | Тип    | Обязательно         | Описание                                       |
|----------------|--------|---------------------|------------------------------------------------|
| `payment_type` | string | Да                  | `tokens`, `subscription`, `listing`            |
| `pricing_key`  | string | Да                  | Ключ тарифа из `GET /payments/packages`        |
| `currency`     | string | Да                  | Всегда `"UZS"` (Multicard принимает только UZS)|
| `listing_type` | string | Для listing         | `"cargo"` или `"warehouse"`                    |
| `listing_id`   | string | Для listing         | UUID объявления                                |
| `return_url`   | string | Нет (рекомендуется) | Куда вернуть пользователя после оплаты        |

## Структура ответа POST /payments/create (201)

| Поле          | Тип    | Описание                                             |
|---------------|--------|------------------------------------------------------|
| `order_id`    | UUID   | ID заказа — использовать для поллинга статуса        |
| `payment_url` | string | URL страницы оплаты Multicard — открыть в браузере  |
| `amount`      | int    | Сумма в тийинах (UZS)                               |
| `currency`    | string | Всегда `"UZS"`                                      |

---

## Ключи тарифов — справочник

### Токены (`payment_type: "tokens"`)

| Ключ                  | Сумма UZS | Токены | Примечание              |
|-----------------------|-----------|--------|-------------------------|
| `tokens_mini`         | 15 000    | 10     |                         |
| `tokens_standard`     | 100 000   | 60     |                         |
| `tokens_premium`      | 150 000   | 50     |                         |
| `tokens_max`          | 180 000   | 150    |                         |
| `tokens_registration` | 0         | 5      | НЕ ПЛАТНЫЙ — только webhook |

### Подписки (`payment_type: "subscription"`)

| Ключ            | Сумма UZS | Дней | Токены в подарок |
|-----------------|-----------|------|------------------|
| `sub_week`      | 50 000    | 7    | —                |
| `sub_month`     | 150 000   | 31   | 100              |
| `sub_pro_month` | 280 000   | 30   | —                |
| `sub_year`      | 1 000 000 | 365  | —                |

### Объявления (`payment_type: "listing"`)

| Ключ                      | Сумма UZS | listing_type | Примечание              |
|---------------------------|-----------|--------------|-------------------------|
| `listing_cargo`           | 30 000    | cargo        |                         |
| `listing_paid`            | 30 000    | cargo        | Альтернативный ключ     |
| `listing_warehouse`       | 30 000    | warehouse    |                         |
| `listing_warehouse_fridge`| 50 000    | warehouse    | Холодильный склад       |
| `listing_dds`             | 30 000    | cargo        | Тест-ключ (устаревший)  |

### Буст (`POST /listings/:type/:id/boost`)

| Ключ         | Сумма UZS | Дней |
|--------------|-----------|------|
| `boost_1day` | 20 000    | 1    |
| `boost_3day` | 50 000    | 3    |
| `boost_7day` | 100 000   | 7    |

---

## Полный цикл оплаты (схема)

```
1. Клиент: POST /payments/create  →  получает payment_url + order_id
2. Клиент: открывает payment_url в браузере / WebView
3. Пользователь: вводит данные карты на странице Multicard
4. Multicard: POST /payments/webhook {"status":"success",...}
5. Сервер: применяет эффект:
   - tokens     → token_balance += tokens_amount
   - subscription → создаёт запись subscription с expires_at
   - listing    → cargo/warehouse.is_paid = true
   - boost      → cargo/warehouse.is_boosted = true + boost_expires_at
6. Клиент: GET /payments/<order_id>  →  polling до status = "paid"
7. Клиент: возвращается на return_url
```

**При отмене / возврате:**

```
Multicard: POST /payments/webhook {"status":"revert",...}
Сервер: откатывает эффект:
   - tokens     → token_balance -= tokens_amount
   - subscription → is_active = false
   - listing    → is_paid = false, status = "archived"
   - boost      → is_boosted = false, boost_expires_at = null
```
