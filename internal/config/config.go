package config

import (
	"fmt"
	"os"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	AppPort string

	DBURL string

	JWTSecret     string
	JWTAccessTTL  time.Duration
	JWTRefreshTTL time.Duration

	TelegramGatewayBaseURL string
	TelegramGatewayToken   string
	TelegramGatewayBypass  bool

	WhatsAppOTPBaseURL string
	WhatsAppOTPToken   string
	WhatsAppOTPBypass  bool

	RahmatMerchantID string
	RahmatSecretKey  string

	StorageType string
	StoragePath string
}

func Load() (*Config, error) {
	_ = godotenv.Load()

	cfg := &Config{
		AppPort:   getEnv("APP_PORT", "8080"),
		DBURL:     mustEnv("DB_URL"),
		JWTSecret: mustEnv("JWT_SECRET"),

		TelegramGatewayBaseURL: getEnv("TELEGRAM_GATEWAY_BASE_URL", "https://gatewayapi.telegram.org"),
		TelegramGatewayToken:   os.Getenv("TELEGRAM_GATEWAY_TOKEN"),
		TelegramGatewayBypass:  os.Getenv("TELEGRAM_GATEWAY_BYPASS") == "true",

		WhatsAppOTPBaseURL: getEnv("WHATSAPP_OTP_BASE_URL", "http://127.0.0.1:3210"),
		WhatsAppOTPToken:   os.Getenv("WHATSAPP_OTP_TOKEN"),
		WhatsAppOTPBypass:  os.Getenv("WHATSAPP_OTP_BYPASS") == "true",

		RahmatMerchantID: os.Getenv("RAHMAT_MERCHANT_ID"),
		RahmatSecretKey:  os.Getenv("RAHMAT_SECRET_KEY"),
		StorageType:      getEnv("STORAGE_TYPE", "local"),
		StoragePath:      getEnv("STORAGE_PATH", "./uploads"),
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
