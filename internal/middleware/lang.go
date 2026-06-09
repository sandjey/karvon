package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"

	"karvon/pkg/i18n"
)

var supportedLangs = map[string]bool{
	"ru": true,
	"uz": true,
	"en": true,
}

// Lang middleware определяет язык запроса и сохраняет в context.
// Приоритет: ?lang= > Accept-Language > default (ru).
func Lang() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set(i18n.LangKey, detectLang(c))
		c.Next()
	}
}

func detectLang(c *gin.Context) string {
	if q := c.Query("lang"); supportedLangs[q] {
		return q
	}
	if accept := c.GetHeader("Accept-Language"); accept != "" {
		for _, part := range strings.Split(accept, ",") {
			tag := strings.ToLower(strings.TrimSpace(strings.Split(strings.Split(part, ";")[0], "-")[0]))
			if supportedLangs[tag] {
				return tag
			}
		}
	}
	return i18n.DefaultLang
}
