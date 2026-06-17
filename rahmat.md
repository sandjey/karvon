---
title: Default module
language_tabs:
  - shell: Shell
  - http: HTTP
  - javascript: JavaScript
  - ruby: Ruby
  - python: Python
  - php: PHP
  - java: Java
  - go: Go
toc_footers: []
includes: []
search: true
code_clipboard: true
highlight_theme: darkula
headingLevel: 2
generator: "@tarslib/widdershins v4.0.30"

---

# Default module

Документация API для партнеров и мерчантов, подключаемых к платежному шлюзу Multicard.

**Стенды:**  
Sandbox endpoint: [https://dev-mesh.multicard.uz/](https://dev-mesh.multicard.uz/)  
Production endpoint: [https://mesh.multicard.uz/](https://mesh.multicard.uz/)

**Тестовые карты**  
8600303655375959 2603  
8600492998494476 2601

Sandbox OTP: 1122

### Формат ответов

Все запросы направляются в формате JSON. В каждом ответе присутствует поле success, которое отражает результат выполнения запроса (true / false). В случае, если результат успешен, полезные данные передаются в объекте data. В случае ошибки, возвращается объект error с полями code и details.

Пример ответа с ошибкой:

``` json
{
    "success": false,
    "error": {
        "code": "ERROR_FIELDS",
        "details": "Поле store_id является обязательным"
        }
    }
}

 ```

Base URLs:

# Authentication

- HTTP Authentication, scheme: bearer

# Авторизация

## POST Получение токена

POST /auth

> Body Parameters

```json
{
    "application_id": "rhmt_test",
    "secret": "Pw18axeBFo8V7NamKHXX"
}
```

### Params

|Name|Location|Type|Required|Description|
|---|---|---|---|---|
|body|body|object| no |none|
|» application_id|body|string| yes |Идентификатор приложения, выданный со стороны Multicard|
|» secret|body|string| yes |Секретный ключ (пароль) приложения|

> Response Examples

> 200 Response

```json
{
    "token": "{{vault:json-web-token}}",
    "role": "dev",
    "expiry": "2023-03-18 16:40:31"
}
```

> 400 Response

```json
{
    "errors": [
        {
            "message": [
                {
                    "message": "application_id required",
                    "field": "application_id"
                },
                {
                    "message": "You must supply a secret",
                    "field": "secret"
                }
            ]
        }
    ]
}
```

### Responses

|HTTP Status Code |Meaning|Description|Data schema|
|---|---|---|---|
|200|[OK](https://tools.ietf.org/html/rfc7231#section-6.3.1)|none|Inline|
|400|[Bad Request](https://tools.ietf.org/html/rfc7231#section-6.5.1)|При попытке получения токена без заполнения обязательных полей, возвращается ответ с кодом 400, и уведомляет пользователя об обязательных полях|Inline|

### Responses Data Schema

HTTP Status Code **200**

|Name|Type|Required|Restrictions|Title|description|
|---|---|---|---|---|---|
|» token|string|true|none||JWT-токен. Необходимо передавать в заголовке <Authorization: Bearer {token}>|
|» role|string|true|none||Роль пользователя или приложения, которому выдан токен|
|» expiry|string|true|none||Время истечения токена, после которого необходимо его запросить заново. Временная зона GMT+5|

HTTP Status Code **400**

|Name|Type|Required|Restrictions|Title|description|
|---|---|---|---|---|---|
|» errors|[object]|true|none||Ошибка|
|»» message|[object]|false|none||Сообщение об ошибке|
|»»» message|string|true|none||Сообщение об обязательном поле для заполнения|
|»»» field|string|true|none||Поле для обязательного заполнения|

# Оплата на платежной странице Multicard

## POST Создание инвойса

POST /payment/invoice

Метод создает инвойс и возвращает ссылку на его оплату (в поле checkout_url). Система мерчанта должна перенаправить клиента на указанный URL.

В случае, если система Партнера формирует QR-код для оплаты, то можно использовать короткую ссылку из поля short_link (возвращается только в продакшн-среде).

Система партнера должна сохранить идентификатор транзакции из поля uuid.

Если передано поле callback_url, при успешной оплате в систему Партнера будет отправлен коллбэк-запрос.

> Body Parameters

```json
{
    "store_id": 6,
    "amount": 500000,
    "invoice_id": "123",
    "lang": "ru",
    "return_url": "https://..",
    "callback_url": "https://..",
    "ofd": [
        {
            "qty": 1,
            "price": 2500000,
            "mxik": "06401004002000000",
            "total": 25000000,
            "package_code": "1506113",
            "name": "кроссовки men's low shoes"
        },
        {
            "qty": 1,
            "price": 2500000,
            "mxik": "06401002004000000",
            "total": 2500000,
            "package_code": "1519041",
            "name": "кроссовки t.ace 2332 white black"
        }
    ]
}
```

### Params

|Name|Location|Type|Required|Description|
|---|---|---|---|---|
|X-Access-Token|header|string| no |none|
|body|body|object| no |none|
|» store_id|body|string| yes |ID (int) или UUID (string) кассы, выданный со стороны Multicard|
|» amount|body|integer| yes |Сумма платежа в тийинах|
|» invoice_id|body|string| yes |Любой идентификатор платежа в системе Партнера. Будет возвращен в callback-запросе. Также по нему можно искать платежи в кабинете мерчанта|
|» lang|body|string| no |Язык страницы чекаута|
|» return_url|body|string| no |URL для перенаправления пользователя после оплаты|
|» return_error_url|body|string| no |URL для перенаправления пользователя после неуспешной оплаты. Если не передан, то используется return_url|
|» callback_url|body|string| yes |URL для отправки callback-запроса|
|» sms|body|string| no |Если передано, то по указанному номер будет отправлена ссылка на инвойс. Формат 998XXXXXXXXX|
|» ttl|body|integer¦null| no |Время жизни инвойса в секундах. После заданного времени, вслучае если оплата не была проведена, инвойс будет отменен. По умолчанию - 1 день|
|» ofd|body|[object]| yes |Данные для формирования фискального чека|
|»» qty|body|integer| yes |Количество единиц товара/услуги|
|»» vat|body|integer| no |НДС (%)|
|»» price|body|integer| yes |Стоимость единицы товара в тийинах|
|»» mxik|body|string| yes |ИКПУ из справочника tasnif.soliq.uz|
|»» total|body|integer| yes |Общая сумма товаров с учетом количества без учета скидок в тийинах|
|»» package_code|body|string| yes |Код упаковки из справочника tasnif.soliq.uz|
|»» name|body|string| yes |Наименование товара/услуги|
|»» tin|body|string| no |ИНН компании|
|»» mark|body|[string]| no |Массив с кодами маркировок каждой единицы товара. Обязателен для маркировочных товаров|

#### Enum

|Name|Value|
|---|---|
|» lang|ru|
|» lang|uz|
|» lang|en|

> Response Examples

```json
{
    "success": true,
    "data": {
        "store_id": 6,
        "amount": 100000,
        "invoice_id": "123",
        "return_url": "https://..",
        "callback_url": "https://..",
        "ofd": [
            {
                "qty": 1,
                "price": 60000000,
                "mxik": "06401004002000000",
                "total": 60000000,
                "package_code": "1506113",
                "name": "кроссовки men's low shoes"
            },
            {
                "qty": 1,
                "price": 55700000,
                "mxik": "06401002004000000",
                "total": 55700000,
                "package_code": "1519041",
                "name": "кроссовки t.ace 2332 white black"
            }
        ],
        "uuid": "f6339f31-6a09-11f0-9a1b-00505680eaf6",
        "products": null,
        "split": null,
        "return_error_url": null,
        "short_link": "https://l.multicard.uz/1m1e12",
        "sms": null,
        "added_on": "2025-07-26 15:19:06",
        "updated_on": "2025-07-26 15:19:06",
        "payment": "{PaymentModel}",
        "checkout_url": "https://app.rhmt.uz/invoice/f6339f31-6a09-11f0-9a1b-00505680eaf6",
        "deeplink": "https://multicard.app/payments/checkout/6?params=%7B%22fixed_amount%22%3A100000%2C%22details%22%3A%7B%22invoice_id%22%3A%22123%22%2C%22invoice_uuid%22%3A%22f6339f31-6a09-11f0-9a1b-00505680eaf6%22%7D%7D"
    }
}
```

```json
{
    "success": false,
    "error": {
        "code": "ERROR_FIELDS",
        "details": "Необходимо заполнить «Invoice Id»."
    }
}
```

> 400 Response

```json
{
  "success": true,
  "error": {
    "code": "string",
    "details": "string"
  }
}
```

### Responses

|HTTP Status Code |Meaning|Description|Data schema|
|---|---|---|---|
|200|[OK](https://tools.ietf.org/html/rfc7231#section-6.3.1)|none|Inline|
|400|[Bad Request](https://tools.ietf.org/html/rfc7231#section-6.5.1)|Ошибка возникает при отправке запроса с пустым полем, к примеру invoice_id|Inline|

### Responses Data Schema

HTTP Status Code **200**

|Name|Type|Required|Restrictions|Title|description|
|---|---|---|---|---|---|
|» success|boolean|true|none||Успешно|
|» data|object|true|none||Данные инвойса|
|»» uuid|string(uuid)|true|none||Уникальный идентификатор транзакции в Multicard|
|»» store_id|integer|true|none||ID кассы, выданный со стороны Multicard|
|»» amount|integer|true|none||Сумма платежа в тийинах|
|»» invoice_id|string|true|none||Идентификатор заказа/пользователя в системе Партнера|
|»» return_url|string¦null|true|none||Ссылка для возврата плательщика на страницу Партнера после неуспешной оплаты|
|»» return_error_url|string¦null|true|none||Адрес (URL), на который пользователь будет перенаправлен в случае ошибки при оплате|
|»» callback_url|string¦null|true|none||URL для отправки callback-запроса|
|»» ofd|[object]¦null|true|none||Данные для формирования фискального чека|
|»»» qty|integer|true|none||Количество единиц товара/услуги|
|»»» vat|integer|false|none||НДС (%)|
|»»» price|integer|true|none||Стоимость единицы товара в тийинах|
|»»» mxik|string|true|none||ИКПУ из справочника tasnif.soliq.uz|
|»»» total|integer|true|none||Общая сумма товаров с учетом количества без учета скидок в тийинах|
|»»» package_code|string|true|none||Код упаковки из справочника tasnif.soliq.uz|
|»»» name|string|true|none||Наименование товара/услуги|
|»»» tin|string|false|none||ИНН компании|
|»»» mark|[string]|false|none||Массив с кодами маркировок каждой единицы товара. Обязателен для маркировочных товаров|
|»» split|[object]¦null|true|none||Информация о разделении платежа между несколькими получателями (split payment). Если null, разделение не используется.|
|»»» type|string|true|none||Тип получателя|
|»»» amount|integer|true|none||Сумма платежа|
|»»» details|string|true|none||Детали платежа|
|»»» recipient|string|false|none||Получатель. Если type=account, то передается uuid банковских реквизитов. По-умолчанию подставляются банковские реквизиты мерчанта.|
|»» sms|string¦null|true|none||Номер телефона пользователя, куда отправлена ссылка на инвойс|
|»» payment|[paymentModel](#schemapaymentmodel)|true|none||Детали платежа|
|»»» id|integer|true|none||ID транзакции в Multicard|
|»»» uuid|string(uuid)|true|none||UUID транзакции в Multicard|
|»»» store_id|integer|true|none||ID кассы, выданный со стороны Multicard|
|»»» payment_amount|integer|true|none||none|
|»»» commission_type|string|true|none||Тип комиссии|
|»»» commission_amount|integer|true|none||none|
|»»» total_amount|integer|true|none||Сумма платежа в тийинах|
|»»» store_invoice_id|string|true|none||Любой идентификатор заказа в системе Партнера. Будет возвращен в callback-запросе. Также по нему можно искать платежаи в кабинете мерчанта|
|»»» status|[PaymentStatusEnum](#schemapaymentstatusenum)|true|none||Статус транзакции|
|»»» callback_url|string¦null|true|none||URL для отправки callback-запроса|
|»»» billing_id|string¦null|true|none||Уникальный ID транзакции в системе Партнера|
|»»» phone|string¦null|true|none||Телефон плательщика в формате 998XXXXXXXXX (при наличии)|
|»»» ps|[PaymentServiceEnum](#schemapaymentserviceenum)|true|none||Платежный сервис/система|
|»»» receipt_url|string¦null|true|none||Ссылка на платежный чек|
|»»» kyc_data|object¦null|true|none||none|
|»»»» last_name|string|true|none||Фамилия плательщика|
|»»»» first_name|string|true|none||Имя плательщика|
|»»»» middle_name|string|true|none||Отчество плательщика|
|»»»» passport|string|true|none||Паспорт плательщика|
|»»»» dob|string|true|none||Дата рождения плательщика (YYYY-MM-DD)|
|»»»» passport_expiry_date|string|true|none||Дата истечения паспорта плательщика|
|»»» device_details|object¦null|true|none||Объект с информацией об устройстве клиента|
|»»»» ip|string|true|none||IP-адрес устройства клиента, с которого выполняется запрос|
|»»»» user_agent|string|true|none||Cтрока User-Agent, содержащая информацию о браузере, операционной системе и типе устройства клиента|
|»»» details|object¦null|true|none||Поля, необходимые для проведения платежа в биллинге. Используется только при оплате за услуги Paynet и МУНИС (мобильная связь, гос.платежи и т.п.). Список полей и их название зависит от конкретной услуги (store_id)|
|»»» card_token|string¦null|true|none||none|
|»»» card_pan|string¦null|true|none||none|
|»»» split|[object]¦null|true|none||none|
|»»» multicard_user_id|integer¦null|true|none||ID пользователя в приложении Multicard (если оплата проведена через приложение)|
|»»» ofd|[object]¦null|true|none||none|
|»»» terminal_id|string¦null|true|none||ID терминала в платежной системе|
|»»» merchant_id|string¦null|true|none||ID мерчанта в платежной системе|
|»»» ps_uniq_id|string¦null|true|none||Reference number (RRN, RefNum) в платежной системе|
|»»» ps_response_code|string¦null|true|none||Код ответа (ошибки) от платежной системы|
|»»» ps_response_msg|string¦null|true|none||Описание ошибки от платежной системы|
|»»» callback_message|string¦null|true|none||Текст ответа от биллинга мерчанта/поставщика|
|»»» payment_time|string(date-time)¦null|true|none||Дата и время оплаты во временной зоне GMT+5 в формате YYYY-mm-dd H:s:s|
|»»» refund_time|string(date-time)|true|none||Дата и время возврата во временной зоне GMT+5 в формате YYYY-mm-dd H:s:s|
|»»» otp_hash|string¦null|true|none||Если значение не null, то требуется подтверждение транзакции через SMS-код (или 3DS)|
|»»» clearing_id|integer¦null|true|none||none|
|»»» tax_receipt_id|integer¦null|true|none||none|
|»»» push_sent_at|string(date-time)¦null|true|none||none|
|»»» store|[storeModel](#schemastoremodel)|true|none||none|
|»»»» id|integer|true|none||none|
|»»»» uuid|string(uuid)|true|none||none|
|»»»» category_id|integer¦null|true|none||none|
|»»»» note|string¦null|true|none||none|
|»»»» logo|string¦null|true|none||Логотип|
|»»»» color|string¦null|true|none||none|
|»»»» view_fields|[object]¦null|true|none||none|
|»»»»» type|string|true|none||Формат поля|
|»»»»» name|string|true|none||Описание поля|
|»»»»» value|any|true|none||Значение|

*oneOf*

|Name|Type|Required|Restrictions|Title|description|
|---|---|---|---|---|---|
|»»»»»» *anonymous*|string|false|none||none|

*xor*

|Name|Type|Required|Restrictions|Title|description|
|---|---|---|---|---|---|
|»»»»»» *anonymous*|integer|false|none||none|

*xor*

|Name|Type|Required|Restrictions|Title|description|
|---|---|---|---|---|---|
|»»»»»» *anonymous*|boolean|false|none||none|

*xor*

|Name|Type|Required|Restrictions|Title|description|
|---|---|---|---|---|---|
|»»»»»» *anonymous*|array|false|none||none|

*xor*

|Name|Type|Required|Restrictions|Title|description|
|---|---|---|---|---|---|
|»»»»»» *anonymous*|object|false|none||none|

*xor*

|Name|Type|Required|Restrictions|Title|description|
|---|---|---|---|---|---|
|»»»»»» *anonymous*|number|false|none||none|

*continued*

|Name|Type|Required|Restrictions|Title|description|
|---|---|---|---|---|---|
|»»»»» key|string|true|none||Ключ|
|»»»»» suggested|boolean¦null|true|none||Рекоммендуемая сумма оплаты|
|»»»» tax_registration|integer|true|none||Флаг фискализации|
|»»»» tax_mxik|string¦null|true|none||ИКПУ по умолчанию из каталога tasnif.soliq.uz|
|»»»» tax_package_code|string¦null|true|none||Код упаковки от ИКПУ по умолчанию из каталога tasnif.soliq.uz|
|»»»» tax_commission_recipient_tin|string¦null|true|none||ИНН комитента для фискализации. Если null, то берется ИНН мерчанта|
|»»»» tg_chat_id|string¦null|true|none||ID телеграм группы для отправки уведомлений о платежах|
|»»»» qr_url|string|true|none||Ссылка на QR-код для приема платежей по данной кассе|
|»»»» bg_img|string|true|none||Фоновая картинка для страницы чекаута|
|»»»» title|string|true|none||Наименование кассы|
|»»»» merchant|[merchantModel](#schemamerchantmodel)|true|none||Информация о клиенте (мерчанте)|
|»»»»» id|integer|true|none||ID клиента (мерчанта) в Multicard|
|»»»»» name|string|true|none||Наименование клиента (мерчанта)|
|»»»»» tin|string|true|none||ИНН мерчанта|
|»»»»» contract_id|string¦null|true|none||Данные о договоре|
|»»»»» bank_account|string¦null|true|none||Транзитный счет для расчетов|
|»»»» contract|object¦null|true|none||Информация о контракте с мерчантом|
|»»»»» id|integer|true|none||none|
|»»»»» num|string|true|none||Номер договора|
|»»»»» date|string(date)|true|none||Дата договора|
|»»»»» service|string|true|none||none|
|»»»»» fee|object|true|none||Комиссия по договору|
|»»»»»» up|string|true|none||none|
|»»»»»» down|string|true|none||none|
|»»»»» edm_document_id|string¦null|true|none||Идентификатор документа в системе электронного документооборота|
|»»»»» edm_status|string¦null|true|none||Статус подписания документа в системе электронного документооборота|
|»»»» merchant_account|[merchantAccount](#schemamerchantaccount)|true|none||none|
|»»»»» id|integer|true|none||ID банковских реквизитов|
|»»»»» uuid|string(uuid)|true|none||UUID банковских реквизитов|
|»»»»» tin|string|true|none||ИНН или ПИНФЛ|
|»»»»» official_name|string|true|none||Наименование|
|»»»»» mfo|string|true|none||МФО|
|»»»»» account_no|string|true|none||Номер счета|
|»»»»» address|string¦null|true|none||Юридический адрес|
|»»»»» director|string¦null|true|none||ФИО директора|
|»»»»» director_pinfl|string¦null|true|none||ПИНФЛ директора|
|»»»»» vat_payer|boolean¦null|true|none||Является ли плательщиков НДС|
|»»»»» is_commitent|boolean¦null|true|none||Является ли комитентом Multicard|
|»»»»» active|boolean|true|none||Флаг активности. Если отключен, значит возмещение не будет отправлено|
|»»» application|[applicationModel](#schemaapplicationmodel)|true|none||none|
|»»»» id|integer|true|none||none|
|»»»» application_id|string|true|none||none|
|»»»» wallet_sum|integer¦null|true|none||none|
|»»»» wallet_sender_account|string¦null|true|none||none|
|»»»» wallet_overdraft|integer|true|none||none|
|»»»» wallet_contract_num|string¦null|true|none||none|
|»»»» otp_required|integer|true|none||none|
|»»»» otp_gateway|string¦null|true|none||none|
|»»»» sms_nickname|string¦null|true|none||none|
|»»» tax|[taxReceiptModel](#schemataxreceiptmodel)¦null|true|none||none|
|»»»» receipt|object¦null|true|none||Объект чека, отправленный в ГНК|
|»»»» f_num|string(uuid)|true|none||Фискальный признак|
|»»»» fm_num|string|true|none||Фискальный терминал|
|»»»» qr_url|string¦null|true|none||URL на фискальный чек|
|»»»» is_refund|boolean|true|none||Является ли чеком возврата|
|»»» refund_tax|[taxReceiptModel](#schemataxreceiptmodel)¦null|true|none||none|
|»»»» receipt|object¦null|true|none||Объект чека, отправленный в ГНК|
|»»»» f_num|string(uuid)|true|none||Фискальный признак|
|»»»» fm_num|string|true|none||Фискальный терминал|
|»»»» qr_url|string¦null|true|none||URL на фискальный чек|
|»»»» is_refund|boolean|true|none||Является ли чеком возврата|
|»»» clearing|[clearingModel](#schemaclearingmodel)¦null|true|none||none|
|»»»» id|integer|true|none||none|
|»»»» merchant|[merchantModel](#schemamerchantmodel)|true|none||Информация о клиенте (мерчанте)|
|»»»»» id|integer|true|none||ID клиента (мерчанта) в Multicard|
|»»»»» name|string|true|none||Наименование клиента (мерчанта)|
|»»»»» tin|string|true|none||ИНН мерчанта|
|»»»»» contract_id|string¦null|true|none||Данные о договоре|
|»»»»» bank_account|string¦null|true|none||Транзитный счет для расчетов|
|»»»» recipient_info|[merchantAccount](#schemamerchantaccount)|true|none||none|
|»»»»» id|integer|true|none||ID банковских реквизитов|
|»»»»» uuid|string(uuid)|true|none||UUID банковских реквизитов|
|»»»»» tin|string|true|none||ИНН или ПИНФЛ|
|»»»»» official_name|string|true|none||Наименование|
|»»»»» mfo|string|true|none||МФО|
|»»»»» account_no|string|true|none||Номер счета|
|»»»»» address|string¦null|true|none||Юридический адрес|
|»»»»» director|string¦null|true|none||ФИО директора|
|»»»»» director_pinfl|string¦null|true|none||ПИНФЛ директора|
|»»»»» vat_payer|boolean¦null|true|none||Является ли плательщиков НДС|
|»»»»» is_commitent|boolean¦null|true|none||Является ли комитентом Multicard|
|»»»»» active|boolean|true|none||Флаг активности. Если отключен, значит возмещение не будет отправлено|
|»»»» sender_info|[merchantAccount](#schemamerchantaccount)|true|none||none|
|»»»»» id|integer|true|none||ID банковских реквизитов|
|»»»»» uuid|string(uuid)|true|none||UUID банковских реквизитов|
|»»»»» tin|string|true|none||ИНН или ПИНФЛ|
|»»»»» official_name|string|true|none||Наименование|
|»»»»» mfo|string|true|none||МФО|
|»»»»» account_no|string|true|none||Номер счета|
|»»»»» address|string¦null|true|none||Юридический адрес|
|»»»»» director|string¦null|true|none||ФИО директора|
|»»»»» director_pinfl|string¦null|true|none||ПИНФЛ директора|
|»»»»» vat_payer|boolean¦null|true|none||Является ли плательщиков НДС|
|»»»»» is_commitent|boolean¦null|true|none||Является ли комитентом Multicard|
|»»»»» active|boolean|true|none||Флаг активности. Если отключен, значит возмещение не будет отправлено|
|»»»» purpose_code|string|true|none||Код назначения платежа|
|»»»» amount|integer|true|none||Сумма платежа в тийинах|
|»»»» details|string|true|none||Детали платежа|
|»»»» status|string|true|none||Статус|
|»»»» payment_time|string|true|none||Время проведения платежа|
|»»»» added_on|string|true|none||Дата создания записи|
|»»»» updated_on|string|true|none||Дата изменения записи|
|»»»» receipt_url|string|true|none||URL на банковскую квитанцию|
|»»» checkout_url|string¦null|true|none||URL страницы для оплаты|
|»»» added_on|string(date-time)|true|none||none|
|»» checkout_url|string|true|none||Ссылка на страницу оплаты|
|»» short_link|string|true|none||Короткая ссылка на страницу оплаты (полезно в случае формирования QR-кода)|
|»» deeplink|string|true|none||Специальная ссылка, которая открывает мобильное приложение или веб-страницу оплаты с заранее подставленными параметрами|
|»» added_on|string|true|none||Дата и время создания инвойса|
|»» updated_on|string¦null|true|none||Дата и время последнего обновления инвойса в системе|

#### Enum

|Name|Value|
|---|---|
|type|account|
|type|wallet|
|type|card|
|commission_type|up|
|commission_type|down|
|status|draft|
|status|progress|
|status|billing|
|status|success|
|status|error|
|status|revert|
|ps|uzcard|
|ps|humo|
|ps|visa|
|ps|mastercard|
|ps|account|
|ps|payme|
|ps|click|
|ps|uzum|
|ps|anorbank|
|ps|oson|
|ps|alif|
|ps|xazna|
|ps|beepul|
|ps|trastpay|
|ps|sbp|
|type|string|
|type|int|
|type|phone|
|type|tree|
|type|hidden|
|type|complex|
|type|select|
|tax_registration|0|
|tax_registration|1|
|tax_registration|2|
|tax_registration|3|
|status|new|
|status|sent|
|status|done|
|status|repeat|
|status|postponed|
|status|blocked|
|status|revert|
|status|canceled|

HTTP Status Code **400**

|Name|Type|Required|Restrictions|Title|description|
|---|---|---|---|---|---|
|» success|boolean|true|none||Статус запроса|
|» error|object|true|none||Описание ошибки|
|»» code|string|true|none||Кодовое сообщение|
|»» details|string|true|none||Описание ошибки|

## GET Получение информации о созданном инвойсе

GET /payment/invoice/{uuid}

Метод позволяет получить детали созданного инвойса. Структура ответа аналогична ответу на запрос создания инвойса (выше)

### Params

|Name|Location|Type|Required|Description|
|---|---|---|---|---|
|uuid|path|string| yes |uuid инвойса после его создания|
|X-Access-Token|header|string| no |none|

> Response Examples

> 200 Response

```json
{
    "success": false,
    "error": {
        "code": "ERROR_NOT_FOUND",
        "details": "Объект не найден"
    }
}
```

> 404 Response

```json
{
  "success": true,
  "error": {
    "code": "string",
    "details": "string"
  }
}
```

### Responses

|HTTP Status Code |Meaning|Description|Data schema|
|---|---|---|---|
|200|[OK](https://tools.ietf.org/html/rfc7231#section-6.3.1)|none|Inline|
|400|[Bad Request](https://tools.ietf.org/html/rfc7231#section-6.5.1)|none|None|
|404|[Not Found](https://tools.ietf.org/html/rfc7231#section-6.5.4)|none|Inline|

### Responses Data Schema

HTTP Status Code **200**

|Name|Type|Required|Restrictions|Title|description|
|---|---|---|---|---|---|
|» success|boolean|true|none||Успешно|
|» data|object|true|none||Данные инвойса|
|»» uuid|string(uuid)|true|none||Уникальный идентификатор транзакции в Multicard|
|»» store_id|integer|true|none||ID кассы, выданный со стороны Multicard|
|»» amount|integer|true|none||Сумма платежа в тийинах|
|»» invoice_id|string|true|none||Идентификатор заказа/пользователя в системе Партнера|
|»» return_url|string¦null|true|none||Ссылка для возврата плательщика на страницу Партнера после неуспешной оплаты|
|»» return_error_url|string¦null|true|none||Адрес (URL), на который пользователь будет перенаправлен в случае ошибки при оплате|
|»» callback_url|string¦null|true|none||URL для отправки callback-запроса|
|»» ofd|[object]¦null|true|none||Данные для формирования фискального чека|
|»»» qty|integer|true|none||Количество единиц товара/услуги|
|»»» vat|integer|false|none||НДС (%)|
|»»» price|integer|true|none||Стоимость единицы товара в тийинах|
|»»» mxik|string|true|none||ИКПУ из справочника tasnif.soliq.uz|
|»»» total|integer|true|none||Общая сумма товаров с учетом количества без учета скидок в тийинах|
|»»» package_code|string|true|none||Код упаковки из справочника tasnif.soliq.uz|
|»»» name|string|true|none||Наименование товара/услуги|
|»»» tin|string|false|none||ИНН компании|
|»»» mark|[string]|false|none||Массив с кодами маркировок каждой единицы товара. Обязателен для маркировочных товаров|
|»» split|[object]¦null|true|none||Информация о разделении платежа между несколькими получателями (split payment). Если null, разделение не используется.|
|»»» type|string|true|none||Тип получателя|
|»»» amount|integer|true|none||Сумма платежа|
|»»» details|string|true|none||Детали платежа|
|»»» recipient|string|false|none||Получатель. Если type=account, то передается uuid банковских реквизитов. По-умолчанию подставляются банковские реквизиты мерчанта.|
|»» sms|string¦null|true|none||Номер телефона пользователя, куда отправлена ссылка на инвойс|
|»» payment|[paymentModel](#schemapaymentmodel)|true|none||Детали платежа|
|»»» id|integer|true|none||ID транзакции в Multicard|
|»»» uuid|string(uuid)|true|none||UUID транзакции в Multicard|
|»»» store_id|integer|true|none||ID кассы, выданный со стороны Multicard|
|»»» payment_amount|integer|true|none||none|
|»»» commission_type|string|true|none||Тип комиссии|
|»»» commission_amount|integer|true|none||none|
|»»» total_amount|integer|true|none||Сумма платежа в тийинах|
|»»» store_invoice_id|string|true|none||Любой идентификатор заказа в системе Партнера. Будет возвращен в callback-запросе. Также по нему можно искать платежаи в кабинете мерчанта|
|»»» status|[PaymentStatusEnum](#schemapaymentstatusenum)|true|none||Статус транзакции|
|»»» callback_url|string¦null|true|none||URL для отправки callback-запроса|
|»»» billing_id|string¦null|true|none||Уникальный ID транзакции в системе Партнера|
|»»» phone|string¦null|true|none||Телефон плательщика в формате 998XXXXXXXXX (при наличии)|
|»»» ps|[PaymentServiceEnum](#schemapaymentserviceenum)|true|none||Платежный сервис/система|
|»»» receipt_url|string¦null|true|none||Ссылка на платежный чек|
|»»» kyc_data|object¦null|true|none||none|
|»»»» last_name|string|true|none||Фамилия плательщика|
|»»»» first_name|string|true|none||Имя плательщика|
|»»»» middle_name|string|true|none||Отчество плательщика|
|»»»» passport|string|true|none||Паспорт плательщика|
|»»»» dob|string|true|none||Дата рождения плательщика (YYYY-MM-DD)|
|»»»» passport_expiry_date|string|true|none||Дата истечения паспорта плательщика|
|»»» device_details|object¦null|true|none||Объект с информацией об устройстве клиента|
|»»»» ip|string|true|none||IP-адрес устройства клиента, с которого выполняется запрос|
|»»»» user_agent|string|true|none||Cтрока User-Agent, содержащая информацию о браузере, операционной системе и типе устройства клиента|
|»»» details|object¦null|true|none||Поля, необходимые для проведения платежа в биллинге. Используется только при оплате за услуги Paynet и МУНИС (мобильная связь, гос.платежи и т.п.). Список полей и их название зависит от конкретной услуги (store_id)|
|»»» card_token|string¦null|true|none||none|
|»»» card_pan|string¦null|true|none||none|
|»»» split|[object]¦null|true|none||none|
|»»» multicard_user_id|integer¦null|true|none||ID пользователя в приложении Multicard (если оплата проведена через приложение)|
|»»» ofd|[object]¦null|true|none||none|
|»»» terminal_id|string¦null|true|none||ID терминала в платежной системе|
|»»» merchant_id|string¦null|true|none||ID мерчанта в платежной системе|
|»»» ps_uniq_id|string¦null|true|none||Reference number (RRN, RefNum) в платежной системе|
|»»» ps_response_code|string¦null|true|none||Код ответа (ошибки) от платежной системы|
|»»» ps_response_msg|string¦null|true|none||Описание ошибки от платежной системы|
|»»» callback_message|string¦null|true|none||Текст ответа от биллинга мерчанта/поставщика|
|»»» payment_time|string(date-time)¦null|true|none||Дата и время оплаты во временной зоне GMT+5 в формате YYYY-mm-dd H:s:s|
|»»» refund_time|string(date-time)|true|none||Дата и время возврата во временной зоне GMT+5 в формате YYYY-mm-dd H:s:s|
|»»» otp_hash|string¦null|true|none||Если значение не null, то требуется подтверждение транзакции через SMS-код (или 3DS)|
|»»» clearing_id|integer¦null|true|none||none|
|»»» tax_receipt_id|integer¦null|true|none||none|
|»»» push_sent_at|string(date-time)¦null|true|none||none|
|»»» store|[storeModel](#schemastoremodel)|true|none||none|
|»»»» id|integer|true|none||none|
|»»»» uuid|string(uuid)|true|none||none|
|»»»» category_id|integer¦null|true|none||none|
|»»»» note|string¦null|true|none||none|
|»»»» logo|string¦null|true|none||Логотип|
|»»»» color|string¦null|true|none||none|
|»»»» view_fields|[object]¦null|true|none||none|
|»»»»» type|string|true|none||Формат поля|
|»»»»» name|string|true|none||Описание поля|
|»»»»» value|any|true|none||Значение|

*oneOf*

|Name|Type|Required|Restrictions|Title|description|
|---|---|---|---|---|---|
|»»»»»» *anonymous*|string|false|none||none|

*xor*

|Name|Type|Required|Restrictions|Title|description|
|---|---|---|---|---|---|
|»»»»»» *anonymous*|integer|false|none||none|

*xor*

|Name|Type|Required|Restrictions|Title|description|
|---|---|---|---|---|---|
|»»»»»» *anonymous*|boolean|false|none||none|

*xor*

|Name|Type|Required|Restrictions|Title|description|
|---|---|---|---|---|---|
|»»»»»» *anonymous*|array|false|none||none|

*xor*

|Name|Type|Required|Restrictions|Title|description|
|---|---|---|---|---|---|
|»»»»»» *anonymous*|object|false|none||none|

*xor*

|Name|Type|Required|Restrictions|Title|description|
|---|---|---|---|---|---|
|»»»»»» *anonymous*|number|false|none||none|

*continued*

|Name|Type|Required|Restrictions|Title|description|
|---|---|---|---|---|---|
|»»»»» key|string|true|none||Ключ|
|»»»»» suggested|boolean¦null|true|none||Рекоммендуемая сумма оплаты|
|»»»» tax_registration|integer|true|none||Флаг фискализации|
|»»»» tax_mxik|string¦null|true|none||ИКПУ по умолчанию из каталога tasnif.soliq.uz|
|»»»» tax_package_code|string¦null|true|none||Код упаковки от ИКПУ по умолчанию из каталога tasnif.soliq.uz|
|»»»» tax_commission_recipient_tin|string¦null|true|none||ИНН комитента для фискализации. Если null, то берется ИНН мерчанта|
|»»»» tg_chat_id|string¦null|true|none||ID телеграм группы для отправки уведомлений о платежах|
|»»»» qr_url|string|true|none||Ссылка на QR-код для приема платежей по данной кассе|
|»»»» bg_img|string|true|none||Фоновая картинка для страницы чекаута|
|»»»» title|string|true|none||Наименование кассы|
|»»»» merchant|[merchantModel](#schemamerchantmodel)|true|none||Информация о клиенте (мерчанте)|
|»»»»» id|integer|true|none||ID клиента (мерчанта) в Multicard|
|»»»»» name|string|true|none||Наименование клиента (мерчанта)|
|»»»»» tin|string|true|none||ИНН мерчанта|
|»»»»» contract_id|string¦null|true|none||Данные о договоре|
|»»»»» bank_account|string¦null|true|none||Транзитный счет для расчетов|
|»»»» contract|object¦null|true|none||Информация о контракте с мерчантом|
|»»»»» id|integer|true|none||none|
|»»»»» num|string|true|none||Номер договора|
|»»»»» date|string(date)|true|none||Дата договора|
|»»»»» service|string|true|none||none|
|»»»»» fee|object|true|none||Комиссия по договору|
|»»»»»» up|string|true|none||none|
|»»»»»» down|string|true|none||none|
|»»»»» edm_document_id|string¦null|true|none||Идентификатор документа в системе электронного документооборота|
|»»»»» edm_status|string¦null|true|none||Статус подписания документа в системе электронного документооборота|
|»»»» merchant_account|[merchantAccount](#schemamerchantaccount)|true|none||none|
|»»»»» id|integer|true|none||ID банковских реквизитов|
|»»»»» uuid|string(uuid)|true|none||UUID банковских реквизитов|
|»»»»» tin|string|true|none||ИНН или ПИНФЛ|
|»»»»» official_name|string|true|none||Наименование|
|»»»»» mfo|string|true|none||МФО|
|»»»»» account_no|string|true|none||Номер счета|
|»»»»» address|string¦null|true|none||Юридический адрес|
|»»»»» director|string¦null|true|none||ФИО директора|
|»»»»» director_pinfl|string¦null|true|none||ПИНФЛ директора|
|»»»»» vat_payer|boolean¦null|true|none||Является ли плательщиков НДС|
|»»»»» is_commitent|boolean¦null|true|none||Является ли комитентом Multicard|
|»»»»» active|boolean|true|none||Флаг активности. Если отключен, значит возмещение не будет отправлено|
|»»» application|[applicationModel](#schemaapplicationmodel)|true|none||none|
|»»»» id|integer|true|none||none|
|»»»» application_id|string|true|none||none|
|»»»» wallet_sum|integer¦null|true|none||none|
|»»»» wallet_sender_account|string¦null|true|none||none|
|»»»» wallet_overdraft|integer|true|none||none|
|»»»» wallet_contract_num|string¦null|true|none||none|
|»»»» otp_required|integer|true|none||none|
|»»»» otp_gateway|string¦null|true|none||none|
|»»»» sms_nickname|string¦null|true|none||none|
|»»» tax|[taxReceiptModel](#schemataxreceiptmodel)¦null|true|none||none|
|»»»» receipt|object¦null|true|none||Объект чека, отправленный в ГНК|
|»»»» f_num|string(uuid)|true|none||Фискальный признак|
|»»»» fm_num|string|true|none||Фискальный терминал|
|»»»» qr_url|string¦null|true|none||URL на фискальный чек|
|»»»» is_refund|boolean|true|none||Является ли чеком возврата|
|»»» refund_tax|[taxReceiptModel](#schemataxreceiptmodel)¦null|true|none||none|
|»»»» receipt|object¦null|true|none||Объект чека, отправленный в ГНК|
|»»»» f_num|string(uuid)|true|none||Фискальный признак|
|»»»» fm_num|string|true|none||Фискальный терминал|
|»»»» qr_url|string¦null|true|none||URL на фискальный чек|
|»»»» is_refund|boolean|true|none||Является ли чеком возврата|
|»»» clearing|[clearingModel](#schemaclearingmodel)¦null|true|none||none|
|»»»» id|integer|true|none||none|
|»»»» merchant|[merchantModel](#schemamerchantmodel)|true|none||Информация о клиенте (мерчанте)|
|»»»»» id|integer|true|none||ID клиента (мерчанта) в Multicard|
|»»»»» name|string|true|none||Наименование клиента (мерчанта)|
|»»»»» tin|string|true|none||ИНН мерчанта|
|»»»»» contract_id|string¦null|true|none||Данные о договоре|
|»»»»» bank_account|string¦null|true|none||Транзитный счет для расчетов|
|»»»» recipient_info|[merchantAccount](#schemamerchantaccount)|true|none||none|
|»»»»» id|integer|true|none||ID банковских реквизитов|
|»»»»» uuid|string(uuid)|true|none||UUID банковских реквизитов|
|»»»»» tin|string|true|none||ИНН или ПИНФЛ|
|»»»»» official_name|string|true|none||Наименование|
|»»»»» mfo|string|true|none||МФО|
|»»»»» account_no|string|true|none||Номер счета|
|»»»»» address|string¦null|true|none||Юридический адрес|
|»»»»» director|string¦null|true|none||ФИО директора|
|»»»»» director_pinfl|string¦null|true|none||ПИНФЛ директора|
|»»»»» vat_payer|boolean¦null|true|none||Является ли плательщиков НДС|
|»»»»» is_commitent|boolean¦null|true|none||Является ли комитентом Multicard|
|»»»»» active|boolean|true|none||Флаг активности. Если отключен, значит возмещение не будет отправлено|
|»»»» sender_info|[merchantAccount](#schemamerchantaccount)|true|none||none|
|»»»»» id|integer|true|none||ID банковских реквизитов|
|»»»»» uuid|string(uuid)|true|none||UUID банковских реквизитов|
|»»»»» tin|string|true|none||ИНН или ПИНФЛ|
|»»»»» official_name|string|true|none||Наименование|
|»»»»» mfo|string|true|none||МФО|
|»»»»» account_no|string|true|none||Номер счета|
|»»»»» address|string¦null|true|none||Юридический адрес|
|»»»»» director|string¦null|true|none||ФИО директора|
|»»»»» director_pinfl|string¦null|true|none||ПИНФЛ директора|
|»»»»» vat_payer|boolean¦null|true|none||Является ли плательщиков НДС|
|»»»»» is_commitent|boolean¦null|true|none||Является ли комитентом Multicard|
|»»»»» active|boolean|true|none||Флаг активности. Если отключен, значит возмещение не будет отправлено|
|»»»» purpose_code|string|true|none||Код назначения платежа|
|»»»» amount|integer|true|none||Сумма платежа в тийинах|
|»»»» details|string|true|none||Детали платежа|
|»»»» status|string|true|none||Статус|
|»»»» payment_time|string|true|none||Время проведения платежа|
|»»»» added_on|string|true|none||Дата создания записи|
|»»»» updated_on|string|true|none||Дата изменения записи|
|»»»» receipt_url|string|true|none||URL на банковскую квитанцию|
|»»» checkout_url|string¦null|true|none||URL страницы для оплаты|
|»»» added_on|string(date-time)|true|none||none|
|»» checkout_url|string|true|none||Ссылка на страницу оплаты|
|»» short_link|string|true|none||Короткая ссылка на страницу оплаты (полезно в случае формирования QR-кода)|
|»» deeplink|string|true|none||Специальная ссылка, которая открывает мобильное приложение или веб-страницу оплаты с заранее подставленными параметрами|
|»» added_on|string|true|none||Дата и время создания инвойса|
|»» updated_on|string¦null|true|none||Дата и время последнего обновления инвойса в системе|

#### Enum

|Name|Value|
|---|---|
|type|account|
|type|wallet|
|type|card|
|commission_type|up|
|commission_type|down|
|status|draft|
|status|progress|
|status|billing|
|status|success|
|status|error|
|status|revert|
|ps|uzcard|
|ps|humo|
|ps|visa|
|ps|mastercard|
|ps|account|
|ps|payme|
|ps|click|
|ps|uzum|
|ps|anorbank|
|ps|oson|
|ps|alif|
|ps|xazna|
|ps|beepul|
|ps|trastpay|
|ps|sbp|
|type|string|
|type|int|
|type|phone|
|type|tree|
|type|hidden|
|type|complex|
|type|select|
|tax_registration|0|
|tax_registration|1|
|tax_registration|2|
|tax_registration|3|
|status|new|
|status|sent|
|status|done|
|status|repeat|
|status|postponed|
|status|blocked|
|status|revert|
|status|canceled|

HTTP Status Code **404**

|Name|Type|Required|Restrictions|Title|description|
|---|---|---|---|---|---|
|» success|boolean|true|none||Статус запроса|
|» error|object|true|none||Ошибка|
|»» code|string|true|none||Кодовое сообщение|
|»» details|string|true|none||Описание ошибки|

## DELETE Удаление (аннулирование) инвойса

DELETE /payment/invoice/{uuid}

Метод аннулирует созданный инвойс, если он еще не оплачен.

### Params

|Name|Location|Type|Required|Description|
|---|---|---|---|---|
|uuid|path|string| yes |uuid инвойса для удаления |
|X-Access-Token|header|string| no |none|

> Response Examples

> 200 Response

```json
{
    "success": true,
    "data": []
}
```

> 400 Response

```json
{
    "success": false,
    "error": {
        "code": "ERROR_FIELDS",
        "details": "Транзакция по инвойсу завершена. Запросите детали платежа"
    }
}
```

### Responses

|HTTP Status Code |Meaning|Description|Data schema|
|---|---|---|---|
|200|[OK](https://tools.ietf.org/html/rfc7231#section-6.3.1)|none|Inline|
|400|[Bad Request](https://tools.ietf.org/html/rfc7231#section-6.5.1)|При попытке аннулировании уже завершенного платежа, возвращается ошибка с кодом 400|Inline|

### Responses Data Schema

HTTP Status Code **200**

|Name|Type|Required|Restrictions|Title|description|
|---|---|---|---|---|---|
|» success|boolean|true|none||none|
|» data|null|true|none||none|

HTTP Status Code **400**

|Name|Type|Required|Restrictions|Title|description|
|---|---|---|---|---|---|
|» success|boolean|true|none||Статус запроса|
|» error|object|true|none||Ошибка|
|»» code|string|true|none||Кодовое сообщение|
|»» details|string|true|none||Описание ошибки|

## PUT Быстрая оплата (PaymeGo, ClickPass и др.)

PUT /payment/{uuid}/scanpay

Списания посредством QR-кода из приложения плательщика. Поддерживаются следующие приложения: Payme, Click, Uzum, Anorbank и Xazna.

В случае успешного выполнения возвращается {PaymentModel}

При получении таймаута, ошибки с кодом ERROR_DEBIT_UNKNOWN или http-статус 500 необходимо проверить статус платежа.

> Body Parameters

```json
{
    "code": "50512"
}
```

### Params

|Name|Location|Type|Required|Description|
|---|---|---|---|---|
|uuid|path|string| yes |uuid инвойса для оплаты через PaymeGo, ClickPass и др.|
|X-Access-Token|header|string| no |none|
|body|body|object| no |none|
|» code|body|string| yes |Считанный код из приложения|

> Response Examples

> 400 Response

```json
{
    "success": false,
    "error": {
        "code": "ERROR_FIELDS",
        "details": "Неверный код, либо приложение не поддерживается"
    }
}
```

> 200 Response

```json
{
  "success": true,
  "data": {
    "id": 0,
    "uuid": "095be615-a8ad-4c33-8e9c-c7612fbf6c9f",
    "store_id": 0,
    "payment_amount": 0,
    "commission_type": "up",
    "commission_amount": 0,
    "total_amount": 0,
    "store_invoice_id": "string",
    "status": "draft",
    "callback_url": "string",
    "billing_id": "string",
    "phone": "string",
    "ps": "uzcard",
    "receipt_url": "string",
    "kyc_data": {
      "last_name": "string",
      "first_name": "string",
      "middle_name": "string",
      "passport": "string",
      "dob": "string",
      "passport_expiry_date": "string"
    },
    "device_details": {
      "ip": "string",
      "user_agent": "string"
    },
    "details": {},
    "card_token": "string",
    "card_pan": "string",
    "split": [
      {
        "type": "account",
        "amount": 0,
        "details": "string",
        "recipient": "string"
      }
    ],
    "multicard_user_id": 0,
    "ofd": [
      {
        "qty": 0,
        "vat": 0,
        "price": 0,
        "mxik": "string",
        "total": 0,
        "package_code": "string",
        "name": "string",
        "tin": "string",
        "mark": [
          "string"
        ]
      }
    ],
    "terminal_id": "string",
    "merchant_id": "string",
    "ps_uniq_id": "string",
    "ps_response_code": "string",
    "ps_response_msg": "string",
    "callback_message": "string",
    "payment_time": "2019-08-24T14:15:22Z",
    "refund_time": "2019-08-24T14:15:22Z",
    "otp_hash": "string",
    "clearing_id": 0,
    "tax_receipt_id": 0,
    "push_sent_at": "2019-08-24T14:15:22Z",
    "store": {
      "id": 0,
      "uuid": "095be615-a8ad-4c33-8e9c-c7612fbf6c9f",
      "category_id": 0,
      "note": "string",
      "logo": "string",
      "color": "string",
      "view_fields": [
        {
          "type": "[",
          "name": "string",
          "value": {},
          "key": "string",
          "suggested": true
        }
      ],
      "tax_registration": 0,
      "tax_mxik": "string",
      "tax_package_code": "string",
      "tax_commission_recipient_tin": "string",
      "tg_chat_id": "string",
      "qr_url": "string",
      "bg_img": "string",
      "title": "string",
      "merchant": {
        "id": 0,
        "name": "string",
        "tin": "string",
        "contract_id": "string",
        "bank_account": "string"
      },
      "contract": {
        "id": 0,
        "num": "string",
        "date": "2019-08-24",
        "service": "string",
        "fee": {
          "up": null,
          "down": null
        },
        "edm_document_id": "string",
        "edm_status": "string"
      },
      "merchant_account": {
        "id": 0,
        "uuid": "095be615-a8ad-4c33-8e9c-c7612fbf6c9f",
        "tin": "string",
        "official_name": "string",
        "mfo": "string",
        "account_no": "string",
        "address": "string",
        "director": "string",
        "director_pinfl": "string",
        "vat_payer": true,
        "is_commitent": true,
        "active": true
      }
    },
    "application": {
      "id": 0,
      "application_id": "string",
      "wallet_sum": 0,
      "wallet_sender_account": "string",
      "wallet_overdraft": 0,
      "wallet_contract_num": "string",
      "otp_required": 0,
      "otp_gateway": "string",
      "sms_nickname": "string"
    },
    "tax": {
      "receipt": {},
      "f_num": "1d2db1d7-fdce-4ae6-9c43-6a437dcdbc89",
      "fm_num": "string",
      "qr_url": "string",
      "is_refund": true
    },
    "refund_tax": {
      "receipt": {},
      "f_num": "1d2db1d7-fdce-4ae6-9c43-6a437dcdbc89",
      "fm_num": "string",
      "qr_url": "string",
      "is_refund": true
    },
    "clearing": {
      "id": 0,
      "merchant": {
        "id": 0,
        "name": "string",
        "tin": "string",
        "contract_id": "string",
        "bank_account": "string"
      },
      "recipient_info": {
        "id": 0,
        "uuid": "095be615-a8ad-4c33-8e9c-c7612fbf6c9f",
        "tin": "string",
        "official_name": "string",
        "mfo": "string",
        "account_no": "string",
        "address": "string",
        "director": "string",
        "director_pinfl": "string",
        "vat_payer": true,
        "is_commitent": true,
        "active": true
      },
      "sender_info": {
        "id": 0,
        "uuid": "095be615-a8ad-4c33-8e9c-c7612fbf6c9f",
        "tin": "string",
        "official_name": "string",
        "mfo": "string",
        "account_no": "string",
        "address": "string",
        "director": "string",
        "director_pinfl": "string",
        "vat_payer": true,
        "is_commitent": true,
        "active": true
      },
      "purpose_code": "string",
      "amount": 0,
      "details": "string",
      "status": "new",
      "payment_time": "string",
      "added_on": "string",
      "updated_on": "string",
      "receipt_url": "string"
    },
    "checkout_url": "string",
    "added_on": "2019-08-24T14:15:22Z"
  }
}
```

### Responses

|HTTP Status Code |Meaning|Description|Data schema|
|---|---|---|---|
|200|[OK](https://tools.ietf.org/html/rfc7231#section-6.3.1)|none|Inline|
|400|[Bad Request](https://tools.ietf.org/html/rfc7231#section-6.5.1)|В случае показа неправильного QR для быстрого платежа, пользователь получает ошибку с кодом 400, и сообщение о неправильном QR|Inline|

### Responses Data Schema

HTTP Status Code **200**

|Name|Type|Required|Restrictions|Title|description|
|---|---|---|---|---|---|
|» success|boolean|true|none||none|
|» data|[paymentModel](#schemapaymentmodel)|true|none||none|
|»» id|integer|true|none||ID транзакции в Multicard|
|»» uuid|string(uuid)|true|none||UUID транзакции в Multicard|
|»» store_id|integer|true|none||ID кассы, выданный со стороны Multicard|
|»» payment_amount|integer|true|none||none|
|»» commission_type|string|true|none||Тип комиссии|
|»» commission_amount|integer|true|none||none|
|»» total_amount|integer|true|none||Сумма платежа в тийинах|
|»» store_invoice_id|string|true|none||Любой идентификатор заказа в системе Партнера. Будет возвращен в callback-запросе. Также по нему можно искать платежаи в кабинете мерчанта|
|»» status|[PaymentStatusEnum](#schemapaymentstatusenum)|true|none||Статус транзакции|
|»» callback_url|string¦null|true|none||URL для отправки callback-запроса|
|»» billing_id|string¦null|true|none||Уникальный ID транзакции в системе Партнера|
|»» phone|string¦null|true|none||Телефон плательщика в формате 998XXXXXXXXX (при наличии)|
|»» ps|[PaymentServiceEnum](#schemapaymentserviceenum)|true|none||Платежный сервис/система|
|»» receipt_url|string¦null|true|none||Ссылка на платежный чек|
|»» kyc_data|object¦null|true|none||none|
|»»» last_name|string|true|none||Фамилия плательщика|
|»»» first_name|string|true|none||Имя плательщика|
|»»» middle_name|string|true|none||Отчество плательщика|
|»»» passport|string|true|none||Паспорт плательщика|
|»»» dob|string|true|none||Дата рождения плательщика (YYYY-MM-DD)|
|»»» passport_expiry_date|string|true|none||Дата истечения паспорта плательщика|
|»» device_details|object¦null|true|none||Объект с информацией об устройстве клиента|
|»»» ip|string|true|none||IP-адрес устройства клиента, с которого выполняется запрос|
|»»» user_agent|string|true|none||Cтрока User-Agent, содержащая информацию о браузере, операционной системе и типе устройства клиента|
|»» details|object¦null|true|none||Поля, необходимые для проведения платежа в биллинге. Используется только при оплате за услуги Paynet и МУНИС (мобильная связь, гос.платежи и т.п.). Список полей и их название зависит от конкретной услуги (store_id)|
|»» card_token|string¦null|true|none||none|
|»» card_pan|string¦null|true|none||none|
|»» split|[object]¦null|true|none||none|
|»»» type|string|true|none||Тип получателя|
|»»» amount|integer|true|none||Сумма платежа|
|»»» details|string|true|none||Детали платежа|
|»»» recipient|string|false|none||Получатель. Если type=account, то передается uuid банковских реквизитов. По-умолчанию подставляются банковские реквизиты мерчанта.|
|»» multicard_user_id|integer¦null|true|none||ID пользователя в приложении Multicard (если оплата проведена через приложение)|
|»» ofd|[object]¦null|true|none||none|
|»»» qty|integer|true|none||Количество единиц товара/услуги|
|»»» vat|integer|false|none||НДС (%)|
|»»» price|integer|true|none||Стоимость единицы товара в тийинах|
|»»» mxik|string|true|none||ИКПУ из справочника tasnif.soliq.uz|
|»»» total|integer|true|none||Общая сумма товаров с учетом количества без учета скидок в тийинах|
|»»» package_code|string|true|none||Код упаковки из справочника tasnif.soliq.uz|
|»»» name|string|true|none||Наименование товара/услуги|
|»»» tin|string|false|none||ИНН компании|
|»»» mark|[string]|false|none||Массив с кодами маркировок каждой единицы товара. Обязателен для маркировочных товаров|
|»» terminal_id|string¦null|true|none||ID терминала в платежной системе|
|»» merchant_id|string¦null|true|none||ID мерчанта в платежной системе|
|»» ps_uniq_id|string¦null|true|none||Reference number (RRN, RefNum) в платежной системе|
|»» ps_response_code|string¦null|true|none||Код ответа (ошибки) от платежной системы|
|»» ps_response_msg|string¦null|true|none||Описание ошибки от платежной системы|
|»» callback_message|string¦null|true|none||Текст ответа от биллинга мерчанта/поставщика|
|»» payment_time|string(date-time)¦null|true|none||Дата и время оплаты во временной зоне GMT+5 в формате YYYY-mm-dd H:s:s|
|»» refund_time|string(date-time)|true|none||Дата и время возврата во временной зоне GMT+5 в формате YYYY-mm-dd H:s:s|
|»» otp_hash|string¦null|true|none||Если значение не null, то требуется подтверждение транзакции через SMS-код (или 3DS)|
|»» clearing_id|integer¦null|true|none||none|
|»» tax_receipt_id|integer¦null|true|none||none|
|»» push_sent_at|string(date-time)¦null|true|none||none|
|»» store|[storeModel](#schemastoremodel)|true|none||none|
|»»» id|integer|true|none||none|
|»»» uuid|string(uuid)|true|none||none|
|»»» category_id|integer¦null|true|none||none|
|»»» note|string¦null|true|none||none|
|»»» logo|string¦null|true|none||Логотип|
|»»» color|string¦null|true|none||none|
|»»» view_fields|[object]¦null|true|none||none|
|»»»» type|string|true|none||Формат поля|
|»»»» name|string|true|none||Описание поля|
|»»»» value|any|true|none||Значение|

*oneOf*

|Name|Type|Required|Restrictions|Title|description|
|---|---|---|---|---|---|
|»»»»» *anonymous*|string|false|none||none|

*xor*

|Name|Type|Required|Restrictions|Title|description|
|---|---|---|---|---|---|
|»»»»» *anonymous*|integer|false|none||none|

*xor*

|Name|Type|Required|Restrictions|Title|description|
|---|---|---|---|---|---|
|»»»»» *anonymous*|boolean|false|none||none|

*xor*

|Name|Type|Required|Restrictions|Title|description|
|---|---|---|---|---|---|
|»»»»» *anonymous*|array|false|none||none|

*xor*

|Name|Type|Required|Restrictions|Title|description|
|---|---|---|---|---|---|
|»»»»» *anonymous*|object|false|none||none|

*xor*

|Name|Type|Required|Restrictions|Title|description|
|---|---|---|---|---|---|
|»»»»» *anonymous*|number|false|none||none|

*continued*

|Name|Type|Required|Restrictions|Title|description|
|---|---|---|---|---|---|
|»»»» key|string|true|none||Ключ|
|»»»» suggested|boolean¦null|true|none||Рекоммендуемая сумма оплаты|
|»»» tax_registration|integer|true|none||Флаг фискализации|
|»»» tax_mxik|string¦null|true|none||ИКПУ по умолчанию из каталога tasnif.soliq.uz|
|»»» tax_package_code|string¦null|true|none||Код упаковки от ИКПУ по умолчанию из каталога tasnif.soliq.uz|
|»»» tax_commission_recipient_tin|string¦null|true|none||ИНН комитента для фискализации. Если null, то берется ИНН мерчанта|
|»»» tg_chat_id|string¦null|true|none||ID телеграм группы для отправки уведомлений о платежах|
|»»» qr_url|string|true|none||Ссылка на QR-код для приема платежей по данной кассе|
|»»» bg_img|string|true|none||Фоновая картинка для страницы чекаута|
|»»» title|string|true|none||Наименование кассы|
|»»» merchant|[merchantModel](#schemamerchantmodel)|true|none||Информация о клиенте (мерчанте)|
|»»»» id|integer|true|none||ID клиента (мерчанта) в Multicard|
|»»»» name|string|true|none||Наименование клиента (мерчанта)|
|»»»» tin|string|true|none||ИНН мерчанта|
|»»»» contract_id|string¦null|true|none||Данные о договоре|
|»»»» bank_account|string¦null|true|none||Транзитный счет для расчетов|
|»»» contract|object¦null|true|none||Информация о контракте с мерчантом|
|»»»» id|integer|true|none||none|
|»»»» num|string|true|none||Номер договора|
|»»»» date|string(date)|true|none||Дата договора|
|»»»» service|string|true|none||none|
|»»»» fee|object|true|none||Комиссия по договору|
|»»»»» up|string|true|none||none|
|»»»»» down|string|true|none||none|
|»»»» edm_document_id|string¦null|true|none||Идентификатор документа в системе электронного документооборота|
|»»»» edm_status|string¦null|true|none||Статус подписания документа в системе электронного документооборота|
|»»» merchant_account|[merchantAccount](#schemamerchantaccount)|true|none||none|
|»»»» id|integer|true|none||ID банковских реквизитов|
|»»»» uuid|string(uuid)|true|none||UUID банковских реквизитов|
|»»»» tin|string|true|none||ИНН или ПИНФЛ|
|»»»» official_name|string|true|none||Наименование|
|»»»» mfo|string|true|none||МФО|
|»»»» account_no|string|true|none||Номер счета|
|»»»» address|string¦null|true|none||Юридический адрес|
|»»»» director|string¦null|true|none||ФИО директора|
|»»»» director_pinfl|string¦null|true|none||ПИНФЛ директора|
|»»»» vat_payer|boolean¦null|true|none||Является ли плательщиков НДС|
|»»»» is_commitent|boolean¦null|true|none||Является ли комитентом Multicard|
|»»»» active|boolean|true|none||Флаг активности. Если отключен, значит возмещение не будет отправлено|
|»» application|[applicationModel](#schemaapplicationmodel)|true|none||none|
|»»» id|integer|true|none||none|
|»»» application_id|string|true|none||none|
|»»» wallet_sum|integer¦null|true|none||none|
|»»» wallet_sender_account|string¦null|true|none||none|
|»»» wallet_overdraft|integer|true|none||none|
|»»» wallet_contract_num|string¦null|true|none||none|
|»»» otp_required|integer|true|none||none|
|»»» otp_gateway|string¦null|true|none||none|
|»»» sms_nickname|string¦null|true|none||none|
|»» tax|[taxReceiptModel](#schemataxreceiptmodel)¦null|true|none||none|
|»»» receipt|object¦null|true|none||Объект чека, отправленный в ГНК|
|»»» f_num|string(uuid)|true|none||Фискальный признак|
|»»» fm_num|string|true|none||Фискальный терминал|
|»»» qr_url|string¦null|true|none||URL на фискальный чек|
|»»» is_refund|boolean|true|none||Является ли чеком возврата|
|»» refund_tax|[taxReceiptModel](#schemataxreceiptmodel)¦null|true|none||none|
|»»» receipt|object¦null|true|none||Объект чека, отправленный в ГНК|
|»»» f_num|string(uuid)|true|none||Фискальный признак|
|»»» fm_num|string|true|none||Фискальный терминал|
|»»» qr_url|string¦null|true|none||URL на фискальный чек|
|»»» is_refund|boolean|true|none||Является ли чеком возврата|
|»» clearing|[clearingModel](#schemaclearingmodel)¦null|true|none||none|
|»»» id|integer|true|none||none|
|»»» merchant|[merchantModel](#schemamerchantmodel)|true|none||Информация о клиенте (мерчанте)|
|»»»» id|integer|true|none||ID клиента (мерчанта) в Multicard|
|»»»» name|string|true|none||Наименование клиента (мерчанта)|
|»»»» tin|string|true|none||ИНН мерчанта|
|»»»» contract_id|string¦null|true|none||Данные о договоре|
|»»»» bank_account|string¦null|true|none||Транзитный счет для расчетов|
|»»» recipient_info|[merchantAccount](#schemamerchantaccount)|true|none||none|
|»»»» id|integer|true|none||ID банковских реквизитов|
|»»»» uuid|string(uuid)|true|none||UUID банковских реквизитов|
|»»»» tin|string|true|none||ИНН или ПИНФЛ|
|»»»» official_name|string|true|none||Наименование|
|»»»» mfo|string|true|none||МФО|
|»»»» account_no|string|true|none||Номер счета|
|»»»» address|string¦null|true|none||Юридический адрес|
|»»»» director|string¦null|true|none||ФИО директора|
|»»»» director_pinfl|string¦null|true|none||ПИНФЛ директора|
|»»»» vat_payer|boolean¦null|true|none||Является ли плательщиков НДС|
|»»»» is_commitent|boolean¦null|true|none||Является ли комитентом Multicard|
|»»»» active|boolean|true|none||Флаг активности. Если отключен, значит возмещение не будет отправлено|
|»»» sender_info|[merchantAccount](#schemamerchantaccount)|true|none||none|
|»»»» id|integer|true|none||ID банковских реквизитов|
|»»»» uuid|string(uuid)|true|none||UUID банковских реквизитов|
|»»»» tin|string|true|none||ИНН или ПИНФЛ|
|»»»» official_name|string|true|none||Наименование|
|»»»» mfo|string|true|none||МФО|
|»»»» account_no|string|true|none||Номер счета|
|»»»» address|string¦null|true|none||Юридический адрес|
|»»»» director|string¦null|true|none||ФИО директора|
|»»»» director_pinfl|string¦null|true|none||ПИНФЛ директора|
|»»»» vat_payer|boolean¦null|true|none||Является ли плательщиков НДС|
|»»»» is_commitent|boolean¦null|true|none||Является ли комитентом Multicard|
|»»»» active|boolean|true|none||Флаг активности. Если отключен, значит возмещение не будет отправлено|
|»»» purpose_code|string|true|none||Код назначения платежа|
|»»» amount|integer|true|none||Сумма платежа в тийинах|
|»»» details|string|true|none||Детали платежа|
|»»» status|string|true|none||Статус|
|»»» payment_time|string|true|none||Время проведения платежа|
|»»» added_on|string|true|none||Дата создания записи|
|»»» updated_on|string|true|none||Дата изменения записи|
|»»» receipt_url|string|true|none||URL на банковскую квитанцию|
|»» checkout_url|string¦null|true|none||URL страницы для оплаты|
|»» added_on|string(date-time)|true|none||none|

#### Enum

|Name|Value|
|---|---|
|commission_type|up|
|commission_type|down|
|status|draft|
|status|progress|
|status|billing|
|status|success|
|status|error|
|status|revert|
|ps|uzcard|
|ps|humo|
|ps|visa|
|ps|mastercard|
|ps|account|
|ps|payme|
|ps|click|
|ps|uzum|
|ps|anorbank|
|ps|oson|
|ps|alif|
|ps|xazna|
|ps|beepul|
|ps|trastpay|
|ps|sbp|
|type|account|
|type|wallet|
|type|card|
|type|string|
|type|int|
|type|phone|
|type|tree|
|type|hidden|
|type|complex|
|type|select|
|tax_registration|0|
|tax_registration|1|
|tax_registration|2|
|tax_registration|3|
|status|new|
|status|sent|
|status|done|
|status|repeat|
|status|postponed|
|status|blocked|
|status|revert|
|status|canceled|

HTTP Status Code **400**

|Name|Type|Required|Restrictions|Title|description|
|---|---|---|---|---|---|
|» success|boolean|true|none||Статус запроса|
|» error|object|true|none||Ошибка|
|»» code|string|true|none||Кодовое сообщение|
|»» details|string|true|none||Сообщение об ошибке|

# Привязка карт (форма)

## POST Callback-запрос

POST /

Данный запрос отправляется со стороны системы Multicard в систему Партнера.

При успешном добавлении карты пользователем, на переданный в поле callback_url адрес, приходит следующий запрос.

В поле card_token указан токен карты для проведения платежных транзакций.

В случае, если на callback-запрос система Партнера ответила http-статусом, отличным от 200, то запрос отправляется через 60 секунд (всего 5 попыток).

> Body Parameters

```json
{
    "id": 3554547,
    "application_id": 9,
    "payer_id": "688470c8060af",
    "card_pan": "986027******1519",
    "card_token": "6225f3c93f7a880142782fa4",
    "phone": "99891234567",
    "holder_name": "AAAAA BBBBB",
    "pinfl": null,
    "ps": "humo",
    "status": "active",
    "added_on": "2025-03-17 03:19:12",
    "card_status": null,
    "sms_inform": null,
    "is_multicard": false
}
```

### Params

|Name|Location|Type|Required|Description|
|---|---|---|---|---|
|body|body|object| no |none|
|» id|body|integer| yes |none|
|» application_id|body|integer| yes |Идентификатор приложения, выданный со стороны Multicard|
|» payer_id|body|string| yes |Значение поля session_id из ответа на запрос привязки карты|
|» card_pan|body|string| yes |Max length:16|
|» card_token|body|string| yes |Max length:255|
|» phone|body|string| yes |Max length:12|
|» holder_name|body|string| yes |Max length:100|
|» pinfl|body|null| yes |Default:|
|» ps|body|string| yes |Allowed values (5)|
|» status|body|string| yes |Allowed values (3)|
|» added_on|body|string| yes |Время добавления карты в формате YYYY-mm-dd H:i:s|
|» card_status|body|null| yes |none|
|» sms_inform|body|null| yes |none|
|» is_multicard|body|boolean| yes |none|

#### Description

**» card_pan**: Max length:16
Маскированный номер карты

**» card_token**: Max length:255
Токен карты для проведения платежных операций с карто

**» phone**: Max length:12
Телефон держателя карты в формате 998XXXXXXXXX (поддерживается только для карт Uzcard и Humo)

**» holder_name**: Max length:100
Имя держателя карт

**» pinfl**: Default:
ПИНФЛ держателя карты (возвращается в случае передачи в запросе на создание привязки для карт Uzcard и Humo)
Max length:14

**» ps**: Allowed values (5)
uzcard,humo,visa,mastercard,unionpay hide all
Платежная система

**» status**: Allowed values (3)
active,
draft,
deleted
 
hide all
Статус привязки. active - карта успешно привязана, можно проводить платежные операции по токену карты

> Response Examples

> 200 Response

```json
{}
```

### Responses

|HTTP Status Code |Meaning|Description|Data schema|
|---|---|---|---|
|200|[OK](https://tools.ietf.org/html/rfc7231#section-6.3.1)|none|Inline|

### Responses Data Schema

## POST Получение ссылки на страницу привязки карты

POST /payment/card/bind

Метод для создания сессии по привязке карты. Срок активности сессии – 15 минут.

Необходимо перенаправить пользователя по URL (открыть WebView), указанному в поле data.form_url.

Для получения результата можно передать callback_url, тогда будет отправлен callback-запрос, либо проверять состояние с помощью метода проверки привязанной карты.

> Body Parameters

```json
{
    "redirect_url": "https://site.uz/success.html",
    "redirect_decline_url": "https://site.uz/decline.html",
    "store_id": 6,
    "callback_url": "https://site.uz/card-callback"
}
```

### Params

|Name|Location|Type|Required|Description|
|---|---|---|---|---|
|body|body|object| no |none|
|» redirect_url|body|string| yes |URL для перенаправления после успешного добавления карты пользователем|
|» redirect_decline_url|body|string| yes |URL для перенаправления при неуспешном добавлении карты или отмены пользователем|
|» store_id|body|integer| yes |ID кассы, выданный со стороны Multicard|
|» callback_url|body|string| yes |Если передан, то по заданному URL будет отправлен callback-запрос (webhook) при успешном добавлении карты. Формат запроса приведен ниже|
|» phone|body|string| yes |Номер телефона пользователя в формате 998XXXXXXXXX. Обязателен для карт Humo. Если передан, то будет проверено соответствие с номер смс-информаирования на карте (для Uzcard/Humo)|
|» pinfl|body|string| no |ПИНФЛ клиента. Если передан, то будет осуществлена проверка на принадлежность карты к данному лицу (только для Uzcard/Humo)|

> Response Examples

> 200 Response

```json
{
    "success": true,
    "data": {
        "session_id": "67f8dd24e9800",
        "form_url": "https://dev-checkout.multicard.uz/card/67f8dd24e9800"
    }
}
```

> 400 Response

```json
{
    "success": false,
    "error": {
        "code": "ERROR_FIELDS",
        "details": "Не переданы обязательные поля: store_id"
    }
}
```

### Responses

|HTTP Status Code |Meaning|Description|Data schema|
|---|---|---|---|
|200|[OK](https://tools.ietf.org/html/rfc7231#section-6.3.1)|none|Inline|
|400|[Bad Request](https://tools.ietf.org/html/rfc7231#section-6.5.1)|При отправке запроса с допущением ошибки валидации, возвращается ошибка с кодом 400|Inline|

### Responses Data Schema

HTTP Status Code **200**

|Name|Type|Required|Restrictions|Title|description|
|---|---|---|---|---|---|
|» success|boolean|true|none||none|
|» data|object|true|none||none|
|»» session_id|string|true|none||Необходимо сохранить данное значение, оно возвращается в callback-запросе в поле payer_id. Также можно проверить состояние через метод проверки привязанной карты|
|»» form_url|string|true|none||URL формы добавления карты (необходимо перенаправить пользователя, либо открыть webView)|

HTTP Status Code **400**

|Name|Type|Required|Restrictions|Title|description|
|---|---|---|---|---|---|
|» success|boolean|true|none||Статус запроса|
|» error|object|true|none||Ошибка|
|»» code|string|true|none||Кодовое сообщение|
|»» details|string|true|none||Описание ошибки|

## GET Проверка состояния привязки карты

GET /payment/card/bind/{session_id}

С помощью данного метода можно проверить состояние привязки карты (вместо ожидания callback-запроса)

### Params

|Name|Location|Type|Required|Description|
|---|---|---|---|---|
|session_id|path|string| yes |Идентификатор сессии окна привязки карты, время действия 15 минут|

> Response Examples

> 200 Response

```json
{
    "success": true,
    "data": {
        "id": 55,
        "application_id": 1,
        "payer_id": "1233321",
        "card_pan": "409784******3066",
        "card_token": "6225f3c93f7a110142782fa4",
        "phone": "998901234567",
        "holder_name": "AAAA BBBBB",
        "pinfl": null,
        "ps": "uzcard",
        "status": "active",
        "added_on": "2022-04-21 14:36:29",
        "card_status": {
            "code": 0,
            "description": "Active"
        },
        "sms_inform": true
    }
}
```

### Responses

|HTTP Status Code |Meaning|Description|Data schema|
|---|---|---|---|
|200|[OK](https://tools.ietf.org/html/rfc7231#section-6.3.1)|none|Inline|

### Responses Data Schema

HTTP Status Code **200**

|Name|Type|Required|Restrictions|Title|description|
|---|---|---|---|---|---|
|» success|boolean|true|none||none|
|» data|object|true|none||none|
|»» id|integer|true|none||none|
|»» application_id|integer|true|none||Идентификатор приложения, выданный со стороны Multicard|
|»» payer_id|string|true|none||Значение поля session_id из ответа на запрос привязки карты|
|»» card_pan|string|true|none||Max length:16<br />Маскированный номер карты|
|»» card_token|string|true|none||Max length:255<br />Токен карты для проведения платежных операций с картой|
|»» phone|string|true|none||Max length:12<br />Телефон держателя карты в формате 998XXXXXXXXX (поддерживается только для карт Uzcard и Humo)|
|»» holder_name|string|true|none||Max length:100<br />Имя держателя карты|
|»» pinfl|null|true|none||Default:<br />ПИНФЛ держателя карты (возвращается в случае передачи в запросе на создание привязки для карт Uzcard и Humo)<br />Max length:14|
|»» ps|string|true|none||Allowed values (5)<br />uzcard,<br />humo,<br />visa,<br />mastercard,<br />unionpay<br />hide all<br />Платежная система|
|»» status|string|true|none||Allowed values (3)<br />active,<br />draft,<br />deleted<br />hide all<br />Статус привязки. active - карта успешно привязана, можно проводить платежные операции по токену карты|
|»» added_on|string|true|none||Время добавления карты в формате YYYY-mm-dd H:i:s|
|»» card_status|object|true|none||Текущий статус карты (например, активна или заблокирована). Содержит код и текстовое описание.|
|»»» code|integer|true|none||Состояние карты. 0 - карта активна|
|»»» description|string|true|none||Описание блокировки карты|
|»» sms_inform|boolean|true|none||Признак подключения SMS-уведомлений (true = включены, false = отключены).|

## GET Получение информации о карте по токену

GET /payment/card/{card_token}

> Body Parameters

```json
{}
```

### Params

|Name|Location|Type|Required|Description|
|---|---|---|---|---|
|card_token|path|string| yes |Идентификатор добавленной карты в систему|
|body|body|object| no |none|

> Response Examples

> 200 Response

```json
{
    "success": true,
    "data": {
        "id": 55,
        "application_id": 1,
        "payer_id": "2313122",
        "card_pan": "409784******3066",
        "card_token": "6225f3c93f7a110142782fa4",
        "phone": "998901234567",
        "holder_name": "AAAA BBBBB",
        "pinfl": null,
        "ps": "uzcard",
        "status": "active",
        "added_on": "2022-04-21 14:36:29",
        "card_status": {
            "code": 0,
            "description": "Active"
        },
        "sms_inform": true
    }
}
```

> 400 Response

```json
{
    "success": false,
    "error": {
        "code": "ERROR_CARD_NOT_FOUND",
        "details": "Неверный номер карты или срок действия"
    }
}
```

### Responses

|HTTP Status Code |Meaning|Description|Data schema|
|---|---|---|---|
|200|[OK](https://tools.ietf.org/html/rfc7231#section-6.3.1)|none|Inline|
|400|[Bad Request](https://tools.ietf.org/html/rfc7231#section-6.5.1)|При поиске несуществуюещей карты, возвращается ошибка с кодом 400, также с текстом|Inline|

### Responses Data Schema

HTTP Status Code **200**

|Name|Type|Required|Restrictions|Title|description|
|---|---|---|---|---|---|
|» success|boolean|true|none||none|
|» data|object|true|none||none|
|»» id|integer|true|none||none|
|»» application_id|integer|true|none||Идентификатор приложения, выданный со стороны Multicard|
|»» payer_id|string|true|none||Значение поля session_id из ответа на запрос привязки карты|
|»» card_pan|string|true|none||Max length:16<br />Маскированный номер карты|
|»» card_token|string|true|none||Max length:255<br />Токен карты для проведения платежных операций с картой|
|»» phone|string|true|none||Max length:12<br />Телефон держателя карты в формате 998XXXXXXXXX (поддерживается только для карт Uzcard и Humo)|
|»» holder_name|string|true|none||Max length:100<br />Имя держателя карты|
|»» pinfl|null|true|none||Default:ПИНФЛ держателя карты (возвращается в случае передачи в запросе на создание привязки для карт Uzcard и Humo)<br />Max length:14|
|»» ps|string|true|none||Allowed values (5)<br />uzcard,<br />humo,<br />visa,<br />mastercard,<br />unionpay<br />hide all<br />Платежная система|
|»» status|string|true|none||Allowed values (3)<br />active,<br />draft,<br />deleted<br />hide all<br />Статус привязки. active - карта успешно привязана, можно проводить платежные операции по токену карты|
|»» added_on|string|true|none||Время добавления карты в формате YYYY-mm-dd H:i:s|
|»» card_status|object|true|none||Текущий статус карты (например, активна или заблокирована). Содержит код и текстовое описание.|
|»»» code|integer|true|none||Состояние карты. 0 - карта активна|
|»»» description|string|true|none||Описание блокировки карты|
|»» sms_inform|boolean|true|none||Признак подключения SMS-уведомлений (true = включены, false = отключены).|

HTTP Status Code **400**

|Name|Type|Required|Restrictions|Title|description|
|---|---|---|---|---|---|
|» success|boolean|true|none||Статус запроса|
|» error|object|true|none||Ошибка|
|»» code|string|true|none||Кодовое сообщение|
|»» details|string|true|none||Описание ошибки|

## DELETE Аннулирование токена карты

DELETE /payment/card/{card_token}

Запрос на удаление карты (аннулирование токена карты)

### Params

|Name|Location|Type|Required|Description|
|---|---|---|---|---|
|card_token|path|string| yes |none|

> Response Examples

> 200 Response

```json
{
    "success": true,
    "data": []
}
```

> 400 Response

```json
{
    "success": false,
    "error": {
        "code": "ERROR_CARD_NOT_FOUND",
        "details": "Неверный номер карты или срок действия"
    }
}
```

### Responses

|HTTP Status Code |Meaning|Description|Data schema|
|---|---|---|---|
|200|[OK](https://tools.ietf.org/html/rfc7231#section-6.3.1)|none|Inline|
|400|[Bad Request](https://tools.ietf.org/html/rfc7231#section-6.5.1)|При попытке аннулировании с недействительным номером карты, возвращается ошибка с кодом 400|Inline|

### Responses Data Schema

HTTP Status Code **200**

|Name|Type|Required|Restrictions|Title|description|
|---|---|---|---|---|---|
|» success|boolean|true|none||none|
|» data|[string]|true|none||none|

HTTP Status Code **400**

|Name|Type|Required|Restrictions|Title|description|
|---|---|---|---|---|---|
|» success|boolean|true|none||Статус запроса|
|» error|object|true|none||Ошибка|
|»» code|string|true|none||Кодовое сообщение|
|»» details|string|true|none||Описание ошибки|

## POST Проверка принадлежности карты к ПИНФЛ

POST /payment/card/check-pinfl

Метод осуществляет проверку соответствия владельца карты. Поддерживается только для карт Uzcard и Humo.

При успешном выполнении в ответе в поле data возвращается true (ПИНФЛ соответствует) или false (не соответствует); в случае, если результат неизвестен - null

> Body Parameters

```json
{
    "pan": "8600303655375959",
    "pinfl": "12345678901234"
}
```

### Params

|Name|Location|Type|Required|Description|
|---|---|---|---|---|
|body|body|object| no |none|
|» pan|body|string| yes |Номер карты (Uzcard или Humo)|
|» pinfl|body|string| yes |Min length:14|

#### Description

**» pinfl**: Min length:14
Max length:14
ПИНФЛ держателя карты для проверки

> Response Examples

> 200 Response

```json
{
    "success": true,
    "data": false
}
```

> 400 Response

```json
{
    "success": false,
    "error": {
        "code": "ERROR_FIELDS",
        "details": "Необходимо заполнить «Pinfl»."
    }
}
```

### Responses

|HTTP Status Code |Meaning|Description|Data schema|
|---|---|---|---|
|200|[OK](https://tools.ietf.org/html/rfc7231#section-6.3.1)|none|Inline|
|400|[Bad Request](https://tools.ietf.org/html/rfc7231#section-6.5.1)|При отправке запроса с допущением ошибки валидации, возвращается ошибка с кодом 400|Inline|

### Responses Data Schema

HTTP Status Code **200**

|Name|Type|Required|Restrictions|Title|description|
|---|---|---|---|---|---|
|» success|boolean|true|none||none|
|» data|boolean|true|none||true - соответствует; false - не соответствует; null - результат неизвестен|

HTTP Status Code **400**

|Name|Type|Required|Restrictions|Title|description|
|---|---|---|---|---|---|
|» success|boolean|true|none||Статус запроса|
|» error|object|true|none||Ошибка|
|»» code|string|true|none||Кодовое сообщение|
|»» details|string|true|none||Описание ошибки|

# Привязка карт (API)

## PUT Подтверждение привязки

PUT /payment/card/{card_token}

> Body Parameters

```json
{
    "otp": "112233"
}
```

### Params

|Name|Location|Type|Required|Description|
|---|---|---|---|---|
|card_token|path|string| yes |Идентификатор добавляемой карты в систему|
|body|body|object| no |none|
|» otp|body|string| yes |Одноразовый пароль для подтверждения операции|
|» pinfl|body|string¦null| no |ПИНФЛ клиента. Если передан, то будет осуществлена проверка на принадлежность карты к данному лицу (только для Uzcard/Humo)|

> Response Examples

```json
{
    "success": true,
    "data": {
        "id": 55,
        "application_id": 1,
        "payer_id": "2313122",
        "card_pan": "409784******3066",
        "card_token": "6225f3c93f7a110142782fa4",
        "phone": "998901234567",
        "holder_name": "AAAA BBBBB",
        "pinfl": null,
        "ps": "uzcard",
        "status": "active",
        "added_on": "2022-04-21 14:36:29",
        "card_status": {
            "code": 0,
            "description": "Active"
        },
        "sms_inform": true
    }
}
```

```json
{
    "success": false,
    "error": {
        "code": "ERROR_WRONG_OTP",
        "details": "Код подтверждения истек, запросите новый"
    }
}
```

```json
{
    "success": false,
    "error": {
        "code": "ERROR_WRONG_OTP",
        "details": "Неверный код подтверждения"
    }
}
```

```json
{
    "success": false,
    "error": {
        "code": "ERROR_FIELDS",
        "details": "Не передан SMS-код"
    }
}
```

> 400 Response

```json
{
  "success": true,
  "error": {
    "code": "string",
    "details": "string"
  }
}
```

### Responses

|HTTP Status Code |Meaning|Description|Data schema|
|---|---|---|---|
|200|[OK](https://tools.ietf.org/html/rfc7231#section-6.3.1)|none|Inline|
|400|[Bad Request](https://tools.ietf.org/html/rfc7231#section-6.5.1)|При вводе некорректного код пароля, система возвращает ошибку с кодом 400, и сообщает клиенту об ошибке|Inline|
|423|[Locked](https://tools.ietf.org/html/rfc2518#section-10.4)|При вводе истёкшого код пароля, система возвращает ошибку с кодом 400, и сообщает клиенту об ошибке|Inline|

### Responses Data Schema

HTTP Status Code **200**

|Name|Type|Required|Restrictions|Title|description|
|---|---|---|---|---|---|
|» success|boolean|true|none||none|
|» data|object|true|none||none|
|»» id|integer|true|none||none|
|»» application_id|integer|true|none||Идентификатор приложения, выданный со стороны Multicard|
|»» payer_id|string|true|none||Значение поля session_id из ответа на запрос привязки карты|
|»» card_pan|string|true|none||Max length:16<br />Маскированный номер карты|
|»» card_token|string|true|none||Max length:255<br />Токен карты для проведения платежных операций с картой|
|»» phone|string|true|none||Телефон плательщика в формате 998XXXXXXXXX (при наличии)|
|»» holder_name|string|true|none||Имя и фамилия владельца карты|
|»» pinfl|null|true|none||ПИНФЛ держателя карты (возвращается в случае передачи в запросе на создание привязки для карт Uzcard и Humo)<br />Max length:14|
|»» ps|string|true|none||Allowed values (5)<br />uzcard,<br />humo,<br />visa,<br />mastercard,<br />unionpay<br />hide all<br />Платежная система|
|»» status|[PaymentStatusEnum](#schemapaymentstatusenum)|true|none||none|
|»» added_on|string|true|none||Время добавления карты в формате YYYY-mm-dd H:i:s|
|»» card_status|object|true|none||none|
|»»» code|integer|true|none||none|
|»»» description|string|true|none||none|
|»» sms_inform|boolean|true|none||none|

#### Enum

|Name|Value|
|---|---|
|status|draft|
|status|progress|
|status|billing|
|status|success|
|status|error|
|status|revert|

HTTP Status Code **400**

|Name|Type|Required|Restrictions|Title|description|
|---|---|---|---|---|---|
|» success|boolean|true|none||Статус запроса|
|» error|object|true|none||Ошибка|
|»» code|string|true|none||Кодовое сообщение|
|»» details|string|true|none||Описание ошибки|

HTTP Status Code **423**

|Name|Type|Required|Restrictions|Title|description|
|---|---|---|---|---|---|
|» success|boolean|true|none||Статус запроса|
|» error|object|true|none||Ошибка|
|»» code|string|true|none||Кодовое сообщение|
|»» details|string|true|none||Описание ошибки|

## POST Добавление карты с помощью карточных данных

POST /payment/card

Метод отправляет SMS-код на номер телефон держателя карты. Отправленный код необходимо передать в методе подтверждения привязки.
Метод доступен только для партнеров, которых имеется сертификат PCI DSS.

> Body Parameters

```json
{
    "pan": "8600303655375959",
    "expiry": "2603"
}
```

### Params

|Name|Location|Type|Required|Description|
|---|---|---|---|---|
|body|body|object| no |none|
|» pan|body|string| yes |Номер карты (Uzcard или Humo)|
|» expiry|body|string| yes |Срок действия карты (yymm)|
|» user_phone|body|string| yes |Номер телефона пользователя в формате 998XXXXXXXXX. Обязателен для карт Humo. Если передан, то будет проверено соответствие с номер смс-информаирования на карте (для Uzcard/Humo)|
|» cvc|body|string| no |none|
|» holder_name|body|string| no |none|
|» session_id|body|string| no |none|

> Response Examples

> 200 Response

```json
{
    "success": true,
    "data": {
        "id": 55,
        "application_id": 1,
        "payer_id": null,
        "card_pan": "409784******3066",
        "card_token": "6225f3c93f7a110142782fa4",
        "phone": null,
        "holder_name": null,
        "pinfl": null,
        "ps": "uzcard",
        "status": "draft",
        "added_on": "2022-04-21 14:36:29",
        "card_status": null,
        "sms_inform": null
    }
}
```

> 400 Response

```json
{
    "success": false,
    "error": {
        "code": "ERROR_CARD_SMS",
        "details": "На карте 986027******1519 не подключено SMS-информирование"
    }
}
```

> 429 Response

```json
{
    "success": false,
    "error": {
        "code": "ERROR_SMS_ALREADY_SENT",
        "details": "SMS с кодом уже отправлено: дождитесь его получения, либо попробуйте позже через 2 минуты"
    }
}
```

### Responses

|HTTP Status Code |Meaning|Description|Data schema|
|---|---|---|---|
|200|[OK](https://tools.ietf.org/html/rfc7231#section-6.3.1)|none|Inline|
|400|[Bad Request](https://tools.ietf.org/html/rfc7231#section-6.5.1)|При попытке добавления карты с не подключенным SMS информированием, система возвращает ошибку с кодом 400, и сообщает клиенту об ошибке|Inline|
|429|[Too Many Requests](https://tools.ietf.org/html/rfc6585#section-4)|При повторной отправки запроса на SMS пароль, система возвращает ошибку с кодом 400, и клиент получает сообщение об ошибке|Inline|

### Responses Data Schema

HTTP Status Code **200**

|Name|Type|Required|Restrictions|Title|description|
|---|---|---|---|---|---|
|» success|boolean|true|none||none|
|» data|object|true|none||none|
|»» id|integer|true|none||none|
|»» application_id|integer|true|none||Идентификатор приложения, выданный со стороны Multicard|
|»» payer_id|null|true|none||Значение поля session_id из ответа на запрос привязки карты|
|»» card_pan|string|true|none||Max length:16<br />Маскированный номер карты|
|»» card_token|string|true|none||Max length:255<br />Токен карты для проведения платежных операций с картой|
|»» phone|null|true|none||Max length:12<br />Телефон держателя карты в формате 998XXXXXXXXX (поддерживается только для карт Uzcard и Humo)|
|»» holder_name|null|true|none||Имя и фамилия владельца карты|
|»» pinfl|null|true|none||ПИНФЛ держателя карты (возвращается в случае передачи в запросе на создание привязки для карт Uzcard и Humo)<br />Max length:14|
|»» ps|string|true|none||Allowed values (5)<br />uzcard,<br />humo,<br />visa,<br />mastercard,<br />unionpay<br />hide all<br />Платежная система|
|»» status|[PaymentStatusEnum](#schemapaymentstatusenum)|true|none||none|
|»» added_on|string|true|none||Время добавления карты в формате YYYY-mm-dd H:i:s|
|»» card_status|null|true|none||none|
|»» sms_inform|null|true|none||none|

#### Enum

|Name|Value|
|---|---|
|status|draft|
|status|progress|
|status|billing|
|status|success|
|status|error|
|status|revert|

HTTP Status Code **400**

|Name|Type|Required|Restrictions|Title|description|
|---|---|---|---|---|---|
|» success|boolean|true|none||Статус запроса|
|» error|object|true|none||Ошибка|
|»» code|string|true|none||Кодовое сообщение|
|»» details|string|true|none||Описание ошибки|

HTTP Status Code **429**

|Name|Type|Required|Restrictions|Title|description|
|---|---|---|---|---|---|
|» success|boolean|true|none||Статус запроса|
|» error|object|true|none||Ошибка|
|»» code|string|true|none||Кодовое сообщение|
|»» details|string|true|none||Описание ошибки|

## GET Проверка карты по ее номеру

GET /payment/card/check/{pan}

### Params

|Name|Location|Type|Required|Description|
|---|---|---|---|---|
|pan|path|string| yes |none|

> Response Examples

> 200 Response

```json
{
    "success": true,
    "data": {
        "ps": "humo",
        "holder_name": "ILXOM IDIYEV",
        "masked_pan": "986027******1519",
        "bank": {
            "bin": "986027",
            "name": "Orient Finance Bank",
            "color": "D1AA3B",
            "logo": "https://cdn.multicard.uz/banks/ofb.png",
            "ps_bank_id": "27",
            "country": "860"
        }
    }
}
```

> 400 Response

```json
{
    "success": false,
    "error": {
        "code": "ERROR_CARD_NOT_FOUND",
        "details": "Неверный номер карты или срок действия"
    }
}
```

### Responses

|HTTP Status Code |Meaning|Description|Data schema|
|---|---|---|---|
|200|[OK](https://tools.ietf.org/html/rfc7231#section-6.3.1)|none|Inline|
|400|[Bad Request](https://tools.ietf.org/html/rfc7231#section-6.5.1)|При поиске карты с недействительным номером, возвращается ошибка с кодом 400|Inline|

### Responses Data Schema

HTTP Status Code **200**

|Name|Type|Required|Restrictions|Title|description|
|---|---|---|---|---|---|
|» success|boolean|true|none||none|
|» data|object|true|none||none|
|»» ps|string|true|none||Платежная система|
|»» holder_name|string¦null|true|none||Имя держателя|
|»» masked_pan|string|true|none||Маскированный PAN|
|»» bank|object|true|none||none|
|»»» bin|string|true|none||BIN|
|»»» name|string|true|none||Наименование банка|
|»»» color|string¦null|true|none||HEX-код цвета фона для логотипа|
|»»» logo|string¦null|true|none||URL на логотип|
|»»» ps_bank_id|string¦null|true|none||Код банка|
|»»» country|string¦null|true|none||Страна|

#### Enum

|Name|Value|
|---|---|
|ps|uzcard|
|ps|humo|
|ps|visa|
|ps|mastercard|

HTTP Status Code **400**

|Name|Type|Required|Restrictions|Title|description|
|---|---|---|---|---|---|
|» success|boolean|true|none||none|
|» error|object|true|none||none|
|»» code|string|true|none||none|
|»» details|string|true|none||none|

# Оплата на странице Партнера

## POST Создание платежа payme/click/uzum и прочие

POST /payment

Пример запроса на проведение оплаты через платежные приложения (Payme, Click, Uzumbank, Anorbank, Oson, Alif, Xazna, Beepul, Trastpay) без открытия промежуточной страницы оплаты Multicard. В ответе в поле checkout_url возвращается ссылка (Universal Link/App Link) на оплату в соответствующем платежном приложении, по которой необходимо перенаправить пользователя.

При создании платежа с помощью данного запроса, отправка запроса подтверждение не требуется. Платеж завершится в случае оплаты пользователем в соответствующем платежном приложении.

> Body Parameters

```json
{
    "payment_system": "payme",
    "amount": 50000,
    "store_id": 6,
    "invoice_id": "test",
    "billing_id": 1122334455,
    "callback_url": "https://",
    "ofd": [
        {
            "qty": 1,
            "vat": 12,
            "price": 60000000,
            "mxik": "06401004002000000",
            "total": 60000000,
            "package_code": "1506113",
            "name": "кроссовки men's low shoes"
        },
        {
            "qty": 1,
            "vat": 12,
            "price": 55700000,
            "mxik": "06401002004000000",
            "total": 55700000,
            "package_code": "1519041",
            "name": "кроссовки t.ace 2332 white black"
        }
    ]
}
```

### Params

|Name|Location|Type|Required|Description|
|---|---|---|---|---|
|body|body|object| no |none|
|» payment_system|body|string| yes |Платежная система через которую совершается платеж|
|» amount|body|integer| yes |Сумма платежа в тийинах|
|» store_id|body|integer| yes |ID кассы Партнера в системе Multicard|
|» invoice_id|body|string| yes |Max length:255|
|» billing_id|body|integer| no |Уникальный ID транзакции в системе Партнера|
|» callback_url|body|string| no |URL для отправки callback-запроса|
|» ofd|body|[object]| no |Данные для формирования фискального чека|
|»» qty|body|integer| yes |Количество единиц товара/услуги|
|»» vat|body|integer| no |НДС (%)|
|»» price|body|integer| yes |Стоимость единицы товара в тийинах|
|»» mxik|body|string| yes |ИКПУ из справочника tasnif.soliq.uz|
|»» total|body|integer| yes |Общая сумма товаров с учетом количества без учета скидок в тийинах|
|»» package_code|body|string| yes |Код упаковки из справочника tasnif.soliq.uz|
|»» name|body|string| yes |Наименование товара/услуги|
|»» tin|body|string| no |ИНН компании|
|»» mark|body|[string]| no |Массив с кодами маркировок каждой единицы товара. Обязателен для маркировочных товаров|

#### Description

**» invoice_id**: Max length:255
Любой идентификатор заказа в системе Партнера. Будет возвращен в callback-запросе. Также по нему можно искать платежи в кабинете мерчанта

#### Enum

|Name|Value|
|---|---|
|» payment_system|payme|
|» payment_system|click|
|» payment_system|uzum|
|» payment_system|anorbank|
|» payment_system|alif|
|» payment_system|oson|
|» payment_system|xazna|
|» payment_system|beepul|
|» payment_system|trastpay|
|» payment_system|sbp|

> Response Examples

> 400 Response

```json
{
    "success": false,
    "error": {
        "code": "ERROR_FIELDS",
        "details": "Значение «Payment System» неверно."
    }
}
```

> 200 Response

```json
{
  "success": true,
  "data": {
    "id": 0,
    "uuid": "095be615-a8ad-4c33-8e9c-c7612fbf6c9f",
    "store_id": 0,
    "payment_amount": 0,
    "commission_type": "up",
    "commission_amount": 0,
    "total_amount": 0,
    "store_invoice_id": "string",
    "status": "draft",
    "callback_url": "string",
    "billing_id": "string",
    "phone": "string",
    "ps": "uzcard",
    "receipt_url": "string",
    "kyc_data": {
      "last_name": "string",
      "first_name": "string",
      "middle_name": "string",
      "passport": "string",
      "dob": "string",
      "passport_expiry_date": "string"
    },
    "device_details": {
      "ip": "string",
      "user_agent": "string"
    },
    "details": {},
    "card_token": "string",
    "card_pan": "string",
    "split": [
      {
        "type": "account",
        "amount": 0,
        "details": "string",
        "recipient": "string"
      }
    ],
    "multicard_user_id": 0,
    "ofd": [
      {
        "qty": 0,
        "vat": 0,
        "price": 0,
        "mxik": "string",
        "total": 0,
        "package_code": "string",
        "name": "string",
        "tin": "string",
        "mark": [
          "string"
        ]
      }
    ],
    "terminal_id": "string",
    "merchant_id": "string",
    "ps_uniq_id": "string",
    "ps_response_code": "string",
    "ps_response_msg": "string",
    "callback_message": "string",
    "payment_time": "2019-08-24T14:15:22Z",
    "refund_time": "2019-08-24T14:15:22Z",
    "otp_hash": "string",
    "clearing_id": 0,
    "tax_receipt_id": 0,
    "push_sent_at": "2019-08-24T14:15:22Z",
    "store": {
      "id": 0,
      "uuid": "095be615-a8ad-4c33-8e9c-c7612fbf6c9f",
      "category_id": 0,
      "note": "string",
      "logo": "string",
      "color": "string",
      "view_fields": [
        {
          "type": "[",
          "name": "string",
          "value": {},
          "key": "string",
          "suggested": true
        }
      ],
      "tax_registration": 0,
      "tax_mxik": "string",
      "tax_package_code": "string",
      "tax_commission_recipient_tin": "string",
      "tg_chat_id": "string",
      "qr_url": "string",
      "bg_img": "string",
      "title": "string",
      "merchant": {
        "id": 0,
        "name": "string",
        "tin": "string",
        "contract_id": "string",
        "bank_account": "string"
      },
      "contract": {
        "id": 0,
        "num": "string",
        "date": "2019-08-24",
        "service": "string",
        "fee": {
          "up": null,
          "down": null
        },
        "edm_document_id": "string",
        "edm_status": "string"
      },
      "merchant_account": {
        "id": 0,
        "uuid": "095be615-a8ad-4c33-8e9c-c7612fbf6c9f",
        "tin": "string",
        "official_name": "string",
        "mfo": "string",
        "account_no": "string",
        "address": "string",
        "director": "string",
        "director_pinfl": "string",
        "vat_payer": true,
        "is_commitent": true,
        "active": true
      }
    },
    "application": {
      "id": 0,
      "application_id": "string",
      "wallet_sum": 0,
      "wallet_sender_account": "string",
      "wallet_overdraft": 0,
      "wallet_contract_num": "string",
      "otp_required": 0,
      "otp_gateway": "string",
      "sms_nickname": "string"
    },
    "tax": {
      "receipt": {},
      "f_num": "1d2db1d7-fdce-4ae6-9c43-6a437dcdbc89",
      "fm_num": "string",
      "qr_url": "string",
      "is_refund": true
    },
    "refund_tax": {
      "receipt": {},
      "f_num": "1d2db1d7-fdce-4ae6-9c43-6a437dcdbc89",
      "fm_num": "string",
      "qr_url": "string",
      "is_refund": true
    },
    "clearing": {
      "id": 0,
      "merchant": {
        "id": 0,
        "name": "string",
        "tin": "string",
        "contract_id": "string",
        "bank_account": "string"
      },
      "recipient_info": {
        "id": 0,
        "uuid": "095be615-a8ad-4c33-8e9c-c7612fbf6c9f",
        "tin": "string",
        "official_name": "string",
        "mfo": "string",
        "account_no": "string",
        "address": "string",
        "director": "string",
        "director_pinfl": "string",
        "vat_payer": true,
        "is_commitent": true,
        "active": true
      },
      "sender_info": {
        "id": 0,
        "uuid": "095be615-a8ad-4c33-8e9c-c7612fbf6c9f",
        "tin": "string",
        "official_name": "string",
        "mfo": "string",
        "account_no": "string",
        "address": "string",
        "director": "string",
        "director_pinfl": "string",
        "vat_payer": true,
        "is_commitent": true,
        "active": true
      },
      "purpose_code": "string",
      "amount": 0,
      "details": "string",
      "status": "new",
      "payment_time": "string",
      "added_on": "string",
      "updated_on": "string",
      "receipt_url": "string"
    },
    "checkout_url": "string",
    "added_on": "2019-08-24T14:15:22Z"
  }
}
```

### Responses

|HTTP Status Code |Meaning|Description|Data schema|
|---|---|---|---|
|200|[OK](https://tools.ietf.org/html/rfc7231#section-6.3.1)|none|Inline|
|400|[Bad Request](https://tools.ietf.org/html/rfc7231#section-6.5.1)|При отправке запроса с несуществующим типом внешних платежных систем, возвращается ошибка с кодом 400|Inline|

### Responses Data Schema

HTTP Status Code **200**

|Name|Type|Required|Restrictions|Title|description|
|---|---|---|---|---|---|
|» success|boolean|true|none||none|
|» data|[paymentModel](#schemapaymentmodel)|true|none||none|
|»» id|integer|true|none||ID транзакции в Multicard|
|»» uuid|string(uuid)|true|none||UUID транзакции в Multicard|
|»» store_id|integer|true|none||ID кассы, выданный со стороны Multicard|
|»» payment_amount|integer|true|none||none|
|»» commission_type|string|true|none||Тип комиссии|
|»» commission_amount|integer|true|none||none|
|»» total_amount|integer|true|none||Сумма платежа в тийинах|
|»» store_invoice_id|string|true|none||Любой идентификатор заказа в системе Партнера. Будет возвращен в callback-запросе. Также по нему можно искать платежаи в кабинете мерчанта|
|»» status|[PaymentStatusEnum](#schemapaymentstatusenum)|true|none||Статус транзакции|
|»» callback_url|string¦null|true|none||URL для отправки callback-запроса|
|»» billing_id|string¦null|true|none||Уникальный ID транзакции в системе Партнера|
|»» phone|string¦null|true|none||Телефон плательщика в формате 998XXXXXXXXX (при наличии)|
|»» ps|[PaymentServiceEnum](#schemapaymentserviceenum)|true|none||Платежный сервис/система|
|»» receipt_url|string¦null|true|none||Ссылка на платежный чек|
|»» kyc_data|object¦null|true|none||none|
|»»» last_name|string|true|none||Фамилия плательщика|
|»»» first_name|string|true|none||Имя плательщика|
|»»» middle_name|string|true|none||Отчество плательщика|
|»»» passport|string|true|none||Паспорт плательщика|
|»»» dob|string|true|none||Дата рождения плательщика (YYYY-MM-DD)|
|»»» passport_expiry_date|string|true|none||Дата истечения паспорта плательщика|
|»» device_details|object¦null|true|none||Объект с информацией об устройстве клиента|
|»»» ip|string|true|none||IP-адрес устройства клиента, с которого выполняется запрос|
|»»» user_agent|string|true|none||Cтрока User-Agent, содержащая информацию о браузере, операционной системе и типе устройства клиента|
|»» details|object¦null|true|none||Поля, необходимые для проведения платежа в биллинге. Используется только при оплате за услуги Paynet и МУНИС (мобильная связь, гос.платежи и т.п.). Список полей и их название зависит от конкретной услуги (store_id)|
|»» card_token|string¦null|true|none||none|
|»» card_pan|string¦null|true|none||none|
|»» split|[object]¦null|true|none||none|
|»»» type|string|true|none||Тип получателя|
|»»» amount|integer|true|none||Сумма платежа|
|»»» details|string|true|none||Детали платежа|
|»»» recipient|string|false|none||Получатель. Если type=account, то передается uuid банковских реквизитов. По-умолчанию подставляются банковские реквизиты мерчанта.|
|»» multicard_user_id|integer¦null|true|none||ID пользователя в приложении Multicard (если оплата проведена через приложение)|
|»» ofd|[object]¦null|true|none||none|
|»»» qty|integer|true|none||Количество единиц товара/услуги|
|»»» vat|integer|false|none||НДС (%)|
|»»» price|integer|true|none||Стоимость единицы товара в тийинах|
|»»» mxik|string|true|none||ИКПУ из справочника tasnif.soliq.uz|
|»»» total|integer|true|none||Общая сумма товаров с учетом количества без учета скидок в тийинах|
|»»» package_code|string|true|none||Код упаковки из справочника tasnif.soliq.uz|
|»»» name|string|true|none||Наименование товара/услуги|
|»»» tin|string|false|none||ИНН компании|
|»»» mark|[string]|false|none||Массив с кодами маркировок каждой единицы товара. Обязателен для маркировочных товаров|
|»» terminal_id|string¦null|true|none||ID терминала в платежной системе|
|»» merchant_id|string¦null|true|none||ID мерчанта в платежной системе|
|»» ps_uniq_id|string¦null|true|none||Reference number (RRN, RefNum) в платежной системе|
|»» ps_response_code|string¦null|true|none||Код ответа (ошибки) от платежной системы|
|»» ps_response_msg|string¦null|true|none||Описание ошибки от платежной системы|
|»» callback_message|string¦null|true|none||Текст ответа от биллинга мерчанта/поставщика|
|»» payment_time|string(date-time)¦null|true|none||Дата и время оплаты во временной зоне GMT+5 в формате YYYY-mm-dd H:s:s|
|»» refund_time|string(date-time)|true|none||Дата и время возврата во временной зоне GMT+5 в формате YYYY-mm-dd H:s:s|
|»» otp_hash|string¦null|true|none||Если значение не null, то требуется подтверждение транзакции через SMS-код (или 3DS)|
|»» clearing_id|integer¦null|true|none||none|
|»» tax_receipt_id|integer¦null|true|none||none|
|»» push_sent_at|string(date-time)¦null|true|none||none|
|»» store|[storeModel](#schemastoremodel)|true|none||none|
|»»» id|integer|true|none||none|
|»»» uuid|string(uuid)|true|none||none|
|»»» category_id|integer¦null|true|none||none|
|»»» note|string¦null|true|none||none|
|»»» logo|string¦null|true|none||Логотип|
|»»» color|string¦null|true|none||none|
|»»» view_fields|[object]¦null|true|none||none|
|»»»» type|string|true|none||Формат поля|
|»»»» name|string|true|none||Описание поля|
|»»»» value|any|true|none||Значение|

*oneOf*

|Name|Type|Required|Restrictions|Title|description|
|---|---|---|---|---|---|
|»»»»» *anonymous*|string|false|none||none|

*xor*

|Name|Type|Required|Restrictions|Title|description|
|---|---|---|---|---|---|
|»»»»» *anonymous*|integer|false|none||none|

*xor*

|Name|Type|Required|Restrictions|Title|description|
|---|---|---|---|---|---|
|»»»»» *anonymous*|boolean|false|none||none|

*xor*

|Name|Type|Required|Restrictions|Title|description|
|---|---|---|---|---|---|
|»»»»» *anonymous*|array|false|none||none|

*xor*

|Name|Type|Required|Restrictions|Title|description|
|---|---|---|---|---|---|
|»»»»» *anonymous*|object|false|none||none|

*xor*

|Name|Type|Required|Restrictions|Title|description|
|---|---|---|---|---|---|
|»»»»» *anonymous*|number|false|none||none|

*continued*

|Name|Type|Required|Restrictions|Title|description|
|---|---|---|---|---|---|
|»»»» key|string|true|none||Ключ|
|»»»» suggested|boolean¦null|true|none||Рекоммендуемая сумма оплаты|
|»»» tax_registration|integer|true|none||Флаг фискализации|
|»»» tax_mxik|string¦null|true|none||ИКПУ по умолчанию из каталога tasnif.soliq.uz|
|»»» tax_package_code|string¦null|true|none||Код упаковки от ИКПУ по умолчанию из каталога tasnif.soliq.uz|
|»»» tax_commission_recipient_tin|string¦null|true|none||ИНН комитента для фискализации. Если null, то берется ИНН мерчанта|
|»»» tg_chat_id|string¦null|true|none||ID телеграм группы для отправки уведомлений о платежах|
|»»» qr_url|string|true|none||Ссылка на QR-код для приема платежей по данной кассе|
|»»» bg_img|string|true|none||Фоновая картинка для страницы чекаута|
|»»» title|string|true|none||Наименование кассы|
|»»» merchant|[merchantModel](#schemamerchantmodel)|true|none||Информация о клиенте (мерчанте)|
|»»»» id|integer|true|none||ID клиента (мерчанта) в Multicard|
|»»»» name|string|true|none||Наименование клиента (мерчанта)|
|»»»» tin|string|true|none||ИНН мерчанта|
|»»»» contract_id|string¦null|true|none||Данные о договоре|
|»»»» bank_account|string¦null|true|none||Транзитный счет для расчетов|
|»»» contract|object¦null|true|none||Информация о контракте с мерчантом|
|»»»» id|integer|true|none||none|
|»»»» num|string|true|none||Номер договора|
|»»»» date|string(date)|true|none||Дата договора|
|»»»» service|string|true|none||none|
|»»»» fee|object|true|none||Комиссия по договору|
|»»»»» up|string|true|none||none|
|»»»»» down|string|true|none||none|
|»»»» edm_document_id|string¦null|true|none||Идентификатор документа в системе электронного документооборота|
|»»»» edm_status|string¦null|true|none||Статус подписания документа в системе электронного документооборота|
|»»» merchant_account|[merchantAccount](#schemamerchantaccount)|true|none||none|
|»»»» id|integer|true|none||ID банковских реквизитов|
|»»»» uuid|string(uuid)|true|none||UUID банковских реквизитов|
|»»»» tin|string|true|none||ИНН или ПИНФЛ|
|»»»» official_name|string|true|none||Наименование|
|»»»» mfo|string|true|none||МФО|
|»»»» account_no|string|true|none||Номер счета|
|»»»» address|string¦null|true|none||Юридический адрес|
|»»»» director|string¦null|true|none||ФИО директора|
|»»»» director_pinfl|string¦null|true|none||ПИНФЛ директора|
|»»»» vat_payer|boolean¦null|true|none||Является ли плательщиков НДС|
|»»»» is_commitent|boolean¦null|true|none||Является ли комитентом Multicard|
|»»»» active|boolean|true|none||Флаг активности. Если отключен, значит возмещение не будет отправлено|
|»» application|[applicationModel](#schemaapplicationmodel)|true|none||none|
|»»» id|integer|true|none||none|
|»»» application_id|string|true|none||none|
|»»» wallet_sum|integer¦null|true|none||none|
|»»» wallet_sender_account|string¦null|true|none||none|
|»»» wallet_overdraft|integer|true|none||none|
|»»» wallet_contract_num|string¦null|true|none||none|
|»»» otp_required|integer|true|none||none|
|»»» otp_gateway|string¦null|true|none||none|
|»»» sms_nickname|string¦null|true|none||none|
|»» tax|[taxReceiptModel](#schemataxreceiptmodel)¦null|true|none||none|
|»»» receipt|object¦null|true|none||Объект чека, отправленный в ГНК|
|»»» f_num|string(uuid)|true|none||Фискальный признак|
|»»» fm_num|string|true|none||Фискальный терминал|
|»»» qr_url|string¦null|true|none||URL на фискальный чек|
|»»» is_refund|boolean|true|none||Является ли чеком возврата|
|»» refund_tax|[taxReceiptModel](#schemataxreceiptmodel)¦null|true|none||none|
|»»» receipt|object¦null|true|none||Объект чека, отправленный в ГНК|
|»»» f_num|string(uuid)|true|none||Фискальный признак|
|»»» fm_num|string|true|none||Фискальный терминал|
|»»» qr_url|string¦null|true|none||URL на фискальный чек|
|»»» is_refund|boolean|true|none||Является ли чеком возврата|
|»» clearing|[clearingModel](#schemaclearingmodel)¦null|true|none||none|
|»»» id|integer|true|none||none|
|»»» merchant|[merchantModel](#schemamerchantmodel)|true|none||Информация о клиенте (мерчанте)|
|»»»» id|integer|true|none||ID клиента (мерчанта) в Multicard|
|»»»» name|string|true|none||Наименование клиента (мерчанта)|
|»»»» tin|string|true|none||ИНН мерчанта|
|»»»» contract_id|string¦null|true|none||Данные о договоре|
|»»»» bank_account|string¦null|true|none||Транзитный счет для расчетов|
|»»» recipient_info|[merchantAccount](#schemamerchantaccount)|true|none||none|
|»»»» id|integer|true|none||ID банковских реквизитов|
|»»»» uuid|string(uuid)|true|none||UUID банковских реквизитов|
|»»»» tin|string|true|none||ИНН или ПИНФЛ|
|»»»» official_name|string|true|none||Наименование|
|»»»» mfo|string|true|none||МФО|
|»»»» account_no|string|true|none||Номер счета|
|»»»» address|string¦null|true|none||Юридический адрес|
|»»»» director|string¦null|true|none||ФИО директора|
|»»»» director_pinfl|string¦null|true|none||ПИНФЛ директора|
|»»»» vat_payer|boolean¦null|true|none||Является ли плательщиков НДС|
|»»»» is_commitent|boolean¦null|true|none||Является ли комитентом Multicard|
|»»»» active|boolean|true|none||Флаг активности. Если отключен, значит возмещение не будет отправлено|
|»»» sender_info|[merchantAccount](#schemamerchantaccount)|true|none||none|
|»»»» id|integer|true|none||ID банковских реквизитов|
|»»»» uuid|string(uuid)|true|none||UUID банковских реквизитов|
|»»»» tin|string|true|none||ИНН или ПИНФЛ|
|»»»» official_name|string|true|none||Наименование|
|»»»» mfo|string|true|none||МФО|
|»»»» account_no|string|true|none||Номер счета|
|»»»» address|string¦null|true|none||Юридический адрес|
|»»»» director|string¦null|true|none||ФИО директора|
|»»»» director_pinfl|string¦null|true|none||ПИНФЛ директора|
|»»»» vat_payer|boolean¦null|true|none||Является ли плательщиков НДС|
|»»»» is_commitent|boolean¦null|true|none||Является ли комитентом Multicard|
|»»»» active|boolean|true|none||Флаг активности. Если отключен, значит возмещение не будет отправлено|
|»»» purpose_code|string|true|none||Код назначения платежа|
|»»» amount|integer|true|none||Сумма платежа в тийинах|
|»»» details|string|true|none||Детали платежа|
|»»» status|string|true|none||Статус|
|»»» payment_time|string|true|none||Время проведения платежа|
|»»» added_on|string|true|none||Дата создания записи|
|»»» updated_on|string|true|none||Дата изменения записи|
|»»» receipt_url|string|true|none||URL на банковскую квитанцию|
|»» checkout_url|string¦null|true|none||URL страницы для оплаты|
|»» added_on|string(date-time)|true|none||none|

#### Enum

|Name|Value|
|---|---|
|commission_type|up|
|commission_type|down|
|status|draft|
|status|progress|
|status|billing|
|status|success|
|status|error|
|status|revert|
|ps|uzcard|
|ps|humo|
|ps|visa|
|ps|mastercard|
|ps|account|
|ps|payme|
|ps|click|
|ps|uzum|
|ps|anorbank|
|ps|oson|
|ps|alif|
|ps|xazna|
|ps|beepul|
|ps|trastpay|
|ps|sbp|
|type|account|
|type|wallet|
|type|card|
|type|string|
|type|int|
|type|phone|
|type|tree|
|type|hidden|
|type|complex|
|type|select|
|tax_registration|0|
|tax_registration|1|
|tax_registration|2|
|tax_registration|3|
|status|new|
|status|sent|
|status|done|
|status|repeat|
|status|postponed|
|status|blocked|
|status|revert|
|status|canceled|

HTTP Status Code **400**

|Name|Type|Required|Restrictions|Title|description|
|---|---|---|---|---|---|
|» success|boolean|true|none||Статус запроса|
|» error|object|true|none||Ошибка|
|»» code|string|true|none||Кодовое сообщение|
|»» details|string|true|none||Описание ошибки|

## PUT Подтверждение платежа

PUT /payment/{payment_uuid}

В случае получения ошибки с кодом **ERROR_DEBIT_UNKNOWN** или **ERROR_CALLBACK_TIMEOUT**, необходимо проверять статус платежа пока не будет получен конечный статус (success или error)

> Body Parameters

```json
{
    "otp": "50112",
    "debit_available": false
}
```

### Params

|Name|Location|Type|Required|Description|
|---|---|---|---|---|
|payment_uuid|path|string| yes |none|
|body|body|object| no |none|
|» otp|body|string| yes |Одноразовый пароль для подтверждения операции|
|» debit_available|body|boolean| yes |Если передан true, в случае, если на карте |

#### Description

**» debit_available**: Если передан true, в случае, если на карте 
недостаточно средств, будет списана 
доступная на карте сумма

> Response Examples

> 200 Response

```json
{
    "success": true,
    "data": {PaymentModel}
}
```

> 400 Response

```json
{
    "success": false,
    "error": {
        "code": "ERROR_WRONG_OTP",
        "details": "Неверный код подтверждения"
    }
}
```

### Responses

|HTTP Status Code |Meaning|Description|Data schema|
|---|---|---|---|
|200|[OK](https://tools.ietf.org/html/rfc7231#section-6.3.1)|none|Inline|
|400|[Bad Request](https://tools.ietf.org/html/rfc7231#section-6.5.1)|При вводе некорректного код пароля, возвращается ошибку с кодом 400|Inline|

### Responses Data Schema

HTTP Status Code **400**

|Name|Type|Required|Restrictions|Title|description|
|---|---|---|---|---|---|
|» success|boolean|true|none||Статус запроса|
|» error|object|true|none||Ошибка|
|»» code|string|true|none||Кодовое сообщение|
|»» details|string|true|none||Описание ошибки|

## PATCH Отправка фискальной ссылки

PATCH /payment/{payment_uuid}/fiscal

Если фискализация платежа осуществляется на стороне Партнера, то необходима отправка фискальной ссылки в систему Multicard с помощью данного метода.

> Body Parameters

```json
{
    "url": "https://ofd.soliq.uz/..."
}
```

### Params

|Name|Location|Type|Required|Description|
|---|---|---|---|---|
|payment_uuid|path|string| yes |none|
|body|body|object| no |none|
|» url|body|string| yes |Ссылка на фискальный чек|
|» is_refund|body|boolean| no |Является возвратным чеком|

> Response Examples

> 200 Response

```json
{}
```

### Responses

|HTTP Status Code |Meaning|Description|Data schema|
|---|---|---|---|
|200|[OK](https://tools.ietf.org/html/rfc7231#section-6.3.1)|none|Inline|

### Responses Data Schema

## DELETE Отмена платежа (возврат средств)

DELETE /payment/{uuid}

> Body Parameters

```json
{}
```

### Params

|Name|Location|Type|Required|Description|
|---|---|---|---|---|
|uuid|path|string| yes |none|
|body|body|object| no |none|

> Response Examples

> 200 Response

```json
{
    "success": true,
    "data": {PaymentModel}
}
```

> 404 Response

```json
{
    "success": false,
    "error": {
        "code": "ERROR_NOT_FOUND",
        "details": "Объект не найден"
    }
}
```

### Responses

|HTTP Status Code |Meaning|Description|Data schema|
|---|---|---|---|
|200|[OK](https://tools.ietf.org/html/rfc7231#section-6.3.1)|none|Inline|
|404|[Not Found](https://tools.ietf.org/html/rfc7231#section-6.5.4)|При попытке отмены уже завершенного платежа, у которой оплата была проведена, возвращается ошибка с кодом 400|Inline|

### Responses Data Schema

HTTP Status Code **404**

|Name|Type|Required|Restrictions|Title|description|
|---|---|---|---|---|---|
|» success|boolean|true|none||Статус запроса|
|» error|object|true|none||Ошибка|
|»» code|string|true|none||Кодовое сообщение|
|»» details|string|true|none||Описание ошибки|

## GET Получение информации о платеже

GET /payment/{uuid}

В ответе возвращается {PaymentModel}

### Params

|Name|Location|Type|Required|Description|
|---|---|---|---|---|
|uuid|path|string| yes |none|

> Response Examples

> 200 Response

```json
{
    "success": true,
    "data": {PaymentModel}
}
```

### Responses

|HTTP Status Code |Meaning|Description|Data schema|
|---|---|---|---|
|200|[OK](https://tools.ietf.org/html/rfc7231#section-6.3.1)|none|Inline|

### Responses Data Schema

## DELETE Частичный возврат

DELETE /payment/{uuid}/partial

Для получения доступа к частичному возврату требуется настройка терминала со стороны Multicard

> Body Parameters

```json

```

```json
{
    "card_pan": "8600010000001234",
    "refund_amount": 22244100,
    "ofd": [
        {
            "name": "Бензин марки A80",
            "mxik": "02710001003000000",
            "package_code": 1282118,
            "price": 795000,
            "qty": 1,
            "total": 1605900
        }
    ]
}
```

### Params

|Name|Location|Type|Required|Description|
|---|---|---|---|---|
|uuid|path|string| yes |none|
|body|body|object| no |none|
|» card_pan|body|string| no |Полный номер карты для возврата. Обязателен в случае оплаты через платежное приложение (Payme, Click и т.д.). Если оплата напрямую с карты Uzcard/Humo/Visa/Mastercard, то передача не требуется|
|» ofd|body|[object]| yes |Данные для фискализации нового чека. Старый фискальный чек будет отменен.|
|»» qty|body|integer| yes |Количество единиц товара/услуги|
|»» vat|body|integer| no |НДС (%)|
|»» price|body|integer| yes |Стоимость единицы товара в тийинах|
|»» mxik|body|string| yes |ИКПУ из справочника tasnif.soliq.uz|
|»» total|body|integer| yes |Общая сумма товаров с учетом количества без учета скидок в тийинах|
|»» package_code|body|string| yes |Код упаковки из справочника tasnif.soliq.uz|
|»» name|body|string| yes |Наименование товара/услуги|
|»» tin|body|string| no |ИНН компании|
|»» mark|body|[string]| no |Массив с кодами маркировок каждой единицы товара. Обязателен для маркировочных товаров|
|» refund_amount|body|integer| yes |Сумма возврата (в тийинах)|

> Response Examples

> 200 Response

```json
{
    "success": true,
    "data": {PaymentModel}
}
```

> 404 Response

```json
{
    "success": false,
    "error": {
        "code": "ERROR_NOT_FOUND",
        "details": "Объект не найден"
    }
}
```

### Responses

|HTTP Status Code |Meaning|Description|Data schema|
|---|---|---|---|
|200|[OK](https://tools.ietf.org/html/rfc7231#section-6.3.1)|none|Inline|
|404|[Not Found](https://tools.ietf.org/html/rfc7231#section-6.5.4)|При попытке отмены уже завершенного платежа, у которой оплата была проведена, возвращается ошибка с кодом 400|Inline|

### Responses Data Schema

HTTP Status Code **200**

|Name|Type|Required|Restrictions|Title|description|
|---|---|---|---|---|---|
|» success|string|true|none||none|
|» data|[paymentModel](#schemapaymentmodel)|true|none||none|
|»» id|integer|true|none||ID транзакции в Multicard|
|»» uuid|string(uuid)|true|none||UUID транзакции в Multicard|
|»» store_id|integer|true|none||ID кассы, выданный со стороны Multicard|
|»» payment_amount|integer|true|none||none|
|»» commission_type|string|true|none||Тип комиссии|
|»» commission_amount|integer|true|none||none|
|»» total_amount|integer|true|none||Сумма платежа в тийинах|
|»» store_invoice_id|string|true|none||Любой идентификатор заказа в системе Партнера. Будет возвращен в callback-запросе. Также по нему можно искать платежаи в кабинете мерчанта|
|»» status|[PaymentStatusEnum](#schemapaymentstatusenum)|true|none||Статус транзакции|
|»» callback_url|string¦null|true|none||URL для отправки callback-запроса|
|»» billing_id|string¦null|true|none||Уникальный ID транзакции в системе Партнера|
|»» phone|string¦null|true|none||Телефон плательщика в формате 998XXXXXXXXX (при наличии)|
|»» ps|[PaymentServiceEnum](#schemapaymentserviceenum)|true|none||Платежный сервис/система|
|»» receipt_url|string¦null|true|none||Ссылка на платежный чек|
|»» kyc_data|object¦null|true|none||none|
|»»» last_name|string|true|none||Фамилия плательщика|
|»»» first_name|string|true|none||Имя плательщика|
|»»» middle_name|string|true|none||Отчество плательщика|
|»»» passport|string|true|none||Паспорт плательщика|
|»»» dob|string|true|none||Дата рождения плательщика (YYYY-MM-DD)|
|»»» passport_expiry_date|string|true|none||Дата истечения паспорта плательщика|
|»» device_details|object¦null|true|none||Объект с информацией об устройстве клиента|
|»»» ip|string|true|none||IP-адрес устройства клиента, с которого выполняется запрос|
|»»» user_agent|string|true|none||Cтрока User-Agent, содержащая информацию о браузере, операционной системе и типе устройства клиента|
|»» details|object¦null|true|none||Поля, необходимые для проведения платежа в биллинге. Используется только при оплате за услуги Paynet и МУНИС (мобильная связь, гос.платежи и т.п.). Список полей и их название зависит от конкретной услуги (store_id)|
|»» card_token|string¦null|true|none||none|
|»» card_pan|string¦null|true|none||none|
|»» split|[object]¦null|true|none||none|
|»»» type|string|true|none||Тип получателя|
|»»» amount|integer|true|none||Сумма платежа|
|»»» details|string|true|none||Детали платежа|
|»»» recipient|string|false|none||Получатель. Если type=account, то передается uuid банковских реквизитов. По-умолчанию подставляются банковские реквизиты мерчанта.|
|»» multicard_user_id|integer¦null|true|none||ID пользователя в приложении Multicard (если оплата проведена через приложение)|
|»» ofd|[object]¦null|true|none||none|
|»»» qty|integer|true|none||Количество единиц товара/услуги|
|»»» vat|integer|false|none||НДС (%)|
|»»» price|integer|true|none||Стоимость единицы товара в тийинах|
|»»» mxik|string|true|none||ИКПУ из справочника tasnif.soliq.uz|
|»»» total|integer|true|none||Общая сумма товаров с учетом количества без учета скидок в тийинах|
|»»» package_code|string|true|none||Код упаковки из справочника tasnif.soliq.uz|
|»»» name|string|true|none||Наименование товара/услуги|
|»»» tin|string|false|none||ИНН компании|
|»»» mark|[string]|false|none||Массив с кодами маркировок каждой единицы товара. Обязателен для маркировочных товаров|
|»» terminal_id|string¦null|true|none||ID терминала в платежной системе|
|»» merchant_id|string¦null|true|none||ID мерчанта в платежной системе|
|»» ps_uniq_id|string¦null|true|none||Reference number (RRN, RefNum) в платежной системе|
|»» ps_response_code|string¦null|true|none||Код ответа (ошибки) от платежной системы|
|»» ps_response_msg|string¦null|true|none||Описание ошибки от платежной системы|
|»» callback_message|string¦null|true|none||Текст ответа от биллинга мерчанта/поставщика|
|»» payment_time|string(date-time)¦null|true|none||Дата и время оплаты во временной зоне GMT+5 в формате YYYY-mm-dd H:s:s|
|»» refund_time|string(date-time)|true|none||Дата и время возврата во временной зоне GMT+5 в формате YYYY-mm-dd H:s:s|
|»» otp_hash|string¦null|true|none||Если значение не null, то требуется подтверждение транзакции через SMS-код (или 3DS)|
|»» clearing_id|integer¦null|true|none||none|
|»» tax_receipt_id|integer¦null|true|none||none|
|»» push_sent_at|string(date-time)¦null|true|none||none|
|»» store|[storeModel](#schemastoremodel)|true|none||none|
|»»» id|integer|true|none||none|
|»»» uuid|string(uuid)|true|none||none|
|»»» category_id|integer¦null|true|none||none|
|»»» note|string¦null|true|none||none|
|»»» logo|string¦null|true|none||Логотип|
|»»» color|string¦null|true|none||none|
|»»» view_fields|[object]¦null|true|none||none|
|»»»» type|string|true|none||Формат поля|
|»»»» name|string|true|none||Описание поля|
|»»»» value|any|true|none||Значение|

*oneOf*

|Name|Type|Required|Restrictions|Title|description|
|---|---|---|---|---|---|
|»»»»» *anonymous*|string|false|none||none|

*xor*

|Name|Type|Required|Restrictions|Title|description|
|---|---|---|---|---|---|
|»»»»» *anonymous*|integer|false|none||none|

*xor*

|Name|Type|Required|Restrictions|Title|description|
|---|---|---|---|---|---|
|»»»»» *anonymous*|boolean|false|none||none|

*xor*

|Name|Type|Required|Restrictions|Title|description|
|---|---|---|---|---|---|
|»»»»» *anonymous*|array|false|none||none|

*xor*

|Name|Type|Required|Restrictions|Title|description|
|---|---|---|---|---|---|
|»»»»» *anonymous*|object|false|none||none|

*xor*

|Name|Type|Required|Restrictions|Title|description|
|---|---|---|---|---|---|
|»»»»» *anonymous*|number|false|none||none|

*continued*

|Name|Type|Required|Restrictions|Title|description|
|---|---|---|---|---|---|
|»»»» key|string|true|none||Ключ|
|»»»» suggested|boolean¦null|true|none||Рекоммендуемая сумма оплаты|
|»»» tax_registration|integer|true|none||Флаг фискализации|
|»»» tax_mxik|string¦null|true|none||ИКПУ по умолчанию из каталога tasnif.soliq.uz|
|»»» tax_package_code|string¦null|true|none||Код упаковки от ИКПУ по умолчанию из каталога tasnif.soliq.uz|
|»»» tax_commission_recipient_tin|string¦null|true|none||ИНН комитента для фискализации. Если null, то берется ИНН мерчанта|
|»»» tg_chat_id|string¦null|true|none||ID телеграм группы для отправки уведомлений о платежах|
|»»» qr_url|string|true|none||Ссылка на QR-код для приема платежей по данной кассе|
|»»» bg_img|string|true|none||Фоновая картинка для страницы чекаута|
|»»» title|string|true|none||Наименование кассы|
|»»» merchant|[merchantModel](#schemamerchantmodel)|true|none||Информация о клиенте (мерчанте)|
|»»»» id|integer|true|none||ID клиента (мерчанта) в Multicard|
|»»»» name|string|true|none||Наименование клиента (мерчанта)|
|»»»» tin|string|true|none||ИНН мерчанта|
|»»»» contract_id|string¦null|true|none||Данные о договоре|
|»»»» bank_account|string¦null|true|none||Транзитный счет для расчетов|
|»»» contract|object¦null|true|none||Информация о контракте с мерчантом|
|»»»» id|integer|true|none||none|
|»»»» num|string|true|none||Номер договора|
|»»»» date|string(date)|true|none||Дата договора|
|»»»» service|string|true|none||none|
|»»»» fee|object|true|none||Комиссия по договору|
|»»»»» up|string|true|none||none|
|»»»»» down|string|true|none||none|
|»»»» edm_document_id|string¦null|true|none||Идентификатор документа в системе электронного документооборота|
|»»»» edm_status|string¦null|true|none||Статус подписания документа в системе электронного документооборота|
|»»» merchant_account|[merchantAccount](#schemamerchantaccount)|true|none||none|
|»»»» id|integer|true|none||ID банковских реквизитов|
|»»»» uuid|string(uuid)|true|none||UUID банковских реквизитов|
|»»»» tin|string|true|none||ИНН или ПИНФЛ|
|»»»» official_name|string|true|none||Наименование|
|»»»» mfo|string|true|none||МФО|
|»»»» account_no|string|true|none||Номер счета|
|»»»» address|string¦null|true|none||Юридический адрес|
|»»»» director|string¦null|true|none||ФИО директора|
|»»»» director_pinfl|string¦null|true|none||ПИНФЛ директора|
|»»»» vat_payer|boolean¦null|true|none||Является ли плательщиков НДС|
|»»»» is_commitent|boolean¦null|true|none||Является ли комитентом Multicard|
|»»»» active|boolean|true|none||Флаг активности. Если отключен, значит возмещение не будет отправлено|
|»» application|[applicationModel](#schemaapplicationmodel)|true|none||none|
|»»» id|integer|true|none||none|
|»»» application_id|string|true|none||none|
|»»» wallet_sum|integer¦null|true|none||none|
|»»» wallet_sender_account|string¦null|true|none||none|
|»»» wallet_overdraft|integer|true|none||none|
|»»» wallet_contract_num|string¦null|true|none||none|
|»»» otp_required|integer|true|none||none|
|»»» otp_gateway|string¦null|true|none||none|
|»»» sms_nickname|string¦null|true|none||none|
|»» tax|[taxReceiptModel](#schemataxreceiptmodel)¦null|true|none||none|
|»»» receipt|object¦null|true|none||Объект чека, отправленный в ГНК|
|»»» f_num|string(uuid)|true|none||Фискальный признак|
|»»» fm_num|string|true|none||Фискальный терминал|
|»»» qr_url|string¦null|true|none||URL на фискальный чек|
|»»» is_refund|boolean|true|none||Является ли чеком возврата|
|»» refund_tax|[taxReceiptModel](#schemataxreceiptmodel)¦null|true|none||none|
|»»» receipt|object¦null|true|none||Объект чека, отправленный в ГНК|
|»»» f_num|string(uuid)|true|none||Фискальный признак|
|»»» fm_num|string|true|none||Фискальный терминал|
|»»» qr_url|string¦null|true|none||URL на фискальный чек|
|»»» is_refund|boolean|true|none||Является ли чеком возврата|
|»» clearing|[clearingModel](#schemaclearingmodel)¦null|true|none||none|
|»»» id|integer|true|none||none|
|»»» merchant|[merchantModel](#schemamerchantmodel)|true|none||Информация о клиенте (мерчанте)|
|»»»» id|integer|true|none||ID клиента (мерчанта) в Multicard|
|»»»» name|string|true|none||Наименование клиента (мерчанта)|
|»»»» tin|string|true|none||ИНН мерчанта|
|»»»» contract_id|string¦null|true|none||Данные о договоре|
|»»»» bank_account|string¦null|true|none||Транзитный счет для расчетов|
|»»» recipient_info|[merchantAccount](#schemamerchantaccount)|true|none||none|
|»»»» id|integer|true|none||ID банковских реквизитов|
|»»»» uuid|string(uuid)|true|none||UUID банковских реквизитов|
|»»»» tin|string|true|none||ИНН или ПИНФЛ|
|»»»» official_name|string|true|none||Наименование|
|»»»» mfo|string|true|none||МФО|
|»»»» account_no|string|true|none||Номер счета|
|»»»» address|string¦null|true|none||Юридический адрес|
|»»»» director|string¦null|true|none||ФИО директора|
|»»»» director_pinfl|string¦null|true|none||ПИНФЛ директора|
|»»»» vat_payer|boolean¦null|true|none||Является ли плательщиков НДС|
|»»»» is_commitent|boolean¦null|true|none||Является ли комитентом Multicard|
|»»»» active|boolean|true|none||Флаг активности. Если отключен, значит возмещение не будет отправлено|
|»»» sender_info|[merchantAccount](#schemamerchantaccount)|true|none||none|
|»»»» id|integer|true|none||ID банковских реквизитов|
|»»»» uuid|string(uuid)|true|none||UUID банковских реквизитов|
|»»»» tin|string|true|none||ИНН или ПИНФЛ|
|»»»» official_name|string|true|none||Наименование|
|»»»» mfo|string|true|none||МФО|
|»»»» account_no|string|true|none||Номер счета|
|»»»» address|string¦null|true|none||Юридический адрес|
|»»»» director|string¦null|true|none||ФИО директора|
|»»»» director_pinfl|string¦null|true|none||ПИНФЛ директора|
|»»»» vat_payer|boolean¦null|true|none||Является ли плательщиков НДС|
|»»»» is_commitent|boolean¦null|true|none||Является ли комитентом Multicard|
|»»»» active|boolean|true|none||Флаг активности. Если отключен, значит возмещение не будет отправлено|
|»»» purpose_code|string|true|none||Код назначения платежа|
|»»» amount|integer|true|none||Сумма платежа в тийинах|
|»»» details|string|true|none||Детали платежа|
|»»» status|string|true|none||Статус|
|»»» payment_time|string|true|none||Время проведения платежа|
|»»» added_on|string|true|none||Дата создания записи|
|»»» updated_on|string|true|none||Дата изменения записи|
|»»» receipt_url|string|true|none||URL на банковскую квитанцию|
|»» checkout_url|string¦null|true|none||URL страницы для оплаты|
|»» added_on|string(date-time)|true|none||none|

#### Enum

|Name|Value|
|---|---|
|commission_type|up|
|commission_type|down|
|status|draft|
|status|progress|
|status|billing|
|status|success|
|status|error|
|status|revert|
|ps|uzcard|
|ps|humo|
|ps|visa|
|ps|mastercard|
|ps|account|
|ps|payme|
|ps|click|
|ps|uzum|
|ps|anorbank|
|ps|oson|
|ps|alif|
|ps|xazna|
|ps|beepul|
|ps|trastpay|
|ps|sbp|
|type|account|
|type|wallet|
|type|card|
|type|string|
|type|int|
|type|phone|
|type|tree|
|type|hidden|
|type|complex|
|type|select|
|tax_registration|0|
|tax_registration|1|
|tax_registration|2|
|tax_registration|3|
|status|new|
|status|sent|
|status|done|
|status|repeat|
|status|postponed|
|status|blocked|
|status|revert|
|status|canceled|

HTTP Status Code **404**

|Name|Type|Required|Restrictions|Title|description|
|---|---|---|---|---|---|
|» success|boolean|true|none||Статус запроса|
|» error|object|true|none||Ошибка|
|»» code|string|true|none||Кодовое сообщение|
|»» details|string|true|none||Описание ошибки|

# Холдирование

## POST Создать заявку на холдирование

POST /payment/hold

> Body Parameters

```json
{
    "card": {
		"token": "6225f3c93f7a880142782fa4"
    },
    "amount": 10000000,
    "store_id": 15,
    "invoice_id": "18177",
    "expiry": 43200
}
```

### Params

|Name|Location|Type|Required|Description|
|---|---|---|---|---|
|body|body|object| no |none|
|» card|body|object| yes |none|
|»» token|body|string| yes |Max length:255|
|» amount|body|integer| yes |Сумма платежа в тийинах|
|» store_id|body|integer| yes |ID (int) или UUID (string) кассы, выданный со стороны Multicard|
|» invoice_id|body|string| yes |Max length:255|
|» expiry|body|integer| yes |Minimum:1|
|» split|body|[object]| no |none|
|»» type|body|string| yes |Тип получателя|
|»» amount|body|integer| yes |Сумма платежа|
|»» details|body|string| yes |Детали платежа|
|»» recipient|body|string| no |Получатель. Если type=account, то передается uuid банковских реквизитов. По-умолчанию подставляются банковские реквизиты мерчанта.|

#### Description

**»» token**: Max length:255
Токен карты для проведения платежных операций с картой

**» invoice_id**: Max length:255
Любой идентификатор заказа в системе Партнера. Будет возвращен в callback-запросе. Также по нему можно искать платежи в кабинете мерчанта

**» expiry**: Minimum:1
Maximum:43200
Срок действия холдирования в минутах. После этого времени средства будут возвращены на карту. Допустимо не более 30 дней (43200)

#### Enum

|Name|Value|
|---|---|
|»» type|account|
|»» type|wallet|
|»» type|card|

> Response Examples

> 200 Response

```json
{
    "success": true,
    "data": {
        "id": 753967,
        "status": "draft",
        "expiry": "2025-03-08 23:12:15",
        "added_on": "2025-03-07 23:12:15",
        "updated_on": null,
        "payment": {PaymentModel}
    }
}
```

> При отправке запроса с допущением ошибки валидации, возвращается ошибка с кодом 400

```json
{
    "success": false,
    "error": {
        "code": "ERROR_FIELDS",
        "details": "Необходимо заполнить «Amount»."
    }
}
```

```json
{
    "success": false,
    "error": {
        "code": "ERROR_FIELDS",
        "details": "Необходимо заполнить «Amount»."
    }
}
```

### Responses

|HTTP Status Code |Meaning|Description|Data schema|
|---|---|---|---|
|200|[OK](https://tools.ietf.org/html/rfc7231#section-6.3.1)|none|Inline|
|400|[Bad Request](https://tools.ietf.org/html/rfc7231#section-6.5.1)|При отправке запроса с допущением ошибки валидации, возвращается ошибка с кодом 400|Inline|

### Responses Data Schema

HTTP Status Code **400**

|Name|Type|Required|Restrictions|Title|description|
|---|---|---|---|---|---|
|» success|boolean|true|none||Статус запроса|
|» error|object|true|none||Ошибка|
|»» code|string|true|none||Кодовое сообщение|
|»» details|string|true|none||Описание ошибки|

## PUT Подтвердить холдирование (блокировка средств на карте)

PUT /payment/hold/{payment_uuid}

> Body Parameters

```json
{
    "otp": "1122"
}
```

### Params

|Name|Location|Type|Required|Description|
|---|---|---|---|---|
|payment_uuid|path|string| yes |Идентификатор созданного платежа|
|body|body|object| no |none|
|» otp|body|string| yes |Одноразовый пароль для подтверждения операции|

> Response Examples

> 200 Response

```json
{
    "success": true,
    "data": {
        "id": 753967,
        "status": "active",
        "expiry": "2025-03-08 23:12:15",
        "added_on": "2025-03-07 23:12:15",
        "updated_on": null,
        "payment": {PaymentModel}
    }
}
```

### Responses

|HTTP Status Code |Meaning|Description|Data schema|
|---|---|---|---|
|200|[OK](https://tools.ietf.org/html/rfc7231#section-6.3.1)|none|Inline|

### Responses Data Schema

## GET Получение информации о холдировании

GET /payment/hold/{payment_uuid}

### Params

|Name|Location|Type|Required|Description|
|---|---|---|---|---|
|payment_uuid|path|string| yes |Идентификатор созданного платежа|

> Response Examples

> 200 Response

```json
{
    "success": true,
    "data": {
        "id": 753967,
        "status": "draft",
        "expiry": "2025-03-08 23:12:15",
        "added_on": "2025-03-07 23:12:15",
        "updated_on": null,
        "payment": {PaymentModel}
    }
}
```

### Responses

|HTTP Status Code |Meaning|Description|Data schema|
|---|---|---|---|
|200|[OK](https://tools.ietf.org/html/rfc7231#section-6.3.1)|none|Inline|

### Responses Data Schema

## DELETE Отмена захолдированных средств (до списания)

DELETE /payment/hold/{payment_uuid}

Метод позволяет вернуть (разблокировать) захолдированные средства на карту клиента до срока истечения холдирования

### Params

|Name|Location|Type|Required|Description|
|---|---|---|---|---|
|payment_uuid|path|string| yes |Идентификатор созданного платежа|

> Response Examples

> 200 Response

```json
{
    "success": true,
    "data": {
        "id": 753967,
        "status": "canceled",
        "expiry": "2025-03-08 23:12:15",
        "added_on": "2025-03-07 23:12:15",
        "updated_on": null,
        "payment": {PaymentModel}
    }
}
```

> 403 Response

```json
{
    "success": false,
    "error": {
        "code": "ERROR_PAYMENT_APPLIED",
        "details": "Холд не активен"
    }
}
```

### Responses

|HTTP Status Code |Meaning|Description|Data schema|
|---|---|---|---|
|200|[OK](https://tools.ietf.org/html/rfc7231#section-6.3.1)|none|Inline|
|403|[Forbidden](https://tools.ietf.org/html/rfc7231#section-6.5.3)|При попытке отмены неактивного холда, и завершенную оплатой, возвращается ошибка 400|Inline|

### Responses Data Schema

HTTP Status Code **403**

|Name|Type|Required|Restrictions|Title|description|
|---|---|---|---|---|---|
|» success|boolean|true|none||Статус запроса|
|» error|object|true|none||Ошибка|
|»» code|string|true|none||Кодовое сообщение|
|»» details|string|true|none||Описание ошибки|

## PUT Списать захолдированные средства

PUT /payment/hold/{payment_uuid}/charge

Метод осуществляет финальное списание захолдированных средств. Допустимо передать в теле сумму списания (должна быть меньше, чем сумма холдирования), иначе будет списана вся захолдированная сумма.

Также в запросе можно передать массивы split и ofd.

В случае получения ошибки с кодом **ERROR_DEBIT_UNKNOWN** или **ERROR_CALLBACK_TIMEOUT**, необходимо проверять статус платежа пока не будет получен конечный статус (success или error)

> Body Parameters

```json
{
    "amount": 100000
}
```

### Params

|Name|Location|Type|Required|Description|
|---|---|---|---|---|
|payment_uuid|path|string| yes |Идентификатор созданного платежа|
|body|body|object| no |none|
|» amount|body|integer| yes |Сумма перевода в тийинах|

> Response Examples

> 200 Response

```json
{
    "success": true,
    "data": {PaymentModel}
}
```

### Responses

|HTTP Status Code |Meaning|Description|Data schema|
|---|---|---|---|
|200|[OK](https://tools.ietf.org/html/rfc7231#section-6.3.1)|none|Inline|

### Responses Data Schema

# Выплаты на карту (payouts)

## POST Создание платежа на выплату c передачей номера карты

POST /payment/credit

(!) При получении ошибки с кодом «ERROR_UNKNOWN» или таймаута на запрос, необходимо проверять статус транзакции до тех пор, пока не получен конечный (error или success)

> Body Parameters

```json
{
    "card": {
        "pan": "8600303655375959"
    },
    "amount": 10000,
    "store_id": 6,
    "invoice_id": "112233",
    "confirmable": true,
    "device_details": {
        "ip": "182.19.100.10",
        "user_agent": "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/128.0.0.0 Safari/537.36"
    },
    "kyc_data": {
        "last_name": "BOLTAYEV",
        "first_name": "ALISHER",
        "middle_name": "MUMINOVICH",
        "passport": "AD21234567",
        "pinfl": "31105892514010",
        "dob": "1989-05-11",
        "passport_expiry_date": "2025-11-17"
    }
}
```

### Params

|Name|Location|Type|Required|Description|
|---|---|---|---|---|
|body|body|object| no |none|
|» card|body|object| yes |none|
|»» pan|body|string| no |Номер карты (Uzcard или Humo)|
|»» token|body|string| no |Токен карты, полученный при првязке|
|» amount|body|integer| yes |Сумма платежа в тийинах|
|» store_id|body|string| yes |ID кассы Партнера в системе Multicard|
|» invoice_id|body|string| yes |Max length:255|
|» confirmable|body|boolean| yes |Если передать false или опустить данное поле, то пополнение карты пройдет одним запросом (без подтверждения)|
|» device_details|body|object| yes |Объект с информацией об устройстве клиента|
|»» ip|body|string| yes |IP-адрес устройства клиента, с которого выполняется запрос|
|»» user_agent|body|string| yes |Cтрока User-Agent, содержащая информацию о браузере, операционной системе и типе устройства клиента|
|» kyc_data|body|object| yes |Поле обязательно к заполнению в случае пополнения более 10 млн.сум|
|»» last_name|body|string| yes |Фамилия владельца карты|
|»» first_name|body|string| yes |Имя владельца карты|
|»» middle_name|body|string| yes |Отчество владельца карты|
|»» passport|body|string| yes |Паспорт владельца карты|
|»» pinfl|body|string| yes |ПИНФЛ владельца карты|
|»» dob|body|string| yes |Дата рождения владельца карты (YYYY-MM-DD)|
|»» passport_expiry_date|body|string| yes |Дата истечения паспорта владельца карты|

#### Description

**» invoice_id**: Max length:255
Любой идентификатор заказа в системе Партнера. Будет возвращен в callback-запросе. Также по нему можно искать платежи в кабинете мерчанта

> Response Examples

```json
{
    "success": true,
    "data": {
        "ps": "humo",
        "store_id": 6,
        "store_invoice_id": "112233",
        "card_token": null,
        "card_pan": "409784******3066",
        "status": "draft",
        "callback_url": null,
        "phone": null,
        "kyc_data": "{\"last_name\":\"BOLTAYEV\",\"first_name\":\"ALISHER\",\"middle_name\":\"MUMINOVICH\",\"passport\":\"AD21234567\",\"pinfl\":\"31105892514010\",\"dob\":\"1989-05-11\",\"passport_expiry_date\":\"2025-11-17\"}",
        "terminal_id": "186368T6",
        "merchant_id": "01200000119700D",
        "commission_type": "up",
        "total_amount": 10000,
        "payment_amount": 10000,
        "commission_amount": 0,
        "id": 11351395,
        "uuid": "a81b7b62-6d86-11f0-9a1b-00505680eaf6",
        "multicard_user_id": null,
        "ps_id": null,
        "ps_uniq_id": null,
        "ps_response_code": null,
        "payment_time": null,
        "added_on": "2025-07-31 01:49:16",
        "store": {StoreModel},
        "receipt_url": null,
        "ps_response_msg": null
    }
}
```

```json
{
    "success": false,
    "error": {
        "code": "ERROR_CARD_NOT_FOUND",
        "details": "Неверный номер карты или срок действия"
    }
}
```

```json
{
    "success": false,
    "error": {
        "code": "ERROR_FIELDS",
        "details": "Для пополнения на эту сумму необходимо пройти идентификацию"
    }
}
```

> 400 Response

```json
{
  "success": true,
  "error": {
    "code": "string",
    "details": "string"
  }
}
```

### Responses

|HTTP Status Code |Meaning|Description|Data schema|
|---|---|---|---|
|200|[OK](https://tools.ietf.org/html/rfc7231#section-6.3.1)|none|Inline|
|400|[Bad Request](https://tools.ietf.org/html/rfc7231#section-6.5.1)|При попытке добавления карты с неправильным номером, система возвращает ошибку с кодом 400|Inline|

### Responses Data Schema

HTTP Status Code **400**

|Name|Type|Required|Restrictions|Title|description|
|---|---|---|---|---|---|
|» success|boolean|true|none||Статус запроса|
|» error|object|true|none||Ошибка|
|»» code|string|true|none||Кодовое сообщение|
|»» details|string|true|none||Описание ошибки|

## PUT Подтвердить выплату

PUT /payment/credit/{payment_uuid}

(!) При получении ошибки с кодом «ERROR_UNKNOWN» или таймаута на запрос, необходимо проверять статус транзакции до тех пор, пока не получен конечный (error или success).

> Body Parameters

```json
{
    "otp": "112233"
}
```

### Params

|Name|Location|Type|Required|Description|
|---|---|---|---|---|
|payment_uuid|path|string| yes |none|
|body|body|object| no |none|
|» otp|body|string| yes |Одноразовый пароль для подтверждения операции|

> Response Examples

> 200 Response

```json
{
    "success": true,
    "data": {
        "id": 11351395,
        "uuid": "a81b7b62-6d86-11f0-9a1b-00505680eaf6",
        "store_id": 6,
        "multicard_user_id": null,
        "ps": "humo",
        "store_invoice_id": "112233",
        "merchant_id": "01200000119700D",
        "terminal_id": "186368T6",
        "payment_amount": 10000,
        "commission_amount": 0,
        "commission_type": "up",
        "total_amount": 10000,
        "phone": null,
        "card_token": null,
        "card_pan": "409784******3066",
        "ps_id": "5002050246",
        "ps_uniq_id": "521120956011",
        "status": "success",
        "ps_response_code": null,
        "payment_time": "2025-07-31 01:51:12",
        "callback_url": null,
        "kyc_data": "{\"last_name\":\"BOLTAYEV\",\"first_name\":\"ALISHER\",\"middle_name\":\"MUMINOVICH\",\"passport\":\"AD21234567\",\"pinfl\":\"31105892514010\",\"dob\":\"1989-05-11\",\"passport_expiry_date\":\"2025-11-17\"}",
        "added_on": "2025-07-31 01:49:16",
        "store": {StoreModel},
        "receipt_url": "https://checkout.multicard.uz/credit/check/a81b7b62-6d86-11f0-9a1b-00505680eaf6",
        "ps_response_msg": null
    }
}
```

### Responses

|HTTP Status Code |Meaning|Description|Data schema|
|---|---|---|---|
|200|[OK](https://tools.ietf.org/html/rfc7231#section-6.3.1)|none|Inline|

### Responses Data Schema

## GET Получение информации о выплате

GET /payment/credit/{payment_uuid}

### Params

|Name|Location|Type|Required|Description|
|---|---|---|---|---|
|payment_uuid|path|string| yes |none|

> Response Examples

> 200 Response

```json
{
    "success": true,
    "data": {
        "id": 11351395,
        "uuid": "a81b7b62-6d86-11f0-9a1b-00505680eaf6",
        "store_id": 6,
        "multicard_user_id": null,
        "ps": "humo",
        "store_invoice_id": "112233",
        "merchant_id": "01200000119700D",
        "terminal_id": "186368T6",
        "payment_amount": 10000,
        "commission_amount": 0,
        "commission_type": "up",
        "total_amount": 10000,
        "phone": null,
        "card_token": null,
        "card_pan": "409784******3066",
        "ps_id": "5002050246",
        "ps_uniq_id": "521120956011",
        "status": "success",
        "ps_response_code": null,
        "payment_time": "2025-07-31 01:51:12",
        "callback_url": null,
        "kyc_data": "{\"last_name\":\"BOLTAYEV\",\"first_name\":\"ALISHER\",\"middle_name\":\"MUMINOVICH\",\"passport\":\"AD21234567\",\"pinfl\":\"31105892514010\",\"dob\":\"1989-05-11\",\"passport_expiry_date\":\"2025-11-17\"}",
        "added_on": "2025-07-31 01:49:16",
        "store": {StoreModel},
        "receipt_url": "https://checkout.multicard.uz/credit/check/a81b7b62-6d86-11f0-9a1b-00505680eaf6",
        "ps_response_msg": null
    }
}
```

### Responses

|HTTP Status Code |Meaning|Description|Data schema|
|---|---|---|---|
|200|[OK](https://tools.ietf.org/html/rfc7231#section-6.3.1)|none|Inline|

### Responses Data Schema

# Дополнительные методы

## GET Информация о приложении

GET /payment/application

> Response Examples

> 200 Response

```json
{
    "success": true,
    "data": {
        "id": 9,
        "application_id": "test-app",
        "wallet_sum": 34643176,
        "wallet_sender_account": null,
        "wallet_overdraft": 10000000,
        "wallet_contract_num": null,
        "otp_required": 1,
        "otp_gateway": null,
        "sms_nickname": null,
        "allow_bank_transaction": 1,
        "otp_length": 6,
        "official_name": "Multicard",
        "offer_url": null,
        "phone": null,
        "added_on": "2021-08-12 00:58:02",
        "updated_on": "2024-08-21 10:00:26"
    }
}
```

### Responses

|HTTP Status Code |Meaning|Description|Data schema|
|---|---|---|---|
|200|[OK](https://tools.ietf.org/html/rfc7231#section-6.3.1)|none|Inline|

### Responses Data Schema

HTTP Status Code **200**

|Name|Type|Required|Restrictions|Title|description|
|---|---|---|---|---|---|
|» success|boolean|true|none||none|
|» data|object|true|none||none|
|»» id|integer|true|none||none|
|»» application_id|string|true|none||Идентификатор приложения, выданный со стороны Multicard|
|»» wallet_sum|integer|true|none||Баланс депозита (кошелька) в тийинах|
|»» wallet_sender_account|null|true|none||none|
|»» wallet_overdraft|integer|true|none||Сумма ухода "в минус" (в тийинах)|
|»» wallet_contract_num|null|true|none||none|
|»» otp_required|integer|true|none||none|
|»» otp_gateway|null|true|none||none|
|»» sms_nickname|null|true|none||none|
|»» allow_bank_transaction|integer|true|none||none|
|»» otp_length|integer|true|none||Длина одноразового пароля (OTP)|
|»» official_name|string|true|none||none|
|»» offer_url|null|true|none||none|
|»» phone|null|true|none||none|
|»» added_on|string|true|none||none|
|»» updated_on|string|true|none||none|

## GET Реквизиты получателя

GET /payment/merchant-account/{recipient}

Получить информацию о банковских реквизитах

### Params

|Name|Location|Type|Required|Description|
|---|---|---|---|---|
|recipient|path|string| yes |none|

> Response Examples

> 200 Response

```json
{
    "success": true,
    "data": {
        "id": 174,
        "uuid": "540311d7-4e1d-11ee-8bac-00505680eaf6",
        "merchant_id": 13,
        "tin": "307578794",
        "official_name": "\"MULTICARD PAYMENT\" AJ",
        "mfo": "00491",
        "account_no": "20208000905255407001",
        "address": "г. Ташкент, Яшнабадский район, ул. Содик Азимов, Туйтепа МСГ, 50- Дом, -  ",
        "director": "ZAKIROV RUSTAM RINATOVICH",
        "director_pinfl": "32307890270337",
        "vat_payer": true,
        "is_commitent": true,
        "contract_no": null,
        "contract_date": null,
        "active": 1,
        "tax_contract_begin": "2024-09-20",
        "tax_contract_end": "2050-01-01"
    }
}
```

### Responses

|HTTP Status Code |Meaning|Description|Data schema|
|---|---|---|---|
|200|[OK](https://tools.ietf.org/html/rfc7231#section-6.3.1)|none|Inline|

### Responses Data Schema

HTTP Status Code **200**

|Name|Type|Required|Restrictions|Title|description|
|---|---|---|---|---|---|
|» success|boolean|true|none||none|
|» data|object|true|none||none|
|»» id|integer|true|none||none|
|»» uuid|string|true|none||Уникальный идентификатор транзакции в Multicard|
|»» merchant_id|integer|true|none||Идентификатор мерчанта ePOS-терминала|
|»» tin|string|true|none||ИНН компании|
|»» official_name|string|true|none||Официальное юридическое наименование организации|
|»» mfo|string|true|none||Банковский идентификационный код (МФО), который указывает на конкретное финансовое учреждение|
|»» account_no|string|true|none||Расчётный счёт организации в банке (р/с), используемый для проведения платежей и расчётов|
|»» address|string|true|none||Юридический адрес организации (город, район, улица, дом и т.д.)|
|»» director|string|true|none||Руководитель организации (ФИО директора)|
|»» director_pinfl|string|true|none||ПИНФЛ директора|
|»» vat_payer|boolean|true|none||Признак плательщика НДС (если true — организация зарегистрирована как плательщик налога на добавленную стоимость)|
|»» is_commitent|boolean|true|none||Признак того, что организация выступает комитентом (если true — организация работает по договору комиссии)|
|»» contract_no|null|true|none||Номер договора|
|»» contract_date|null|true|none||Дата заключения договора|
|»» active|integer|true|none||Статус договора, если 1 - актив.|
|»» tax_contract_begin|string|true|none||Дата начала действия договора с налоговыми органами|
|»» tax_contract_end|string|true|none||Дата окончания действия договора с налоговыми органами|

## GET Реестр проведенных платежей

GET /payment/store/{store_id}/history

### Params

|Name|Location|Type|Required|Description|
|---|---|---|---|---|
|store_id|path|string| yes |none|
|offset|query|integer| yes |Отступ записей в ответе|
|limit|query|integer| yes |Кол-во записей в ответе|
|only_status|query|string| no |фильтрация транзакций с конкретным статусом|
|start_date|query|string| yes |Время в формате YYYY-mm-dd H:i:s. Временная зона GMT+5|
|end_date|query|string| yes |Время в формате YYYY-mm-dd H:i:s. Временная зона GMT+5|

#### Enum

|Name|Value|
|---|---|
|only_status|success|
|only_status|error|
|only_status|revert|
|only_status|progress|
|only_status|billing|
|only_status|draft|

> Response Examples

> 200 Response

```json
{
  "success": true,
  "data": {
    "pagination": {
      "total": 0,
      "offset": 0,
      "limit": 0
    },
    "stat": [
      {
        "status": "draft",
        "ps": "uzcard",
        "payment_amount": 0
      }
    ],
    "list": [
      {
        "id": 0,
        "uuid": "095be615-a8ad-4c33-8e9c-c7612fbf6c9f",
        "status": "draft",
        "ps": "uzcard",
        "store_invoice_id": "string",
        "billing_id": "string",
        "payment_time": "string",
        "payment_amount": 0,
        "commission_amount": 0,
        "card_token": "string",
        "card_pan": "string",
        "ps_uniq_id": "string",
        "refund_time": "2019-08-24T14:15:22Z",
        "added_on": "2019-08-24T14:15:22Z"
      }
    ]
  }
}
```

### Responses

|HTTP Status Code |Meaning|Description|Data schema|
|---|---|---|---|
|200|[OK](https://tools.ietf.org/html/rfc7231#section-6.3.1)|none|Inline|
|400|[Bad Request](https://tools.ietf.org/html/rfc7231#section-6.5.1)|none|None|
|403|[Forbidden](https://tools.ietf.org/html/rfc7231#section-6.5.3)|none|None|

### Responses Data Schema

HTTP Status Code **200**

|Name|Type|Required|Restrictions|Title|description|
|---|---|---|---|---|---|
|» success|boolean|true|none||none|
|» data|object|true|none||none|
|»» pagination|object|true|none||none|
|»»» total|integer|true|none||none|
|»»» offset|integer|true|none||none|
|»»» limit|integer|true|none||none|
|»» stat|[object]|true|none||Агрегированная информация по транзакциям|
|»»» status|[PaymentStatusEnum](#schemapaymentstatusenum)|true|none||Статус|
|»»» ps|[PaymentServiceEnum](#schemapaymentserviceenum)|true|none||Платежный метод (система или сервис)|
|»»» payment_amount|integer¦null|true|none||сумма платежей по заданному платежному методу в указанном статусе в тийинах|
|»» list|[object]|true|none||Список транзакций|
|»»» id|integer|true|none||none|
|»»» uuid|string(uuid)|true|none||none|
|»»» status|[PaymentStatusEnum](#schemapaymentstatusenum)|true|none||none|
|»»» ps|[PaymentServiceEnum](#schemapaymentserviceenum)|true|none||Платежная система или сервис|
|»»» store_invoice_id|string|true|none||none|
|»»» billing_id|string|false|none||none|
|»»» payment_time|string|false|none||none|
|»»» payment_amount|integer|true|none||none|
|»»» commission_amount|integer|true|none||none|
|»»» card_token|string|true|none||none|
|»»» card_pan|string|false|none||none|
|»»» ps_uniq_id|string|false|none||none|
|»»» refund_time|string(date-time)|false|none||none|
|»»» added_on|string(date-time)|true|none||none|

#### Enum

|Name|Value|
|---|---|
|status|draft|
|status|progress|
|status|billing|
|status|success|
|status|error|
|status|revert|
|ps|uzcard|
|ps|humo|
|ps|visa|
|ps|mastercard|
|ps|account|
|ps|payme|
|ps|click|
|ps|uzum|
|ps|anorbank|
|ps|oson|
|ps|alif|
|ps|xazna|
|ps|beepul|
|ps|trastpay|
|ps|sbp|
|status|draft|
|status|progress|
|status|billing|
|status|success|
|status|error|
|status|revert|
|ps|uzcard|
|ps|humo|
|ps|visa|
|ps|mastercard|
|ps|account|
|ps|payme|
|ps|click|
|ps|uzum|
|ps|anorbank|
|ps|oson|
|ps|alif|
|ps|xazna|
|ps|beepul|
|ps|trastpay|
|ps|sbp|

## GET История проведенных выплат на карты (пополнений)

GET /payment/store/{store_id}/credit-history

### Params

|Name|Location|Type|Required|Description|
|---|---|---|---|---|
|store_id|path|string| yes |none|
|offset|query|integer| yes |Отступ записей в ответе|
|limit|query|integer| yes |Кол-во записей в ответе|
|only_status|query|string| yes |фильтрация транзакций с конкретным статусом|
|start_date|query|string| yes |Время в формате YYYY-mm-dd H:i:s. Временная зона GMT+5|
|end_date|query|string| yes |Время в формате YYYY-mm-dd H:i:s. Временная зона GMT+5|

> Response Examples

> 200 Response

```json
{
    "success": true,
    "data": {
        "pagination": {
            "total": 9,
            "offset": 0,
            "limit": 10
        },
        "stat": [
            {
                "status": "draft",
                "ps": "uzcard",
                "payment_amount": "300000"
            },
            {
                "status": "success",
                "ps": "uzcard",
                "payment_amount": "50000"
            },
            {
                "status": "error",
                "ps": "uzcard",
                "payment_amount": "100000"
            }
        ],
        "list": [
            {
                "id": 25788622,
                "uuid": "fc207b3d-79b8-11ef-bfaa-00505680eaf6",
                "status": "draft",
                "ps": "uzcard",
                "store_invoice_id": "909903929",
                "billing_id": null,
                "payment_time": null,
                "payment_amount": 50000,
                "commission_amount": 1250,
                "total_amount": 50000,
                "card_token": "65b81d37b53977001c14391a",
                "card_pan": "860006******2278",
                "added_on": "2024-09-23 19:34:48"
            }
        ]
    }
}
```

### Responses

|HTTP Status Code |Meaning|Description|Data schema|
|---|---|---|---|
|200|[OK](https://tools.ietf.org/html/rfc7231#section-6.3.1)|none|Inline|

### Responses Data Schema

HTTP Status Code **200**

|Name|Type|Required|Restrictions|Title|description|
|---|---|---|---|---|---|
|» success|boolean|true|none||none|
|» data|object|true|none||none|
|»» pagination|object|true|none||none|
|»»» total|integer|true|none||none|
|»»» offset|integer|true|none||none|
|»»» limit|integer|true|none||none|
|»» stat|[object]|true|none||Агрегированная информация по транзакциям|
|»»» status|[PaymentStatusEnum](#schemapaymentstatusenum)|true|none||Статус|
|»»» ps|[PaymentServiceEnum](#schemapaymentserviceenum)|true|none||Платежный метод (система или сервис)|
|»»» payment_amount|integer¦null|true|none||сумма платежей по заданному платежному методу в указанном статусе в тийинах|
|»» list|[object]|true|none||Список транзакций|
|»»» id|integer|true|none||none|
|»»» uuid|string(uuid)|true|none||none|
|»»» status|[PaymentStatusEnum](#schemapaymentstatusenum)|true|none||none|
|»»» ps|[PaymentServiceEnum](#schemapaymentserviceenum)|true|none||Платежная система или сервис|
|»»» store_invoice_id|string|true|none||none|
|»»» billing_id|string|false|none||none|
|»»» payment_time|string|false|none||none|
|»»» payment_amount|integer|true|none||none|
|»»» commission_amount|integer|true|none||none|
|»»» card_token|string|true|none||none|
|»»» card_pan|string|false|none||none|
|»»» ps_uniq_id|string|false|none||none|
|»»» added_on|string(date-time)|true|none||none|

#### Enum

|Name|Value|
|---|---|
|status|draft|
|status|progress|
|status|billing|
|status|success|
|status|error|
|status|revert|
|ps|uzcard|
|ps|humo|
|ps|visa|
|ps|mastercard|
|ps|account|
|ps|payme|
|ps|click|
|ps|uzum|
|ps|anorbank|
|ps|oson|
|ps|alif|
|ps|xazna|
|ps|beepul|
|ps|trastpay|
|ps|sbp|
|status|draft|
|status|progress|
|status|billing|
|status|success|
|status|error|
|status|revert|
|ps|uzcard|
|ps|humo|
|ps|visa|
|ps|mastercard|
|ps|account|
|ps|payme|
|ps|click|
|ps|uzum|
|ps|anorbank|
|ps|oson|
|ps|alif|
|ps|xazna|
|ps|beepul|
|ps|trastpay|
|ps|sbp|

# Data Schema

<h2 id="tocS_applicationModel">applicationModel</h2>

<a id="schemaapplicationmodel"></a>
<a id="schema_applicationModel"></a>
<a id="tocSapplicationmodel"></a>
<a id="tocsapplicationmodel"></a>

```json
{
  "id": 0,
  "application_id": "string",
  "wallet_sum": 0,
  "wallet_sender_account": "string",
  "wallet_overdraft": 0,
  "wallet_contract_num": "string",
  "otp_required": 0,
  "otp_gateway": "string",
  "sms_nickname": "string"
}

```

### Attribute

|Name|Type|Required|Restrictions|Title|Description|
|---|---|---|---|---|---|
|id|integer|true|none||none|
|application_id|string|true|none||none|
|wallet_sum|integer¦null|true|none||none|
|wallet_sender_account|string¦null|true|none||none|
|wallet_overdraft|integer|true|none||none|
|wallet_contract_num|string¦null|true|none||none|
|otp_required|integer|true|none||none|
|otp_gateway|string¦null|true|none||none|
|sms_nickname|string¦null|true|none||none|

<h2 id="tocS_paymentModel">paymentModel</h2>

<a id="schemapaymentmodel"></a>
<a id="schema_paymentModel"></a>
<a id="tocSpaymentmodel"></a>
<a id="tocspaymentmodel"></a>

```json
{
  "id": 0,
  "uuid": "095be615-a8ad-4c33-8e9c-c7612fbf6c9f",
  "store_id": 0,
  "payment_amount": 0,
  "commission_type": "up",
  "commission_amount": 0,
  "total_amount": 0,
  "store_invoice_id": "string",
  "status": "draft",
  "callback_url": "string",
  "billing_id": "string",
  "phone": "string",
  "ps": "uzcard",
  "receipt_url": "string",
  "kyc_data": {
    "last_name": "string",
    "first_name": "string",
    "middle_name": "string",
    "passport": "string",
    "dob": "string",
    "passport_expiry_date": "string"
  },
  "device_details": {
    "ip": "string",
    "user_agent": "string"
  },
  "details": {},
  "card_token": "string",
  "card_pan": "string",
  "split": [
    {
      "type": "account",
      "amount": 0,
      "details": "string",
      "recipient": "string"
    }
  ],
  "multicard_user_id": 0,
  "ofd": [
    {
      "qty": 0,
      "vat": 0,
      "price": 0,
      "mxik": "string",
      "total": 0,
      "package_code": "string",
      "name": "string",
      "tin": "string",
      "mark": [
        "string"
      ]
    }
  ],
  "terminal_id": "string",
  "merchant_id": "string",
  "ps_uniq_id": "string",
  "ps_response_code": "string",
  "ps_response_msg": "string",
  "callback_message": "string",
  "payment_time": "2019-08-24T14:15:22Z",
  "refund_time": "2019-08-24T14:15:22Z",
  "otp_hash": "string",
  "clearing_id": 0,
  "tax_receipt_id": 0,
  "push_sent_at": "2019-08-24T14:15:22Z",
  "store": {
    "id": 0,
    "uuid": "095be615-a8ad-4c33-8e9c-c7612fbf6c9f",
    "category_id": 0,
    "note": "string",
    "logo": "string",
    "color": "string",
    "view_fields": [
      {
        "type": "string",
        "name": "string",
        "value": "string",
        "key": "string",
        "suggested": true
      }
    ],
    "tax_registration": 0,
    "tax_mxik": "string",
    "tax_package_code": "string",
    "tax_commission_recipient_tin": "string",
    "tg_chat_id": "string",
    "qr_url": "string",
    "bg_img": "string",
    "title": "string",
    "merchant": {
      "id": 0,
      "name": "string",
      "tin": "string",
      "contract_id": "string",
      "bank_account": "string"
    },
    "contract": {
      "id": 0,
      "num": "string",
      "date": "2019-08-24",
      "service": "string",
      "fee": {
        "up": "string",
        "down": "string"
      },
      "edm_document_id": "string",
      "edm_status": "string"
    },
    "merchant_account": {
      "id": 0,
      "uuid": "095be615-a8ad-4c33-8e9c-c7612fbf6c9f",
      "tin": "string",
      "official_name": "string",
      "mfo": "string",
      "account_no": "string",
      "address": "string",
      "director": "string",
      "director_pinfl": "string",
      "vat_payer": true,
      "is_commitent": true,
      "active": true
    }
  },
  "application": {
    "id": 0,
    "application_id": "string",
    "wallet_sum": 0,
    "wallet_sender_account": "string",
    "wallet_overdraft": 0,
    "wallet_contract_num": "string",
    "otp_required": 0,
    "otp_gateway": "string",
    "sms_nickname": "string"
  },
  "tax": {
    "receipt": {},
    "f_num": "1d2db1d7-fdce-4ae6-9c43-6a437dcdbc89",
    "fm_num": "string",
    "qr_url": "string",
    "is_refund": true
  },
  "refund_tax": {
    "receipt": {},
    "f_num": "1d2db1d7-fdce-4ae6-9c43-6a437dcdbc89",
    "fm_num": "string",
    "qr_url": "string",
    "is_refund": true
  },
  "clearing": {
    "id": 0,
    "merchant": {
      "id": 0,
      "name": "string",
      "tin": "string",
      "contract_id": "string",
      "bank_account": "string"
    },
    "recipient_info": {
      "id": 0,
      "uuid": "095be615-a8ad-4c33-8e9c-c7612fbf6c9f",
      "tin": "string",
      "official_name": "string",
      "mfo": "string",
      "account_no": "string",
      "address": "string",
      "director": "string",
      "director_pinfl": "string",
      "vat_payer": true,
      "is_commitent": true,
      "active": true
    },
    "sender_info": {
      "id": 0,
      "uuid": "095be615-a8ad-4c33-8e9c-c7612fbf6c9f",
      "tin": "string",
      "official_name": "string",
      "mfo": "string",
      "account_no": "string",
      "address": "string",
      "director": "string",
      "director_pinfl": "string",
      "vat_payer": true,
      "is_commitent": true,
      "active": true
    },
    "purpose_code": "string",
    "amount": 0,
    "details": "string",
    "status": "new",
    "payment_time": "string",
    "added_on": "string",
    "updated_on": "string",
    "receipt_url": "string"
  },
  "checkout_url": "string",
  "added_on": "2019-08-24T14:15:22Z"
}

```

### Attribute

|Name|Type|Required|Restrictions|Title|Description|
|---|---|---|---|---|---|
|id|integer|true|none||ID транзакции в Multicard|
|uuid|string(uuid)|true|none||UUID транзакции в Multicard|
|store_id|integer|true|none||ID кассы, выданный со стороны Multicard|
|payment_amount|integer|true|none||none|
|commission_type|string|true|none||Тип комиссии|
|commission_amount|integer|true|none||none|
|total_amount|integer|true|none||Сумма платежа в тийинах|
|store_invoice_id|string|true|none||Любой идентификатор заказа в системе Партнера. Будет возвращен в callback-запросе. Также по нему можно искать платежаи в кабинете мерчанта|
|status|[PaymentStatusEnum](#schemapaymentstatusenum)|true|none||Статус транзакции|
|callback_url|string¦null|true|none||URL для отправки callback-запроса|
|billing_id|string¦null|true|none||Уникальный ID транзакции в системе Партнера|
|phone|string¦null|true|none||Телефон плательщика в формате 998XXXXXXXXX (при наличии)|
|ps|[PaymentServiceEnum](#schemapaymentserviceenum)|true|none||Платежный сервис/система|
|receipt_url|string¦null|true|none||Ссылка на платежный чек|
|kyc_data|object¦null|true|none||none|
|» last_name|string|true|none||Фамилия плательщика|
|» first_name|string|true|none||Имя плательщика|
|» middle_name|string|true|none||Отчество плательщика|
|» passport|string|true|none||Паспорт плательщика|
|» dob|string|true|none||Дата рождения плательщика (YYYY-MM-DD)|
|» passport_expiry_date|string|true|none||Дата истечения паспорта плательщика|
|device_details|object¦null|true|none||Объект с информацией об устройстве клиента|
|» ip|string|true|none||IP-адрес устройства клиента, с которого выполняется запрос|
|» user_agent|string|true|none||Cтрока User-Agent, содержащая информацию о браузере, операционной системе и типе устройства клиента|
|details|object¦null|true|none||Поля, необходимые для проведения платежа в биллинге. Используется только при оплате за услуги Paynet и МУНИС (мобильная связь, гос.платежи и т.п.). Список полей и их название зависит от конкретной услуги (store_id)|
|card_token|string¦null|true|none||none|
|card_pan|string¦null|true|none||none|
|split|[splitRequest](#schemasplitrequest)|true|none||none|
|multicard_user_id|integer¦null|true|none||ID пользователя в приложении Multicard (если оплата проведена через приложение)|
|ofd|[ofdRequest](#schemaofdrequest)|true|none||none|
|terminal_id|string¦null|true|none||ID терминала в платежной системе|
|merchant_id|string¦null|true|none||ID мерчанта в платежной системе|
|ps_uniq_id|string¦null|true|none||Reference number (RRN, RefNum) в платежной системе|
|ps_response_code|string¦null|true|none||Код ответа (ошибки) от платежной системы|
|ps_response_msg|string¦null|true|none||Описание ошибки от платежной системы|
|callback_message|string¦null|true|none||Текст ответа от биллинга мерчанта/поставщика|
|payment_time|string(date-time)¦null|true|none||Дата и время оплаты во временной зоне GMT+5 в формате YYYY-mm-dd H:s:s|
|refund_time|string(date-time)|true|none||Дата и время возврата во временной зоне GMT+5 в формате YYYY-mm-dd H:s:s|
|otp_hash|string¦null|true|none||Если значение не null, то требуется подтверждение транзакции через SMS-код (или 3DS)|
|clearing_id|integer¦null|true|none||none|
|tax_receipt_id|integer¦null|true|none||none|
|push_sent_at|string(date-time)¦null|true|none||none|
|store|[storeModel](#schemastoremodel)|true|none||none|
|application|[applicationModel](#schemaapplicationmodel)|true|none||none|
|tax|[taxReceiptModel](#schemataxreceiptmodel)|true|none||none|
|refund_tax|[taxReceiptModel](#schemataxreceiptmodel)|true|none||none|
|clearing|[clearingModel](#schemaclearingmodel)|true|none||none|
|checkout_url|string¦null|true|none||URL страницы для оплаты|
|added_on|string(date-time)|true|none||none|

#### Enum

|Name|Value|
|---|---|
|commission_type|up|
|commission_type|down|

<h2 id="tocS_cardModel">cardModel</h2>

<a id="schemacardmodel"></a>
<a id="schema_cardModel"></a>
<a id="tocScardmodel"></a>
<a id="tocscardmodel"></a>

```json
{
  "card_token": "string",
  "card_pan": "string",
  "user_phone": "string",
  "pinfl": "string",
  "expiry": "string",
  "holder_name": "string",
  "redirect_url": "string",
  "redirect_decline_url": "string",
  "session_id": "string",
  "form_url": "string"
}

```

### Attribute

|Name|Type|Required|Restrictions|Title|Description|
|---|---|---|---|---|---|
|card_token|string|true|none||Max length:255<br />Токен карты для проведения платежных операций с картой|
|card_pan|string|true|none||Max length:16<br />Маскированный номер карты|
|user_phone|string|false|none||Max length:12<br />Телефон держателя карты в формате 998XXXXXXXXX (поддерживается только для карт Uzcard и Humo)|
|pinfl|string|false|none||ПИНФЛ держателя карты (возвращается в случае передачи в запросе на создание привязки для карт Uzcard и Humo)<br />Max length:14|
|expiry|string|true|none||Срок действия карты (YYMM)|
|holder_name|string|false|none||Имя и фамилия владельца карты|
|redirect_url|string|false|none||URL для перенаправления после успешного добавления карты пользователем|
|redirect_decline_url|string|false|none||URL для перенаправления при неуспешном добавлении карты или отмены пользователем|
|session_id|string|false|none||Необходимо сохранить данное значение, оно возвращается в callback-запросе в поле payer_id. Также можно проверить состояние через метод проверки привязанной карты|
|form_url|string|false|none||URL формы добавления карты (необходимо перенаправить пользователя, либо открыть webView)|

<h2 id="tocS_splitRequest">splitRequest</h2>

<a id="schemasplitrequest"></a>
<a id="schema_splitRequest"></a>
<a id="tocSsplitrequest"></a>
<a id="tocssplitrequest"></a>

```json
[
  {
    "type": "account",
    "amount": 0,
    "details": "string",
    "recipient": "string"
  }
]

```

### Attribute

|Name|Type|Required|Restrictions|Title|Description|
|---|---|---|---|---|---|
|type|string|true|none||Тип получателя|
|amount|integer|true|none||Сумма платежа|
|details|string|true|none||Детали платежа|
|recipient|string|false|none||Получатель. Если type=account, то передается uuid банковских реквизитов. По-умолчанию подставляются банковские реквизиты мерчанта.|

#### Enum

|Name|Value|
|---|---|
|type|account|
|type|wallet|
|type|card|

<h2 id="tocS_ofdRequest">ofdRequest</h2>

<a id="schemaofdrequest"></a>
<a id="schema_ofdRequest"></a>
<a id="tocSofdrequest"></a>
<a id="tocsofdrequest"></a>

```json
[
  {
    "qty": 0,
    "vat": 0,
    "price": 0,
    "mxik": "string",
    "total": 0,
    "package_code": "string",
    "name": "string",
    "tin": "string",
    "mark": [
      "string"
    ]
  }
]

```

### Attribute

|Name|Type|Required|Restrictions|Title|Description|
|---|---|---|---|---|---|
|qty|integer|true|none||Количество единиц товара/услуги|
|vat|integer|false|none||НДС (%)|
|price|integer|true|none||Стоимость единицы товара в тийинах|
|mxik|string|true|none||ИКПУ из справочника tasnif.soliq.uz|
|total|integer|true|none||Общая сумма товаров с учетом количества без учета скидок в тийинах|
|package_code|string|true|none||Код упаковки из справочника tasnif.soliq.uz|
|name|string|true|none||Наименование товара/услуги|
|tin|string|false|none||ИНН компании|
|mark|[string]|false|none||Массив с кодами маркировок каждой единицы товара. Обязателен для маркировочных товаров|

<h2 id="tocS_PaymentStatusEnum">PaymentStatusEnum</h2>

<a id="schemapaymentstatusenum"></a>
<a id="schema_PaymentStatusEnum"></a>
<a id="tocSpaymentstatusenum"></a>
<a id="tocspaymentstatusenum"></a>

```json
"draft"

```

### Attribute

|Name|Type|Required|Restrictions|Title|Description|
|---|---|---|---|---|---|
|*anonymous*|string|false|none||none|

#### Enum

|Name|Value|
|---|---|
|*anonymous*|draft|
|*anonymous*|progress|
|*anonymous*|billing|
|*anonymous*|success|
|*anonymous*|error|
|*anonymous*|revert|

<h2 id="tocS_PaymentServiceEnum">PaymentServiceEnum</h2>

<a id="schemapaymentserviceenum"></a>
<a id="schema_PaymentServiceEnum"></a>
<a id="tocSpaymentserviceenum"></a>
<a id="tocspaymentserviceenum"></a>

```json
"uzcard"

```

Платежная система или сервис

### Attribute

|Name|Type|Required|Restrictions|Title|Description|
|---|---|---|---|---|---|
|*anonymous*|string|false|none||Платежная система или сервис|

#### Enum

|Name|Value|
|---|---|
|*anonymous*|uzcard|
|*anonymous*|humo|
|*anonymous*|visa|
|*anonymous*|mastercard|
|*anonymous*|account|
|*anonymous*|payme|
|*anonymous*|click|
|*anonymous*|uzum|
|*anonymous*|anorbank|
|*anonymous*|oson|
|*anonymous*|alif|
|*anonymous*|xazna|
|*anonymous*|beepul|
|*anonymous*|trastpay|
|*anonymous*|sbp|

<h2 id="tocS_storeModel">storeModel</h2>

<a id="schemastoremodel"></a>
<a id="schema_storeModel"></a>
<a id="tocSstoremodel"></a>
<a id="tocsstoremodel"></a>

```json
{
  "id": 0,
  "uuid": "095be615-a8ad-4c33-8e9c-c7612fbf6c9f",
  "category_id": 0,
  "note": "string",
  "logo": "string",
  "color": "string",
  "view_fields": [
    {
      "type": "string",
      "name": "string",
      "value": "string",
      "key": "string",
      "suggested": true
    }
  ],
  "tax_registration": 0,
  "tax_mxik": "string",
  "tax_package_code": "string",
  "tax_commission_recipient_tin": "string",
  "tg_chat_id": "string",
  "qr_url": "string",
  "bg_img": "string",
  "title": "string",
  "merchant": {
    "id": 0,
    "name": "string",
    "tin": "string",
    "contract_id": "string",
    "bank_account": "string"
  },
  "contract": {
    "id": 0,
    "num": "string",
    "date": "2019-08-24",
    "service": "string",
    "fee": {
      "up": "string",
      "down": "string"
    },
    "edm_document_id": "string",
    "edm_status": "string"
  },
  "merchant_account": {
    "id": 0,
    "uuid": "095be615-a8ad-4c33-8e9c-c7612fbf6c9f",
    "tin": "string",
    "official_name": "string",
    "mfo": "string",
    "account_no": "string",
    "address": "string",
    "director": "string",
    "director_pinfl": "string",
    "vat_payer": true,
    "is_commitent": true,
    "active": true
  }
}

```

### Attribute

|Name|Type|Required|Restrictions|Title|Description|
|---|---|---|---|---|---|
|id|integer|true|none||none|
|uuid|string(uuid)|true|none||none|
|category_id|integer¦null|true|none||none|
|note|string¦null|true|none||none|
|logo|string¦null|true|none||Логотип|
|color|string¦null|true|none||none|
|view_fields|[billingFieldsModel](#schemabillingfieldsmodel)|true|none||none|
|tax_registration|integer|true|none||Флаг фискализации|
|tax_mxik|string¦null|true|none||ИКПУ по умолчанию из каталога tasnif.soliq.uz|
|tax_package_code|string¦null|true|none||Код упаковки от ИКПУ по умолчанию из каталога tasnif.soliq.uz|
|tax_commission_recipient_tin|string¦null|true|none||ИНН комитента для фискализации. Если null, то берется ИНН мерчанта|
|tg_chat_id|string¦null|true|none||ID телеграм группы для отправки уведомлений о платежах|
|qr_url|string|true|none||Ссылка на QR-код для приема платежей по данной кассе|
|bg_img|string|true|none||Фоновая картинка для страницы чекаута|
|title|string|true|none||Наименование кассы|
|merchant|[merchantModel](#schemamerchantmodel)|true|none||Информация о клиенте (мерчанте)|
|contract|object¦null|true|none||Информация о контракте с мерчантом|
|» id|integer|true|none||none|
|» num|string|true|none||Номер договора|
|» date|string(date)|true|none||Дата договора|
|» service|string|true|none||none|
|» fee|object|true|none||Комиссия по договору|
|»» up|string|true|none||none|
|»» down|string|true|none||none|
|» edm_document_id|string¦null|true|none||Идентификатор документа в системе электронного документооборота|
|» edm_status|string¦null|true|none||Статус подписания документа в системе электронного документооборота|
|merchant_account|[merchantAccount](#schemamerchantaccount)|true|none||none|

#### Enum

|Name|Value|
|---|---|
|tax_registration|0|
|tax_registration|1|
|tax_registration|2|
|tax_registration|3|

<h2 id="tocS_merchantModel">merchantModel</h2>

<a id="schemamerchantmodel"></a>
<a id="schema_merchantModel"></a>
<a id="tocSmerchantmodel"></a>
<a id="tocsmerchantmodel"></a>

```json
{
  "id": 0,
  "name": "string",
  "tin": "string",
  "contract_id": "string",
  "bank_account": "string"
}

```

### Attribute

|Name|Type|Required|Restrictions|Title|Description|
|---|---|---|---|---|---|
|id|integer|true|none||ID клиента (мерчанта) в Multicard|
|name|string|true|none||Наименование клиента (мерчанта)|
|tin|string|true|none||ИНН мерчанта|
|contract_id|string¦null|true|none||Данные о договоре|
|bank_account|string¦null|true|none||Транзитный счет для расчетов|

<h2 id="tocS_billingFieldsModel">billingFieldsModel</h2>

<a id="schemabillingfieldsmodel"></a>
<a id="schema_billingFieldsModel"></a>
<a id="tocSbillingfieldsmodel"></a>
<a id="tocsbillingfieldsmodel"></a>

```json
[
  {
    "type": "string",
    "name": "string",
    "value": "string",
    "key": "string",
    "suggested": true
  }
]

```

### Attribute

|Name|Type|Required|Restrictions|Title|Description|
|---|---|---|---|---|---|
|type|string|true|none||Формат поля|
|name|string|true|none||Описание поля|
|value|any|true|none||Значение|

oneOf

|Name|Type|Required|Restrictions|Title|Description|
|---|---|---|---|---|---|
|» *anonymous*|string|false|none||none|

xor

|Name|Type|Required|Restrictions|Title|Description|
|---|---|---|---|---|---|
|» *anonymous*|integer|false|none||none|

xor

|Name|Type|Required|Restrictions|Title|Description|
|---|---|---|---|---|---|
|» *anonymous*|boolean|false|none||none|

xor

|Name|Type|Required|Restrictions|Title|Description|
|---|---|---|---|---|---|
|» *anonymous*|array|false|none||none|

xor

|Name|Type|Required|Restrictions|Title|Description|
|---|---|---|---|---|---|
|» *anonymous*|object|false|none||none|

xor

|Name|Type|Required|Restrictions|Title|Description|
|---|---|---|---|---|---|
|» *anonymous*|number|false|none||none|

continued

|Name|Type|Required|Restrictions|Title|Description|
|---|---|---|---|---|---|
|key|string|true|none||Ключ|
|suggested|boolean¦null|true|none||Рекоммендуемая сумма оплаты|

#### Enum

|Name|Value|
|---|---|
|type|string|
|type|int|
|type|phone|
|type|tree|
|type|hidden|
|type|complex|
|type|select|

<h2 id="tocS_merchantAccount">merchantAccount</h2>

<a id="schemamerchantaccount"></a>
<a id="schema_merchantAccount"></a>
<a id="tocSmerchantaccount"></a>
<a id="tocsmerchantaccount"></a>

```json
{
  "id": 0,
  "uuid": "095be615-a8ad-4c33-8e9c-c7612fbf6c9f",
  "tin": "string",
  "official_name": "string",
  "mfo": "string",
  "account_no": "string",
  "address": "string",
  "director": "string",
  "director_pinfl": "string",
  "vat_payer": true,
  "is_commitent": true,
  "active": true
}

```

### Attribute

|Name|Type|Required|Restrictions|Title|Description|
|---|---|---|---|---|---|
|id|integer|true|none||ID банковских реквизитов|
|uuid|string(uuid)|true|none||UUID банковских реквизитов|
|tin|string|true|none||ИНН или ПИНФЛ|
|official_name|string|true|none||Наименование|
|mfo|string|true|none||МФО|
|account_no|string|true|none||Номер счета|
|address|string¦null|true|none||Юридический адрес|
|director|string¦null|true|none||ФИО директора|
|director_pinfl|string¦null|true|none||ПИНФЛ директора|
|vat_payer|boolean¦null|true|none||Является ли плательщиков НДС|
|is_commitent|boolean¦null|true|none||Является ли комитентом Multicard|
|active|boolean|true|none||Флаг активности. Если отключен, значит возмещение не будет отправлено|

<h2 id="tocS_taxReceiptModel">taxReceiptModel</h2>

<a id="schemataxreceiptmodel"></a>
<a id="schema_taxReceiptModel"></a>
<a id="tocStaxreceiptmodel"></a>
<a id="tocstaxreceiptmodel"></a>

```json
{
  "receipt": {},
  "f_num": "1d2db1d7-fdce-4ae6-9c43-6a437dcdbc89",
  "fm_num": "string",
  "qr_url": "string",
  "is_refund": true
}

```

### Attribute

|Name|Type|Required|Restrictions|Title|Description|
|---|---|---|---|---|---|
|receipt|object¦null|true|none||Объект чека, отправленный в ГНК|
|f_num|string(uuid)|true|none||Фискальный признак|
|fm_num|string|true|none||Фискальный терминал|
|qr_url|string¦null|true|none||URL на фискальный чек|
|is_refund|boolean|true|none||Является ли чеком возврата|

<h2 id="tocS_clearingModel">clearingModel</h2>

<a id="schemaclearingmodel"></a>
<a id="schema_clearingModel"></a>
<a id="tocSclearingmodel"></a>
<a id="tocsclearingmodel"></a>

```json
{
  "id": 0,
  "merchant": {
    "id": 0,
    "name": "string",
    "tin": "string",
    "contract_id": "string",
    "bank_account": "string"
  },
  "recipient_info": {
    "id": 0,
    "uuid": "095be615-a8ad-4c33-8e9c-c7612fbf6c9f",
    "tin": "string",
    "official_name": "string",
    "mfo": "string",
    "account_no": "string",
    "address": "string",
    "director": "string",
    "director_pinfl": "string",
    "vat_payer": true,
    "is_commitent": true,
    "active": true
  },
  "sender_info": {
    "id": 0,
    "uuid": "095be615-a8ad-4c33-8e9c-c7612fbf6c9f",
    "tin": "string",
    "official_name": "string",
    "mfo": "string",
    "account_no": "string",
    "address": "string",
    "director": "string",
    "director_pinfl": "string",
    "vat_payer": true,
    "is_commitent": true,
    "active": true
  },
  "purpose_code": "string",
  "amount": 0,
  "details": "string",
  "status": "new",
  "payment_time": "string",
  "added_on": "string",
  "updated_on": "string",
  "receipt_url": "string"
}

```

### Attribute

|Name|Type|Required|Restrictions|Title|Description|
|---|---|---|---|---|---|
|id|integer|true|none||none|
|merchant|[merchantModel](#schemamerchantmodel)|true|none||Информация о клиенте (мерчанте)|
|recipient_info|[merchantAccount](#schemamerchantaccount)|true|none||none|
|sender_info|[merchantAccount](#schemamerchantaccount)|true|none||none|
|purpose_code|string|true|none||Код назначения платежа|
|amount|integer|true|none||Сумма платежа в тийинах|
|details|string|true|none||Детали платежа|
|status|string|true|none||Статус|
|payment_time|string|true|none||Время проведения платежа|
|added_on|string|true|none||Дата создания записи|
|updated_on|string|true|none||Дата изменения записи|
|receipt_url|string|true|none||URL на банковскую квитанцию|

#### Enum

|Name|Value|
|---|---|
|status|new|
|status|sent|
|status|done|
|status|repeat|
|status|postponed|
|status|blocked|
|status|revert|
|status|canceled|

