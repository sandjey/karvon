package handler

import (
	"github.com/gin-gonic/gin"

	"karvon/internal/repository"
)

type StatsHandler struct {
	repo *repository.StatsRepo
}

func NewStatsHandler(repo *repository.StatsRepo) *StatsHandler {
	return &StatsHandler{repo: repo}
}

func (h *StatsHandler) RegisterRoutes(rg *gin.RouterGroup) {
	rg.GET("/stats/public", h.Public)
}

func (h *StatsHandler) Public(c *gin.Context) {
	stats, err := h.repo.PublicStats(c.Request.Context())
	if err != nil {
		InternalError(c)
		return
	}
	OK(c, stats)
}
