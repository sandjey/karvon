package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	AppPort   string
	PublicURL string // base URL returned in file upload responses

	DBURL    string
	RedisURL string

	JWTSecret     string
	JWTAccessTTL  time.Duration
	JWTRefreshTTL time.Duration

	TelegramGatewayBaseURL string
	TelegramGatewayToken   string
	TelegramGatewayBypass  bool

	WhatsAppOTPBaseURL string
	WhatsAppOTPToken   string
	WhatsAppOTPBypass  bool

	// РЈРЅРёРІРµСЂСЃР°Р»СЊРЅС‹Р№ OTP-РєРѕРґ (QA): РїСЂРѕС…РѕРґРёС‚ РґР»СЏ Р»СЋР±РѕРіРѕ С‚РµР»РµС„РѕРЅР°, РµСЃР»Рё Р·Р°РґР°РЅ
	UniversalOTP string

	// URL СЃРІРѕРµРіРѕ СЃРµСЂРІРµСЂР° РєР°СЂС‚ (tileserver) РґР»СЏ С„СЂРѕРЅС‚РµРЅРґР°
	MapTilesURL string

	MulticardBaseURL     string
	MulticardAppID       string
	MulticardSecret      string
	MulticardStoreID     int
	MulticardMXIK        string
	MulticardPackageCode string
	MulticardCallbackURL string
	MulticardReturnURL   string

	StorageType string
	StoragePath string

	// РЎРєСЂС‹С‚С‹Р№ СЃС‚Р°С‚РёРє-Р°РґРјРёРЅ РґР»СЏ РІС…РѕРґР° РІ Р°РґРјРёРЅРєСѓ
	AdminLogin    string
	AdminPassword string

	// SMTP РґР»СЏ email-СѓРІРµРґРѕРјР»РµРЅРёР№ (РѕРїС†РёРѕРЅР°Р»СЊРЅРѕ)
	SMTPHost     string
	SMTPPort     string
	SMTPUser     string
	SMTPPassword string
	SMTPFrom     string
}

func Load() (*Config, error) {
	_ = godotenv.Load()

	cfg := &Config{
		AppPort:   getEnv("APP_PORT", "8080"),
		PublicURL: getEnv("PUBLIC_URL", ""),
		DBURL:     mustEnv("DB_URL"),
		RedisURL:  getEnv("REDIS_URL", "redis://localhost:6379"),
		JWTSecret: mustEnv("JWT_SECRET"),

		TelegramGatewayBaseURL: getEnv("TELEGRAM_GATEWAY_BASE_URL", "https://gatewayapi.telegram.org"),
		TelegramGatewayToken:   os.Getenv("TELEGRAM_GATEWAY_TOKEN"),
		TelegramGatewayBypass:  os.Getenv("TELEGRAM_GATEWAY_BYPASS") == "true",

		WhatsAppOTPBaseURL: getEnv("WHATSAPP_OTP_BASE_URL", "http://127.0.0.1:3210"),
		WhatsAppOTPToken:   os.Getenv("WHATSAPP_OTP_TOKEN"),
		WhatsAppOTPBypass:  os.Getenv("WHATSAPP_OTP_BYPASS") == "true",

		UniversalOTP: os.Getenv("UNIVERSAL_OTP"),
		MapTilesURL:  os.Getenv("MAP_TILES_URL"),

		MulticardBaseURL:     getEnv("MULTICARD_BASE_URL", "https://dev-mesh.multicard.uz"),
		MulticardAppID:       getEnv("MULTICARD_APP_ID", os.Getenv("RAHMAT_MERCHANT_ID")),
		MulticardSecret:      getEnv("MULTICARD_SECRET", os.Getenv("RAHMAT_SECRET_KEY")),
		MulticardStoreID:     getEnvInt("MULTICARD_STORE_ID", 6),
		MulticardMXIK:        getEnv("MULTICARD_MXIK", "10305001001000000"),
		MulticardPackageCode: getEnv("MULTICARD_PACKAGE_CODE", "1514918"),
		MulticardCallbackURL: os.Getenv("MULTICARD_CALLBACK_URL"),
		MulticardReturnURL:   os.Getenv("MULTICARD_RETURN_URL"),
		StorageType:      getEnv("STORAGE_TYPE", "local"),
		StoragePath:      getEnv("STORAGE_PATH", "./uploads"),

		AdminLogin:    getEnv("ADMIN_LOGIN", "superadmin"),
		AdminPassword: getEnv("ADMIN_PASSWORD", "ctm_admin_2026"),

		SMTPHost:     os.Getenv("SMTP_HOST"),
		SMTPPort:     getEnv("SMTP_PORT", "587"),
		SMTPUser:     os.Getenv("SMTP_USER"),
		SMTPPassword: os.Getenv("SMTP_PASSWORD"),
		SMTPFrom:     getEnv("SMTP_FROM", os.Getenv("SMTP_USER")),
	}

	var err error

	cfg.JWTAccessTTL, err = time.ParseDuration(getEnv("JWT_ACCESS_TTL", "15m"))
	if err != nil {
		return nil, fmt.Errorf("invalid JWT_ACCESS_TTL: %w", err)
	}

	cfg.JWTRefreshTTL, err = time.ParseDuration(getEnv("JWT_REFRESH_TTL", "720h"))
	if err != nil {
		return nil, fmt.Errorf("invalid JWT_REFRESH_TTL: %w", err)
	}

	return cfg, nil
}

func mustEnv(key string) string {
	v := os.Getenv(key)
	if v == "" {
		panic(fmt.Sprintf("required env variable %s is not set", key))
	}
	return v
}

func getEnv(key, defaultVal string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultVal
}

func getEnvInt(key string, defaultVal int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return defaultVal
}
