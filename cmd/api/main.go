package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"

	"karvon/internal/config"
	"karvon/internal/handler"
	"karvon/internal/middleware"
	"karvon/internal/model"
	"karvon/internal/repository"
	"karvon/internal/service"
	jwtpkg "karvon/pkg/jwt"
	"karvon/pkg/notifier"
	"karvon/pkg/rahmat"
	"karvon/pkg/storage"
)

func main() {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.RFC3339})

	cfg, err := config.Load()
	if err != nil {
		log.Fatal().Err(err).Msg("failed to load config")
	}

	// ── База данных ──────────────────────────────────────────────────────────
	gormDB, err := connectGORM(cfg.DBURL)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to connect to database")
	}
	log.Info().Msg("database connected")

	if err := runMigrations(gormDB); err != nil {
		log.Fatal().Err(err).Msg("failed to run migrations")
	}
	log.Info().Msg("migrations applied")

	if err := seedPricingConfig(gormDB); err != nil {
		log.Warn().Err(err).Msg("pricing config seed warning")
	}

	sqlDB, _ := gormDB.DB()
	_ = sqlx.NewDb(sqlDB, "pgx") // для будущих репозиториев через sqlx

	// ── Зависимости ──────────────────────────────────────────────────────────
	jwtMgr := jwtpkg.NewManager(cfg.JWTSecret, cfg.JWTAccessTTL, cfg.JWTRefreshTTL)

	var tgNotifier notifier.Notifier
	if cfg.TelegramGatewayToken != "" || cfg.TelegramGatewayBypass {
		tgNotifier = notifier.NewTelegramGateway(
			cfg.TelegramGatewayBaseURL,
			cfg.TelegramGatewayToken,
			cfg.TelegramGatewayBypass,
		)
		log.Info().Bool("bypass", cfg.TelegramGatewayBypass).Msg("telegram gateway notifier initialized")
	}

	waNotifier := notifier.NewWhatsApp(cfg.WhatsAppOTPBaseURL, cfg.WhatsAppOTPToken, cfg.WhatsAppOTPBypass)
	log.Info().Str("url", cfg.WhatsAppOTPBaseURL).Bool("bypass", cfg.WhatsAppOTPBypass).Msg("whatsapp notifier initialized")

	userRepo      := repository.NewUserRepo(gormDB)
	otpRepo       := repository.NewOTPRepo(gormDB)
	tokenRepo     := repository.NewTokenRepo(gormDB)
	companyRepo   := repository.NewCompanyRepo(gormDB)
	notifRepo     := repository.NewNotificationRepo(gormDB)
	cargoRepo     := repository.NewCargoRepo(gormDB)
	warehouseRepo := repository.NewWarehouseRepo(gormDB)
	contactRepo   := repository.NewContactRepo(gormDB)
	favoriteRepo  := repository.NewFavoriteRepo(gormDB)
	routeRepo     := repository.NewRouteRepo(gormDB)
	pricingRepo   := repository.NewPricingRepo(gormDB)
	adminRepo     := repository.NewAdminRepo(gormDB)
	mediaRepo     := repository.NewMediaRepo(gormDB)
	paymentRepo   := repository.NewPaymentRepo(gormDB)

	authSvc      := service.NewAuthService(userRepo, otpRepo, tokenRepo, pricingRepo, jwtMgr, waNotifier, tgNotifier, cfg.UniversalOTP)
	userSvc      := service.NewUserService(userRepo, companyRepo)
	companySvc   := service.NewCompanyService(companyRepo)
	moderatorSvc := service.NewModeratorService(companyRepo, notifRepo)
	cargoSvc     := service.NewCargoService(cargoRepo, companyRepo, routeRepo, notifRepo, favoriteRepo, mediaRepo)
	warehouseSvc := service.NewWarehouseService(warehouseRepo, companyRepo, mediaRepo)
	pricingSvc   := service.NewPricingService(pricingRepo)
	rahmatClient := rahmat.NewClient(rahmat.Config{
		BaseURL:     cfg.MulticardBaseURL,
		AppID:       cfg.MulticardAppID,
		Secret:      cfg.MulticardSecret,
		StoreID:     cfg.MulticardStoreID,
		MXIK:        cfg.MulticardMXIK,
		PackageCode: cfg.MulticardPackageCode,
		CallbackURL: cfg.MulticardCallbackURL,
		ReturnURL:   cfg.MulticardReturnURL,
	})
	paymentSvc   := service.NewPaymentService(paymentRepo, pricingRepo, userRepo, cargoRepo, warehouseRepo, rahmatClient)
	contactSvc   := service.NewContactService(cargoRepo, warehouseRepo, contactRepo, userRepo, notifRepo)
	favoriteSvc  := service.NewFavoriteService(favoriteRepo, cargoRepo, warehouseRepo)
	routeSvc     := service.NewRouteService(routeRepo)
	notifSvc     := service.NewNotificationService(notifRepo)
	adminSvc     := service.NewAdminService(adminRepo, userRepo, tokenRepo, cargoRepo, warehouseRepo, pricingSvc, jwtMgr, cfg.AdminLogin, cfg.AdminPassword)

	// Сид скрытого статик-админа
	if err := adminSvc.SeedSuperAdmin(context.Background()); err != nil {
		log.Warn().Err(err).Msg("super admin seed warning")
	} else {
		log.Info().Str("login", cfg.AdminLogin).Msg("super admin ensured")
	}

	store := storage.NewLocal(cfg.StoragePath)

	authMiddleware          := middleware.Auth(jwtMgr, userRepo)
	verifiedMiddleware      := middleware.CompanyVerified(companyRepo)
	moderatorRoleMiddleware := middleware.Role("moderator", "super_admin")
	superAdminMiddleware    := middleware.Role("super_admin")

	// Фоновые задачи
	startBackgroundWorkers(cargoRepo, warehouseRepo, otpRepo)

	// ── HTTP роутер ──────────────────────────────────────────────────────────
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(middleware.Lang())
	r.Use(requestLogger())

	// статические файлы — отдаём загруженные файлы по URL
	r.Static("/uploads", cfg.StoragePath)

	r.GET("/health", handler.HealthCheck)

	v1 := r.Group("/api/v1")
	handler.NewAuthHandler(authSvc).RegisterRoutes(v1, authMiddleware)
	handler.NewUserHandler(userSvc).RegisterRoutes(v1, authMiddleware)
	handler.NewCompanyHandler(companySvc).RegisterRoutes(v1, authMiddleware)
	uploadBaseURL := cfg.PublicURL
	if uploadBaseURL == "" {
		uploadBaseURL = "http://localhost:" + cfg.AppPort
	}
	handler.NewUploadHandler(store, uploadBaseURL).RegisterRoutes(v1, authMiddleware)
	handler.NewModeratorHandler(moderatorSvc).RegisterRoutes(v1, authMiddleware, moderatorRoleMiddleware)
	handler.NewGeoHandler().RegisterRoutes(v1)
	handler.NewConfigHandler(cfg.MapTilesURL).RegisterRoutes(v1)
	handler.NewCargoHandler(cargoSvc).RegisterRoutes(v1, authMiddleware, verifiedMiddleware)
	handler.NewWarehouseHandler(warehouseSvc).RegisterRoutes(v1, authMiddleware, verifiedMiddleware)
	handler.NewContactHandler(contactSvc, pricingSvc).RegisterRoutes(v1, authMiddleware)
	handler.NewPaymentsHandler(pricingSvc, paymentSvc).RegisterRoutes(v1, authMiddleware)
	handler.NewFavoriteHandler(favoriteSvc).RegisterRoutes(v1, authMiddleware)
	handler.NewRouteHandler(routeSvc).RegisterRoutes(v1, authMiddleware)
	handler.NewNotificationHandler(notifSvc).RegisterRoutes(v1, authMiddleware)
	handler.NewSearchHandler(cargoSvc, warehouseSvc).RegisterRoutes(v1)
	handler.NewAdminHandler(adminSvc).RegisterRoutes(v1, authMiddleware, superAdminMiddleware)

	// ── Сервер ───────────────────────────────────────────────────────────────
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%s", cfg.AppPort),
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
	}

	go func() {
		log.Info().Str("addr", srv.Addr).Msg("server started")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal().Err(err).Msg("server error")
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	log.Info().Msg("shutting down...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_ = srv.Shutdown(ctx)
}

func connectGORM(dsn string) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.New(postgres.Config{
		DSN:        dsn,
		DriverName: "postgres", // lib/pq — MD5 auth, без SCRAM
	}), &gorm.Config{
		Logger: gormlogger.Default.LogMode(gormlogger.Silent),
	})
	if err != nil {
		return nil, err
	}
	sqlDB, _ := db.DB()
	sqlDB.SetMaxOpenConns(25)
	sqlDB.SetMaxIdleConns(5)
	sqlDB.SetConnMaxLifetime(5 * time.Minute)
	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("db ping: %w", err)
	}
	return db, nil
}

func runMigrations(db *gorm.DB) error {
	if err := createEnumTypes(db); err != nil {
		return fmt.Errorf("create enum types: %w", err)
	}
	return db.AutoMigrate(
		&model.User{},
		&model.OTPCode{},
		&model.RefreshToken{},
		&model.Company{},
		&model.CargoListing{},
		&model.FleetVehicle{},
		&model.TransportListing{},
		&model.WarehouseListing{},
		&model.ListingMedia{},
		&model.ContactView{},
		&model.TokenTransaction{},
		&model.PaymentOrder{},
		&model.Subscription{},
		&model.PricingConfig{},
		&model.SavedRoute{},
		&model.Favorite{},
		&model.Notification{},
	)
}

func createEnumTypes(db *gorm.DB) error {
	enums := []struct{ name, vals string }{
		{"user_role", "'user','moderator','super_admin'"},
		{"company_status", "'pending','approved','rejected','docs_requested'"},
		{"company_org_type", "'ooo','ao','ip','ltd','gmbh','co_ltd'"},
		{"cargo_category_enum", "'stroymat','food','textile','metal','chemical','wood','electronics','other'"},
		{"quantity_unit_enum", "'ton','places','pallet','m3'"},
		{"divisibility_enum", "'ftl','ltl','dogruz'"},
		{"packaging_enum", "'bulk','pallets','bags','barrels','rolls','boxes','liquid','oversized'"},
		{"vat_mode_enum", "'yes','no','unspecified'"},
		{"cargo_type_enum", "'domestic','international'"},
		{"loading_type_enum", "'ftl','ltl','partial'"},
		{"customs_type_enum", "'export','import','transit'"},
		{"rate_mode_enum", "'negotiable','fixed','announcement'"},
		{"currency_enum", "'UZS','USD'"},
		{"cargo_status_enum", "'active','archived','completed'"},
		{"waypoint_type_enum", "'border','customs','reload','passthrough'"},
		{"vehicle_composition_enum", "'tractor_semi','truck_trailer'"},
		{"transport_status_enum", "'active','archived'"},
		{"warehouse_type_enum", "'regular','cold','customs'"},
		{"heating_type_enum", "'heated','unheated','open','closed'"},
		{"warehouse_status_enum", "'active','archived'"},
		{"media_entity_type_enum", "'cargo','transport','warehouse','vehicle'"},
		{"media_file_type_enum", "'photo','document'"},
		{"contact_listing_type_enum", "'cargo','transport','warehouse'"},
		{"token_transaction_type_enum", "'credit','debit'"},
		{"token_reason_enum", "'registration','purchase','contact_view','manual'"},
		{"payment_type_enum", "'tokens','subscription','listing','boost'"},
		{"payment_status_enum", "'pending','paid','failed','cancelled'"},
		{"payment_currency_enum", "'UZS','USD'"},
		{"subscription_plan_enum", "'week','month','year'"},
		{"favorite_listing_type_enum", "'cargo','transport','warehouse'"},
	}
	for _, e := range enums {
		sql := fmt.Sprintf(`DO $$ BEGIN IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname='%s') THEN CREATE TYPE %s AS ENUM (%s); END IF; END $$;`, e.name, e.name, e.vals)
		if err := db.Exec(sql).Error; err != nil {
			return fmt.Errorf("enum %s: %w", e.name, err)
		}
	}
	// Дополнения к существующим enum (идемпотентно).
	alters := []string{
		"ALTER TYPE media_file_type_enum ADD VALUE IF NOT EXISTS 'video'",
	}
	for _, a := range alters {
		if err := db.Exec(a).Error; err != nil {
			return fmt.Errorf("alter enum: %w", err)
		}
	}
	return nil
}

func seedPricingConfig(db *gorm.DB) error {
	type priceSeed struct {
		Key          string
		Label        string
		TokensAmount int
		DurationDays int
		PriceUZS     float64
		PriceUSD     float64
	}
	seeds := []priceSeed{
		{Key: "tokens_registration", Label: "Бонус при регистрации", TokensAmount: 5},
		{Key: "tokens_mini", Label: "Мини пакет (5 токенов)", TokensAmount: 5, PriceUZS: 15000, PriceUSD: 1.5},
		{Key: "tokens_basic", Label: "Базовый пакет (20 токенов)", TokensAmount: 20, PriceUZS: 50000, PriceUSD: 5},
		{Key: "tokens_standard", Label: "Стандарт пакет (50 токенов)", TokensAmount: 50, PriceUZS: 100000, PriceUSD: 9},
		{Key: "tokens_max", Label: "Макс пакет (100 токенов)", TokensAmount: 100, PriceUZS: 180000, PriceUSD: 16},
		{Key: "sub_week", Label: "Подписка на неделю", DurationDays: 7, PriceUZS: 50000, PriceUSD: 5},
		{Key: "sub_month", Label: "Подписка на месяц", DurationDays: 30, PriceUZS: 150000, PriceUSD: 13},
		{Key: "sub_year", Label: "Подписка на год", DurationDays: 365, PriceUZS: 1000000, PriceUSD: 90},
		{Key: "listing_paid", Label: "Платное объявление", PriceUZS: 30000, PriceUSD: 3},
		{Key: "boost_1day", Label: "Буст на 1 день", DurationDays: 1, PriceUZS: 20000, PriceUSD: 2},
		{Key: "boost_3day", Label: "Буст на 3 дня", DurationDays: 3, PriceUZS: 50000, PriceUSD: 4.5},
		{Key: "boost_7day", Label: "Буст на 7 дней", DurationDays: 7, PriceUZS: 100000, PriceUSD: 9},
	}
	for _, s := range seeds {
		var existing model.PricingConfig
		err := db.Where("key = ?", s.Key).First(&existing).Error
		if err != nil {
			// не найден — создаём
			db.Create(&model.PricingConfig{
				Key: s.Key, Label: s.Label,
				TokensAmount: s.TokensAmount, DurationDays: s.DurationDays,
				PriceUZS: s.PriceUZS, PriceUSD: s.PriceUSD, IsActive: true,
			})
		} else if existing.PriceUZS == 0 && s.PriceUZS > 0 {
			// существует, но цена 0 — проставляем дефолт
			db.Model(&existing).Updates(map[string]interface{}{
				"price_uzs": s.PriceUZS, "price_usd": s.PriceUSD,
			})
		}
	}
	return nil
}

func requestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		log.Info().
			Str("method", c.Request.Method).
			Str("path", c.Request.URL.Path).
			Int("status", c.Writer.Status()).
			Dur("latency", time.Since(start)).
			Msg("req")
	}
}
