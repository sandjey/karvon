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

type orgTypeItem struct {
	Key   string `json:"key"`
	Label string `json:"label"`
}

var orgTypesByLang = map[string][]orgTypeItem{
	"ru": {
		{Key: "ooo", Label: "ООО — Общество с ограниченной ответственностью"},
		{Key: "ao", Label: "АО — Акционерное общество"},
		{Key: "ip", Label: "ИП — Индивидуальный предприниматель"},
		{Key: "ltd", Label: "Ltd — Частная компания с ограниченной ответственностью"},
		{Key: "gmbh", Label: "GmbH — Общество с ограниченной ответственностью (Германия)"},
		{Key: "co_ltd", Label: "Co. Ltd — Компания с ограниченной ответственностью"},
		{Key: "mchj", Label: "МЧДЖ — Общество с ограниченной ответственностью (Узбекистан)"},
		{Key: "xk", Label: "ЧП — Частное предприятие"},
		{Key: "oaj", Label: "ОАЖ — Открытое акционерное общество (Узбекистан)"},
		{Key: "yat", Label: "ЯТ — Индивидуальный предприниматель (Узбекистан)"},
		{Key: "qmj", Label: "КМЖ — Общество с дополнительной ответственностью"},
	},
	"uz": {
		{Key: "ooo", Label: "OOO — Mas'uliyati cheklangan jamiyat (Rossiya)"},
		{Key: "ao", Label: "AO — Aksiyadorlik jamiyati (Rossiya)"},
		{Key: "ip", Label: "YT — Yakka tartibdagi tadbirkor (Rossiya)"},
		{Key: "ltd", Label: "Ltd — Xususiy cheklangan kompaniya"},
		{Key: "gmbh", Label: "GmbH — Cheklangan javobgarlikli jamiyat (Germaniya)"},
		{Key: "co_ltd", Label: "Co. Ltd — Cheklangan mas'uliyatli kompaniya"},
		{Key: "mchj", Label: "MCHJ — Mas'uliyati cheklangan jamiyat"},
		{Key: "xk", Label: "XK — Xususiy korxona"},
		{Key: "oaj", Label: "OAJ — Ochiq aksiyadorlik jamiyati"},
		{Key: "yat", Label: "YAT — Yakka tartibdagi tadbirkor"},
		{Key: "qmj", Label: "QMJ — Qo'shimcha mas'uliyatli jamiyat"},
	},
	"en": {
		{Key: "ooo", Label: "LLC — Limited Liability Company (Russian form)"},
		{Key: "ao", Label: "JSC — Joint Stock Company (Russian form)"},
		{Key: "ip", Label: "SE — Sole Entrepreneur (Russian form)"},
		{Key: "ltd", Label: "Ltd — Private Limited Company"},
		{Key: "gmbh", Label: "GmbH — Gesellschaft mit beschränkter Haftung (Germany)"},
		{Key: "co_ltd", Label: "Co. Ltd — Company Limited"},
		{Key: "mchj", Label: "LLC — Mas'uliyati Cheklangan Jamiyat (Uzbek form)"},
		{Key: "xk", Label: "PE — Private Enterprise (Uzbek form)"},
		{Key: "oaj", Label: "OJSC — Open Joint Stock Company (Uzbek form)"},
		{Key: "yat", Label: "IE — Individual Entrepreneur (Uzbek form)"},
		{Key: "qmj", Label: "ALC — Additional Liability Company (Uzbek form)"},
	},
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

	orgTypes := orgTypesByLang["ru"]
	if v, ok := orgTypesByLang[lang]; ok {
		orgTypes = v
	}

	OK(c, gin.H{
		"map_tiles_url": h.mapTilesURL,
		"languages":     []string{"ru", "uz", "en"},
		"categories":    items,
		"org_types":     orgTypes,
	})
}
