package main

import (
	"context"
	"time"

	"github.com/rs/zerolog/log"

	"karvon/internal/model"
	"karvon/internal/repository"
)

// startBackgroundWorkers запускает фоновые задачи (тикеры).
func startBackgroundWorkers(cargoRepo *repository.CargoRepo, notifRepo *repository.NotificationRepo, otpRepo *repository.OTPRepo) {
	go expireListingsWorker(cargoRepo, notifRepo)
	go otpCleanupWorker(otpRepo)
}

// expireListingsWorker каждые 30 минут архивирует истёкшие грузы и шлёт уведомления.
func expireListingsWorker(cargoRepo *repository.CargoRepo, notifRepo *repository.NotificationRepo) {
	ticker := time.NewTicker(30 * time.Minute)
	defer ticker.Stop()
	for {
		ctx := context.Background()
		expired, err := cargoRepo.ExpireDue(ctx)
		if err != nil {
			log.Warn().Err(err).Msg("expire listings worker error")
		} else if len(expired) > 0 {
			for _, c := range expired {
				body := "Срок действия вашего объявления истёк, оно перемещено в архив."
				_ = notifRepo.Create(ctx, &model.Notification{
					UserID: c.UserID,
					Type:   "listing_expired",
					Title:  "Объявление архивировано",
					Body:   &body,
				})
			}
			log.Info().Int("count", len(expired)).Msg("archived expired cargo listings")
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
