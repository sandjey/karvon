package main

import (
	"context"
	"time"

	"github.com/rs/zerolog/log"

	"karvon/internal/model"
	"karvon/internal/repository"
)

// startBackgroundWorkers запускает фоновые задачи (тикеры).
func startBackgroundWorkers(
	cargoRepo *repository.CargoRepo,
	warehouseRepo *repository.WarehouseRepo,
	otpRepo *repository.OTPRepo,
	paymentRepo *repository.PaymentRepo,
	notifRepo *repository.NotificationRepo,
) {
	go boostExpiryWorker(cargoRepo, warehouseRepo)
	go otpCleanupWorker(otpRepo)
	go subscriptionExpiryWorker(paymentRepo, notifRepo)
}

// boostExpiryWorker каждый час снимает истёкшие бусты с грузов и складов.
func boostExpiryWorker(cargoRepo *repository.CargoRepo, warehouseRepo *repository.WarehouseRepo) {
	ticker := time.NewTicker(time.Hour)
	defer ticker.Stop()
	for {
		ctx := context.Background()
		nc, err := cargoRepo.ExpireBoosts(ctx)
		if err != nil {
			log.Warn().Err(err).Msg("cargo boost expiry error")
		}
		nw, err := warehouseRepo.ExpireBoosts(ctx)
		if err != nil {
			log.Warn().Err(err).Msg("warehouse boost expiry error")
		}
		if nc+nw > 0 {
			log.Info().Int64("cargo", nc).Int64("warehouse", nw).Msg("boosts expired")
		}
		<-ticker.C
	}
}

// otpCleanupWorker раз в сутки удаляет OTP-коды старше 24 часов.
func otpCleanupWorker(otpRepo *repository.OTPRepo) {
	ticker := time.NewTicker(24 * time.Hour)
	defer ticker.Stop()
	for {
		n, err := otpRepo.DeleteOlderThan(context.Background(), 24*time.Hour)
		if err != nil {
			log.Warn().Err(err).Msg("otp cleanup worker error")
		} else if n > 0 {
			log.Info().Int64("deleted", n).Msg("cleaned old OTP codes")
		}
		<-ticker.C
	}
}

// subscriptionExpiryWorker раз в сутки отправляет уведомление за 3 дня до конца подписки.
func subscriptionExpiryWorker(paymentRepo *repository.PaymentRepo, notifRepo *repository.NotificationRepo) {
	ticker := time.NewTicker(24 * time.Hour)
	defer ticker.Stop()
	for {
		ctx := context.Background()
		now := time.Now()
		from := now.Add(3 * 24 * time.Hour)
		to := now.Add(4 * 24 * time.Hour)
		subs, err := paymentRepo.FindExpiringSoon(ctx, from, to)
		if err != nil {
			log.Warn().Err(err).Msg("subscription expiry worker error")
		} else {
			for _, s := range subs {
				body := "Ваша подписка истекает через 3 дня. Продлите её, чтобы сохранить безлимитный доступ к контактам."
				_ = notifRepo.Create(ctx, &model.Notification{
					UserID: s.UserID,
					Type:   "subscription_expiring",
					Title:  "Подписка заканчивается",
					Body:   &body,
				})
			}
			if len(subs) > 0 {
				log.Info().Int("count", len(subs)).Msg("subscription expiry notifications sent")
			}
		}
		<-ticker.C
	}
}
