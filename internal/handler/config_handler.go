package handler

import "github.com/gin-gonic/gin"

// ConfigHandler отдаёт публичную конфигурацию фронтенду (адрес карты, языки).
type ConfigHandler struct {
	mapTilesURL string
}

func NewConfigHandler(mapTilesURL string) *ConfigHandler {
	return &ConfigHandler{mapTilesURL: mapTilesURL}
}

func (h *ConfigHandler) RegisterRoutes(rg *gin.RouterGroup) {
	rg.GET("/config", h.Config)
}

func (h *ConfigHandler) Config(c *gin.Context) {
	OK(c, gin.H{
		"map_tiles_url": h.mapTilesURL,
		"languages":     []string{"ru", "uz"},
	})
}
