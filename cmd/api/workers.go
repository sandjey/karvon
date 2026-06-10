package main

import (
	"context"
	"time"

	"github.com/rs/zerolog/log"

	"karvon/internal/repository"
)

// startBackgroundWorkers запускает фоновые задачи (тикеры).
func startBackgroundWorkers(cargoRepo *repository.CargoRepo, warehouseRepo *repository.WarehouseRepo, otpRepo *repository.OTPRepo) {
	go boostExpiryWorker(cargoRepo, warehouseRepo)
	go otpCleanupWorker(otpRepo)
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
