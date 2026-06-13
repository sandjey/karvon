package handler

import (
	"github.com/gin-gonic/gin"

	"karvon/internal/repository"
	"karvon/pkg/i18n"
)

type ConfigHandler struct {
	mapTilesURL  string
	categoryRepo *repository.CategoryRepo
}

func NewConfigHandler(mapTilesURL string, categoryRepo *repository.CategoryRepo) *ConfigHandler {
	return &ConfigHandler{mapTilesURL: mapTilesURL, categoryRepo: categoryRepo}
}

func (h *ConfigHandler) RegisterRoutes(rg *gin.RouterGroup) {
	rg.GET("/config", h.Config)
}

func (h *ConfigHandler) Config(c *gin.Context) {
	lang := i18n.Lang(c)

	categories, _ := h.categoryRepo.ListActive(c.Request.Context())

	type categoryItem struct {
		Key   string `json:"key"`
		Label string `json:"label"`
	}
	items := make([]categoryItem, 0, len(categories))
	for _, cat := range categories {
		label := cat.LabelRu
		switch lang {
		case "uz":
			label = cat.LabelUz
		case "en":
			label = cat.LabelEn
		}
		items = append(items, categoryItem{Key: cat.Key, Label: label})
	}

	OK(c, gin.H{
		"map_tiles_url": h.mapTilesURL,
		"languages":     []string{"ru", "uz", "en"},
		"categories":    items,
	})
}
