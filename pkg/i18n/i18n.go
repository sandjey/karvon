package i18n

import "github.com/gin-gonic/gin"

const LangKey = "lang"
const DefaultLang = "ru"

var messages = map[string]map[string]string{
	"ru": {
		// errors
		"UNAUTHORIZED":            "Не авторизован",
		"FORBIDDEN":               "Нет прав",
		"USER_BLOCKED":            "Пользователь заблокирован",
		"NOT_FOUND":               "Не найдено",
		"USER_NOT_FOUND":          "Пользователь не найден",
		"COMPANY_NOT_FOUND":       "Компания не найдена",
		"INTERNAL_ERROR":          "Внутренняя ошибка сервера",
		"VALIDATION_ERROR":        "Ошибка валидации данных",
		"OTP_INVALID":             "Неверный код подтверждения",
		"OTP_EXPIRED":             "Код истёк. Запросите новый",
		"OTP_MAX_ATTEMPTS":        "Превышен лимит запросов OTP. Попробуйте через час",
		"NAME_ALREADY_SET":        "Имя уже установлено",
		"ALREADY_EXISTS":          "Запись уже существует",
		"COMPANY_NOT_VERIFIED":    "Компания не верифицирована",
		"COMPANY_NOT_EDITABLE":    "Компанию нельзя редактировать в текущем статусе",
		"INSUFFICIENT_TOKENS":     "Недостаточно токенов",
		"FREE_LISTING_USED":       "Бесплатное объявление уже использовано",
		"PAYMENT_FAILED":          "Ошибка платежа",
		"SERVICE_UNAVAILABLE":     "Сервис временно недоступен",
		"TELEGRAM_NOT_CONFIGURED": "Telegram не настроен",
		"WHATSAPP_NOT_CONFIGURED": "WhatsApp сервис недоступен",
		"FILE_TOO_LARGE":          "Файл слишком большой. Максимум 10 МБ",
		"FILE_TYPE_NOT_ALLOWED":   "Тип файла не разрешён. Допустимы: изображения и PDF",
		// success
		"OTP_SENT":              "OTP код отправлен",
		"AUTH_SUCCESS":          "Успешная авторизация",
		"REGISTRATION_COMPLETE": "Регистрация завершена",
		"LOGOUT_SUCCESS":        "Выход выполнен",
		"PROFILE_UPDATED":       "Профиль обновлён",
		"COMPANY_CREATED":       "Компания отправлена на проверку",
		"COMPANY_UPDATED":       "Данные компании обновлены",
		"FILE_UPLOADED":         "Файл успешно загружен",
		// moderator
		"COMPANY_APPROVED": "Компания одобрена",
		"COMPANY_REJECTED": "Компания отклонена",
		"DOCS_REQUESTED":   "Документы запрошены у пользователя",
		// listings / contacts / etc.
		"LISTING_NOT_FOUND":     "Объявление не найдено",
		"INVALID_CREDENTIALS":   "Неверный логин или пароль",
		"CARGO_CREATED":         "Объявление о грузе создано",
		"CARGO_UPDATED":         "Объявление обновлено",
		"CARGO_DELETED":         "Объявление удалено",
		"CARGO_STATUS_UPDATED":  "Статус объявления обновлён",
		"TEMPLATE_SAVED":        "Шаблон сохранён",
		"WAREHOUSE_CREATED":     "Склад создан",
		"WAREHOUSE_UPDATED":     "Склад обновлён",
		"WAREHOUSE_DELETED":     "Склад удалён",
		"CONTACT_OPENED":        "Контакт открыт",
		"FAVORITE_ADDED":        "Добавлено в избранное",
		"FAVORITE_REMOVED":      "Удалено из избранного",
		"ROUTE_SAVED":           "Маршрут сохранён",
		"ROUTE_DELETED":         "Маршрут удалён",
		"ROUTE_UPDATED":         "Маршрут обновлён",
		"NOTIFICATION_READ":     "Уведомление прочитано",
		"NOTIFICATIONS_READ":    "Все уведомления прочитаны",
		"MODERATOR_CREATED":     "Модератор создан",
		"MODERATOR_DELETED":     "Модератор удалён",
		"USER_BLOCK_UPDATED":    "Статус блокировки обновлён",
		"TOKENS_TOPPED_UP":      "Токены начислены",
		"PRICING_UPDATED":       "Тариф обновлён",
		"LISTING_DELETED":       "Объявление удалено",
		"LISTING_BLOCK_UPDATED": "Статус объявления обновлён",
	},
	"uz": {
		// errors
		"UNAUTHORIZED":            "Avtorizatsiya qilinmagan",
		"FORBIDDEN":               "Ruxsat yo'q",
		"USER_BLOCKED":            "Foydalanuvchi bloklangan",
		"NOT_FOUND":               "Topilmadi",
		"USER_NOT_FOUND":          "Foydalanuvchi topilmadi",
		"COMPANY_NOT_FOUND":       "Kompaniya topilmadi",
		"INTERNAL_ERROR":          "Ichki server xatosi",
		"VALIDATION_ERROR":        "Ma'lumotlarni tekshirishda xato",
		"OTP_INVALID":             "Noto'g'ri tasdiqlash kodi",
		"OTP_EXPIRED":             "Kod muddati o'tdi. Yangi kod so'rang",
		"OTP_MAX_ATTEMPTS":        "OTP so'rovlar limiti oshib ketdi. Bir soatdan keyin urinib ko'ring",
		"NAME_ALREADY_SET":        "Ism allaqachon o'rnatilgan",
		"ALREADY_EXISTS":          "Yozuv allaqachon mavjud",
		"COMPANY_NOT_VERIFIED":    "Kompaniya tasdiqlanmagan",
		"COMPANY_NOT_EDITABLE":    "Kompaniyani joriy holatda tahrirlash mumkin emas",
		"INSUFFICIENT_TOKENS":     "Tokenlar yetarli emas",
		"FREE_LISTING_USED":       "Bepul e'lon allaqachon ishlatilgan",
		"PAYMENT_FAILED":          "To'lov xatosi",
		"SERVICE_UNAVAILABLE":     "Xizmat vaqtincha mavjud emas",
		"TELEGRAM_NOT_CONFIGURED": "Telegram sozlanmagan",
		"WHATSAPP_NOT_CONFIGURED": "WhatsApp xizmati mavjud emas",
		"FILE_TOO_LARGE":          "Fayl juda katta. Maksimal hajm 10 MB",
		"FILE_TYPE_NOT_ALLOWED":   "Fayl turi ruxsat etilmagan. Rasm va PDF fayllar ruxsat etiladi",
		// success
		"OTP_SENT":              "OTP kodi yuborildi",
		"AUTH_SUCCESS":          "Muvaffaqiyatli kirish",
		"REGISTRATION_COMPLETE": "Ro'yxatdan o'tish yakunlandi",
		"LOGOUT_SUCCESS":        "Chiqish amalga oshirildi",
		"PROFILE_UPDATED":       "Profil yangilandi",
		"COMPANY_CREATED":       "Kompaniya tekshiruvga yuborildi",
		"COMPANY_UPDATED":       "Kompaniya ma'lumotlari yangilandi",
		"FILE_UPLOADED":         "Fayl muvaffaqiyatli yuklandi",
		// moderator
		"COMPANY_APPROVED": "Kompaniya tasdiqlandi",
		"COMPANY_REJECTED": "Kompaniya rad etildi",
		"DOCS_REQUESTED":   "Foydalanuvchidan hujjatlar so'raldi",
		// listings / contacts / etc.
		"LISTING_NOT_FOUND":     "E'lon topilmadi",
		"INVALID_CREDENTIALS":   "Login yoki parol noto'g'ri",
		"CARGO_CREATED":         "Yuk e'loni yaratildi",
		"CARGO_UPDATED":         "E'lon yangilandi",
		"CARGO_DELETED":         "E'lon o'chirildi",
		"CARGO_STATUS_UPDATED":  "E'lon holati yangilandi",
		"TEMPLATE_SAVED":        "Shablon saqlandi",
		"WAREHOUSE_CREATED":     "Ombor yaratildi",
		"WAREHOUSE_UPDATED":     "Ombor yangilandi",
		"WAREHOUSE_DELETED":     "Ombor o'chirildi",
		"CONTACT_OPENED":        "Kontakt ochildi",
		"FAVORITE_ADDED":        "Sevimlilarga qo'shildi",
		"FAVORITE_REMOVED":      "Sevimlilardan o'chirildi",
		"ROUTE_SAVED":           "Marshrut saqlandi",
		"ROUTE_DELETED":         "Marshrut o'chirildi",
		"ROUTE_UPDATED":         "Marshrut yangilandi",
		"NOTIFICATION_READ":     "Bildirishnoma o'qildi",
		"NOTIFICATIONS_READ":    "Barcha bildirishnomalar o'qildi",
		"MODERATOR_CREATED":     "Moderator yaratildi",
		"MODERATOR_DELETED":     "Moderator o'chirildi",
		"USER_BLOCK_UPDATED":    "Bloklash holati yangilandi",
		"TOKENS_TOPPED_UP":      "Tokenlar qo'shildi",
		"PRICING_UPDATED":       "Tarif yangilandi",
		"LISTING_DELETED":       "E'lon o'chirildi",
		"LISTING_BLOCK_UPDATED": "E'lon holati yangilandi",
	},
	"en": {
		// errors
		"UNAUTHORIZED":            "Unauthorized",
		"FORBIDDEN":               "Forbidden",
		"USER_BLOCKED":            "User is blocked",
		"NOT_FOUND":               "Not found",
		"USER_NOT_FOUND":          "User not found",
		"COMPANY_NOT_FOUND":       "Company not found",
		"INTERNAL_ERROR":          "Internal server error",
		"VALIDATION_ERROR":        "Validation error",
		"OTP_INVALID":             "Invalid OTP code",
		"OTP_EXPIRED":             "OTP expired. Please request a new one",
		"OTP_MAX_ATTEMPTS":        "OTP request limit exceeded. Try again in an hour",
		"NAME_ALREADY_SET":        "Name is already set",
		"ALREADY_EXISTS":          "Record already exists",
		"COMPANY_NOT_VERIFIED":    "Company is not verified",
		"COMPANY_NOT_EDITABLE":    "Company cannot be edited in its current status",
		"INSUFFICIENT_TOKENS":     "Insufficient tokens",
		"FREE_LISTING_USED":       "Free listing already used",
		"PAYMENT_FAILED":          "Payment failed",
		"SERVICE_UNAVAILABLE":     "Service temporarily unavailable",
		"TELEGRAM_NOT_CONFIGURED": "Telegram is not configured",
		"WHATSAPP_NOT_CONFIGURED": "WhatsApp service unavailable",
		"FILE_TOO_LARGE":          "File is too large. Maximum 10 MB",
		"FILE_TYPE_NOT_ALLOWED":   "File type not allowed. Images and PDF are accepted",
		// success
		"OTP_SENT":              "OTP code sent",
		"AUTH_SUCCESS":          "Successfully authorized",
		"REGISTRATION_COMPLETE": "Registration complete",
		"LOGOUT_SUCCESS":        "Logged out successfully",
		"PROFILE_UPDATED":       "Profile updated",
		"COMPANY_CREATED":       "Company submitted for verification",
		"COMPANY_UPDATED":       "Company updated",
		"FILE_UPLOADED":         "File uploaded successfully",
		// moderator
		"COMPANY_APPROVED": "Company approved",
		"COMPANY_REJECTED": "Company rejected",
		"DOCS_REQUESTED":   "Documents requested from user",
		// listings / contacts / etc.
		"LISTING_NOT_FOUND":     "Listing not found",
		"INVALID_CREDENTIALS":   "Invalid login or password",
		"CARGO_CREATED":         "Cargo listing created",
		"CARGO_UPDATED":         "Listing updated",
		"CARGO_DELETED":         "Listing deleted",
		"CARGO_STATUS_UPDATED":  "Listing status updated",
		"TEMPLATE_SAVED":        "Template saved",
		"WAREHOUSE_CREATED":     "Warehouse created",
		"WAREHOUSE_UPDATED":     "Warehouse updated",
		"WAREHOUSE_DELETED":     "Warehouse deleted",
		"CONTACT_OPENED":        "Contact opened",
		"FAVORITE_ADDED":        "Added to favorites",
		"FAVORITE_REMOVED":      "Removed from favorites",
		"ROUTE_SAVED":           "Route saved",
		"ROUTE_DELETED":         "Route deleted",
		"ROUTE_UPDATED":         "Route updated",
		"NOTIFICATION_READ":     "Notification marked read",
		"NOTIFICATIONS_READ":    "All notifications marked read",
		"MODERATOR_CREATED":     "Moderator created",
		"MODERATOR_DELETED":     "Moderator deleted",
		"USER_BLOCK_UPDATED":    "Block status updated",
		"TOKENS_TOPPED_UP":      "Tokens credited",
		"PRICING_UPDATED":       "Pricing updated",
		"LISTING_DELETED":       "Listing deleted",
		"LISTING_BLOCK_UPDATED": "Listing block status updated",
	},
}

// Lang returns the language stored in gin context, falling back to default.
func Lang(c *gin.Context) string {
	if v, ok := c.Get(LangKey); ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return DefaultLang
}

// T translates a message key using the language from gin context.
func T(c *gin.Context, code string) string {
	return TLang(Lang(c), code)
}

// TLang translates a message key using an explicit language string.
func TLang(lang, code string) string {
	if m, ok := messages[lang]; ok {
		if msg, ok := m[code]; ok {
			return msg
		}
	}
	if m, ok := messages[DefaultLang]; ok {
		if msg, ok := m[code]; ok {
			return msg
		}
	}
	return code
}
