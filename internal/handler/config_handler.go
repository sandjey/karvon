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

type dictItem struct {
	Key   string `json:"key"`
	Label string `json:"label"`
}

var orgTypesByLang = map[string][]dictItem{
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

// ── Справочники (cargo) ──────────────────────────────────────────────────────

var quantityUnitsByLang = map[string][]dictItem{
	"ru": {
		{Key: "ton", Label: "Тонна"},
		{Key: "places", Label: "Мест"},
		{Key: "pallet", Label: "Паллет"},
		{Key: "m3", Label: "Кубометр (м³)"},
	},
	"uz": {
		{Key: "ton", Label: "Tonna"},
		{Key: "places", Label: "Birlik"},
		{Key: "pallet", Label: "Pallet"},
		{Key: "m3", Label: "Kubometr (m³)"},
	},
	"en": {
		{Key: "ton", Label: "Ton"},
		{Key: "places", Label: "Pieces"},
		{Key: "pallet", Label: "Pallet"},
		{Key: "m3", Label: "Cubic meter (m³)"},
	},
}

var divisibilityByLang = map[string][]dictItem{
	"ru": {
		{Key: "ftl", Label: "FTL — Полный груз"},
		{Key: "ltl", Label: "LTL — Сборный груз"},
		{Key: "dogruz", Label: "Догруз"},
	},
	"uz": {
		{Key: "ftl", Label: "FTL — To'liq yuk"},
		{Key: "ltl", Label: "LTL — Yig'ma yuk"},
		{Key: "dogruz", Label: "Doghruz (qo'shimcha yuk)"},
	},
	"en": {
		{Key: "ftl", Label: "FTL — Full Truck Load"},
		{Key: "ltl", Label: "LTL — Less than Truck Load"},
		{Key: "dogruz", Label: "Co-loading (partial fill)"},
	},
}

var packagingByLang = map[string][]dictItem{
	"ru": {
		{Key: "bulk", Label: "Навалом"},
		{Key: "pallets", Label: "Паллеты"},
		{Key: "bags", Label: "Мешки"},
		{Key: "barrels", Label: "Бочки"},
		{Key: "rolls", Label: "Рулоны"},
		{Key: "boxes", Label: "Ящики / коробки"},
		{Key: "liquid", Label: "Наливной груз"},
		{Key: "oversized", Label: "Негабаритный груз"},
	},
	"uz": {
		{Key: "bulk", Label: "Vagonaga"},
		{Key: "pallets", Label: "Palletlarda"},
		{Key: "bags", Label: "Xaltalar"},
		{Key: "barrels", Label: "Bochkalar"},
		{Key: "rolls", Label: "Rulonlar"},
		{Key: "boxes", Label: "Qutular"},
		{Key: "liquid", Label: "Suyuq yuk"},
		{Key: "oversized", Label: "Haddan tashqari o'lchamli yuk"},
	},
	"en": {
		{Key: "bulk", Label: "Bulk"},
		{Key: "pallets", Label: "Pallets"},
		{Key: "bags", Label: "Bags / sacks"},
		{Key: "barrels", Label: "Barrels"},
		{Key: "rolls", Label: "Rolls"},
		{Key: "boxes", Label: "Boxes / crates"},
		{Key: "liquid", Label: "Liquid cargo"},
		{Key: "oversized", Label: "Oversized cargo"},
	},
}

var bodyTypesByLang = map[string][]dictItem{
	"ru": {
		{Key: "tent", Label: "Тентованный"},
		{Key: "ref", Label: "Рефрижератор"},
		{Key: "container", Label: "Контейнер"},
		{Key: "cistern", Label: "Цистерна"},
		{Key: "carcarrier", Label: "Автовоз"},
		{Key: "board", Label: "Бортовой"},
		{Key: "isothermal", Label: "Изотермический"},
	},
	"uz": {
		{Key: "tent", Label: "Tentli"},
		{Key: "ref", Label: "Refrizherator"},
		{Key: "container", Label: "Konteyner"},
		{Key: "cistern", Label: "Sisterna"},
		{Key: "carcarrier", Label: "Avtovozu"},
		{Key: "board", Label: "Bortli"},
		{Key: "isothermal", Label: "Izotermik"},
	},
	"en": {
		{Key: "tent", Label: "Tarpaulin / curtainsider"},
		{Key: "ref", Label: "Refrigerated"},
		{Key: "container", Label: "Container"},
		{Key: "cistern", Label: "Tanker"},
		{Key: "carcarrier", Label: "Car carrier"},
		{Key: "board", Label: "Flatbed / open"},
		{Key: "isothermal", Label: "Isothermal"},
	},
}

var loadingMethodsByLang = map[string][]dictItem{
	"ru": {
		{Key: "self", Label: "Самовывоз"},
		{Key: "loader", Label: "Погрузчик"},
		{Key: "crane", Label: "Кран"},
	},
	"uz": {
		{Key: "self", Label: "O'z transport"},
		{Key: "loader", Label: "Yuk ko'taruvchi"},
		{Key: "crane", Label: "Kran"},
	},
	"en": {
		{Key: "self", Label: "Self pickup"},
		{Key: "loader", Label: "Forklift"},
		{Key: "crane", Label: "Crane"},
	},
}

var deliveryGeographyByLang = map[string][]dictItem{
	"ru": {
		{Key: "uz", Label: "Узбекистан (внутренние перевозки)"},
		{Key: "cis", Label: "СНГ"},
		{Key: "world", Label: "Международные (весь мир)"},
	},
	"uz": {
		{Key: "uz", Label: "O'zbekiston (ichki tashish)"},
		{Key: "cis", Label: "MDH"},
		{Key: "world", Label: "Xalqaro (butun dunyo)"},
	},
	"en": {
		{Key: "uz", Label: "Uzbekistan (domestic)"},
		{Key: "cis", Label: "CIS countries"},
		{Key: "world", Label: "International (worldwide)"},
	},
}

var permitsByLang = map[string][]dictItem{
	"ru": {
		{Key: "TIR", Label: "TIR — Книжка МДП"},
		{Key: "CMR", Label: "CMR — Международная накладная"},
		{Key: "phyto", Label: "Фитосанитарный сертификат"},
		{Key: "CITES", Label: "CITES — Разрешение на перевозку животных"},
		{Key: "EKMT", Label: "ЕКМТ — Многостороннее разрешение"},
	},
	"uz": {
		{Key: "TIR", Label: "TIR — TIR daftarchasi"},
		{Key: "CMR", Label: "CMR — Xalqaro yuk xati"},
		{Key: "phyto", Label: "Fitosanitariya sertifikati"},
		{Key: "CITES", Label: "CITES — Hayvonlarni tashish ruxsati"},
		{Key: "EKMT", Label: "EKMT — Ko'p tomonlama ruxsatnoma"},
	},
	"en": {
		{Key: "TIR", Label: "TIR — TIR Carnet"},
		{Key: "CMR", Label: "CMR — International consignment note"},
		{Key: "phyto", Label: "Phytosanitary certificate"},
		{Key: "CITES", Label: "CITES — Wildlife transport permit"},
		{Key: "EKMT", Label: "ECMT — Multilateral road transport permit"},
	},
}

var incotermsByLang = map[string][]dictItem{
	"ru": {
		{Key: "EXW", Label: "EXW — Франко-завод"},
		{Key: "FCA", Label: "FCA — Франко-перевозчик"},
		{Key: "CPT", Label: "CPT — Перевозка оплачена до"},
		{Key: "CIP", Label: "CIP — Перевозка и страхование оплачены до"},
		{Key: "DAP", Label: "DAP — Поставка в месте назначения"},
		{Key: "DPU", Label: "DPU — Поставка на место выгрузки"},
		{Key: "DDP", Label: "DDP — Поставка с оплатой пошлин"},
		{Key: "FAS", Label: "FAS — Франко вдоль борта судна"},
		{Key: "FOB", Label: "FOB — Франко-борт"},
		{Key: "CFR", Label: "CFR — Стоимость и фрахт"},
		{Key: "CIF", Label: "CIF — Стоимость, страхование и фрахт"},
	},
	"uz": {
		{Key: "EXW", Label: "EXW — Zavoddan"},
		{Key: "FCA", Label: "FCA — Tashuvchiga"},
		{Key: "CPT", Label: "CPT — Tashish to'langan"},
		{Key: "CIP", Label: "CIP — Tashish va sug'urta to'langan"},
		{Key: "DAP", Label: "DAP — Belgilangan joyga yetkazib berish"},
		{Key: "DPU", Label: "DPU — Tushirib qo'yish joyiga yetkazib berish"},
		{Key: "DDP", Label: "DDP — Bojlar to'langan holda yetkazib berish"},
		{Key: "FAS", Label: "FAS — Kema bortiga qadar"},
		{Key: "FOB", Label: "FOB — Kema bortida"},
		{Key: "CFR", Label: "CFR — Narx va frakht"},
		{Key: "CIF", Label: "CIF — Narx, sug'urta va frakht"},
	},
	"en": {
		{Key: "EXW", Label: "EXW — Ex Works"},
		{Key: "FCA", Label: "FCA — Free Carrier"},
		{Key: "CPT", Label: "CPT — Carriage Paid To"},
		{Key: "CIP", Label: "CIP — Carriage and Insurance Paid To"},
		{Key: "DAP", Label: "DAP — Delivered at Place"},
		{Key: "DPU", Label: "DPU — Delivered at Place Unloaded"},
		{Key: "DDP", Label: "DDP — Delivered Duty Paid"},
		{Key: "FAS", Label: "FAS — Free Alongside Ship"},
		{Key: "FOB", Label: "FOB — Free on Board"},
		{Key: "CFR", Label: "CFR — Cost and Freight"},
		{Key: "CIF", Label: "CIF — Cost, Insurance and Freight"},
	},
}

var adrClassesByLang = map[string][]dictItem{
	"ru": {
		{Key: "1", Label: "Класс 1 — Взрывчатые вещества"},
		{Key: "2", Label: "Класс 2 — Газы"},
		{Key: "3", Label: "Класс 3 — Легковоспламеняющиеся жидкости"},
		{Key: "4", Label: "Класс 4 — Легковоспламеняющиеся твёрдые вещества"},
		{Key: "5", Label: "Класс 5 — Окисляющие вещества и органические пероксиды"},
		{Key: "6", Label: "Класс 6 — Токсичные и инфекционные вещества"},
		{Key: "7", Label: "Класс 7 — Радиоактивные материалы"},
		{Key: "8", Label: "Класс 8 — Едкие и коррозионные вещества"},
		{Key: "9", Label: "Класс 9 — Прочие опасные вещества"},
	},
	"uz": {
		{Key: "1", Label: "1-sinf — Portlovchi moddalar"},
		{Key: "2", Label: "2-sinf — Gazlar"},
		{Key: "3", Label: "3-sinf — Oson alangalanuvchi suyuqliklar"},
		{Key: "4", Label: "4-sinf — Oson alangalanuvchi qattiq moddalar"},
		{Key: "5", Label: "5-sinf — Oksidlovchi moddalar va organik peroksidlar"},
		{Key: "6", Label: "6-sinf — Zaharli va yuqumli moddalar"},
		{Key: "7", Label: "7-sinf — Radioaktiv materiallar"},
		{Key: "8", Label: "8-sinf — Kislotali va korroziv moddalar"},
		{Key: "9", Label: "9-sinf — Boshqa xavfli moddalar"},
	},
	"en": {
		{Key: "1", Label: "Class 1 — Explosives"},
		{Key: "2", Label: "Class 2 — Gases"},
		{Key: "3", Label: "Class 3 — Flammable liquids"},
		{Key: "4", Label: "Class 4 — Flammable solids"},
		{Key: "5", Label: "Class 5 — Oxidising substances & organic peroxides"},
		{Key: "6", Label: "Class 6 — Toxic and infectious substances"},
		{Key: "7", Label: "Class 7 — Radioactive materials"},
		{Key: "8", Label: "Class 8 — Corrosives"},
		{Key: "9", Label: "Class 9 — Miscellaneous dangerous goods"},
	},
}

// ── Справочники (склад) ──────────────────────────────────────────────────────

var warehouseTypesByLang = map[string][]dictItem{
	"ru": {
		{Key: "regular", Label: "Обычный склад"},
		{Key: "cold", Label: "Холодильный склад"},
		{Key: "customs", Label: "Таможенный склад"},
	},
	"uz": {
		{Key: "regular", Label: "Oddiy ombor"},
		{Key: "cold", Label: "Sovuq ombor"},
		{Key: "customs", Label: "Bojxona ombori"},
	},
	"en": {
		{Key: "regular", Label: "Regular warehouse"},
		{Key: "cold", Label: "Cold storage"},
		{Key: "customs", Label: "Bonded / customs warehouse"},
	},
}

var heatingTypesByLang = map[string][]dictItem{
	"ru": {
		{Key: "heated", Label: "Отапливаемый"},
		{Key: "unheated", Label: "Неотапливаемый"},
		{Key: "open", Label: "Открытая площадка"},
		{Key: "closed", Label: "Закрытый (без отопления)"},
	},
	"uz": {
		{Key: "heated", Label: "Isitilgan"},
		{Key: "unheated", Label: "Isitilmagan"},
		{Key: "open", Label: "Ochiq maydon"},
		{Key: "closed", Label: "Yopiq (isitishsiz)"},
	},
	"en": {
		{Key: "heated", Label: "Heated"},
		{Key: "unheated", Label: "Unheated"},
		{Key: "open", Label: "Open yard"},
		{Key: "closed", Label: "Closed (no heating)"},
	},
}

var coldChamberTypesByLang = map[string][]dictItem{
	"ru": {
		{Key: "shock_freeze", Label: "Шоковая заморозка"},
		{Key: "freeze", Label: "Заморозка (−18°C и ниже)"},
		{Key: "regular_cold", Label: "Холодильник (0…+8°C)"},
		{Key: "cooling", Label: "Охлаждение (+8…+18°C)"},
	},
	"uz": {
		{Key: "shock_freeze", Label: "Tez muzlatish"},
		{Key: "freeze", Label: "Muzlatish (−18°C va pastroq)"},
		{Key: "regular_cold", Label: "Sovutgich (0…+8°C)"},
		{Key: "cooling", Label: "Sovutish (+8…+18°C)"},
	},
	"en": {
		{Key: "shock_freeze", Label: "Blast freezing"},
		{Key: "freeze", Label: "Deep freeze (−18°C and below)"},
		{Key: "regular_cold", Label: "Refrigerated (0…+8°C)"},
		{Key: "cooling", Label: "Cooling (+8…+18°C)"},
	},
}

var storageTypesByLang = map[string][]dictItem{
	"ru": {
		{Key: "dry", Label: "Сухое хранение"},
		{Key: "humid", Label: "Влажное хранение"},
		{Key: "heated", Label: "Отапливаемое"},
		{Key: "cooled", Label: "Охлаждённое"},
		{Key: "cold", Label: "Холодное"},
		{Key: "frozen", Label: "Замороженное"},
	},
	"uz": {
		{Key: "dry", Label: "Quruq saqlash"},
		{Key: "humid", Label: "Nam saqlash"},
		{Key: "heated", Label: "Isitilgan"},
		{Key: "cooled", Label: "Sovutilgan"},
		{Key: "cold", Label: "Sovuq"},
		{Key: "frozen", Label: "Muzlatilgan"},
	},
	"en": {
		{Key: "dry", Label: "Dry storage"},
		{Key: "humid", Label: "Humid storage"},
		{Key: "heated", Label: "Heated storage"},
		{Key: "cooled", Label: "Cooled storage"},
		{Key: "cold", Label: "Cold storage"},
		{Key: "frozen", Label: "Frozen storage"},
	},
}

var warehouseServicesByLang = map[string][]dictItem{
	"ru": {
		{Key: "packing", Label: "Упаковка"},
		{Key: "unpacking", Label: "Распаковка"},
		{Key: "labeling", Label: "Маркировка"},
		{Key: "sorting", Label: "Сортировка"},
		{Key: "palleting", Label: "Паллетирование"},
		{Key: "customs_clearance", Label: "Таможенное оформление"},
		{Key: "loading", Label: "Погрузка"},
		{Key: "unloading", Label: "Разгрузка"},
		{Key: "transportation", Label: "Транспортировка (доставка)"},
		{Key: "photo_report", Label: "Фотоотчёт о товаре"},
		{Key: "inventory", Label: "Инвентаризация"},
	},
	"uz": {
		{Key: "packing", Label: "Qadoqlash"},
		{Key: "unpacking", Label: "Qadoqdan chiqarish"},
		{Key: "labeling", Label: "Belgilash"},
		{Key: "sorting", Label: "Saralash"},
		{Key: "palleting", Label: "Palletlash"},
		{Key: "customs_clearance", Label: "Bojxona rasmiylashtiruvi"},
		{Key: "loading", Label: "Yuklash"},
		{Key: "unloading", Label: "Tushirish"},
		{Key: "transportation", Label: "Tashish (yetkazib berish)"},
		{Key: "photo_report", Label: "Tovar haqida fotohisobot"},
		{Key: "inventory", Label: "Inventarizatsiya"},
	},
	"en": {
		{Key: "packing", Label: "Packing"},
		{Key: "unpacking", Label: "Unpacking"},
		{Key: "labeling", Label: "Labeling"},
		{Key: "sorting", Label: "Sorting"},
		{Key: "palleting", Label: "Palletizing"},
		{Key: "customs_clearance", Label: "Customs clearance"},
		{Key: "loading", Label: "Loading"},
		{Key: "unloading", Label: "Unloading"},
		{Key: "transportation", Label: "Transportation (delivery)"},
		{Key: "photo_report", Label: "Photo report on goods"},
		{Key: "inventory", Label: "Inventory management"},
	},
}

var infrastructureByLang = map[string][]dictItem{
	"ru": {
		{Key: "forklift", Label: "Вилочный погрузчик"},
		{Key: "crane", Label: "Кран"},
		{Key: "truck_ramp", Label: "Рампа / доклевеллер"},
		{Key: "railway", Label: "Железнодорожный подъезд"},
		{Key: "security", Label: "Охрана (круглосуточно)"},
		{Key: "cctv", Label: "Видеонаблюдение"},
		{Key: "fire_alarm", Label: "Пожарная сигнализация"},
		{Key: "scales", Label: "Весы"},
		{Key: "parking", Label: "Парковка для грузовиков"},
		{Key: "fumigation", Label: "Фумигация"},
	},
	"uz": {
		{Key: "forklift", Label: "Vilkali yuk ko'targich"},
		{Key: "crane", Label: "Kran"},
		{Key: "truck_ramp", Label: "Ramp / dok leveller"},
		{Key: "railway", Label: "Temir yo'l kirishi"},
		{Key: "security", Label: "Qo'riqlash (24/7)"},
		{Key: "cctv", Label: "Videokuzatuv"},
		{Key: "fire_alarm", Label: "Yong'in signalizatsiyasi"},
		{Key: "scales", Label: "Tarozalar"},
		{Key: "parking", Label: "Yuk mashinalar uchun to'xtash joyi"},
		{Key: "fumigation", Label: "Fumigatsiya"},
	},
	"en": {
		{Key: "forklift", Label: "Forklift"},
		{Key: "crane", Label: "Crane"},
		{Key: "truck_ramp", Label: "Loading dock / dock leveler"},
		{Key: "railway", Label: "Railway siding"},
		{Key: "security", Label: "24/7 security"},
		{Key: "cctv", Label: "CCTV surveillance"},
		{Key: "fire_alarm", Label: "Fire alarm system"},
		{Key: "scales", Label: "Weighing scales"},
		{Key: "parking", Label: "Truck parking"},
		{Key: "fumigation", Label: "Fumigation"},
	},
}

var specializationByLang = map[string][]dictItem{
	"ru": {
		{Key: "food", Label: "Продукты питания"},
		{Key: "pharma", Label: "Фармацевтика"},
		{Key: "chemical", Label: "Химия"},
		{Key: "textile", Label: "Текстиль"},
		{Key: "metal", Label: "Металл"},
		{Key: "electronics", Label: "Электроника"},
		{Key: "auto", Label: "Автозапчасти"},
		{Key: "hazmat", Label: "Опасные грузы"},
		{Key: "oversized", Label: "Негабаритные грузы"},
		{Key: "documents", Label: "Документы / архивы"},
	},
	"uz": {
		{Key: "food", Label: "Oziq-ovqat"},
		{Key: "pharma", Label: "Farmatsevtika"},
		{Key: "chemical", Label: "Kimyo"},
		{Key: "textile", Label: "To'qimachilik"},
		{Key: "metal", Label: "Metall"},
		{Key: "electronics", Label: "Elektronika"},
		{Key: "auto", Label: "Avto ehtiyot qismlari"},
		{Key: "hazmat", Label: "Xavfli yuкlar"},
		{Key: "oversized", Label: "Haddan tashqari o'lchamli yuкlar"},
		{Key: "documents", Label: "Hujjatlar / arxivlar"},
	},
	"en": {
		{Key: "food", Label: "Food products"},
		{Key: "pharma", Label: "Pharmaceuticals"},
		{Key: "chemical", Label: "Chemicals"},
		{Key: "textile", Label: "Textile"},
		{Key: "metal", Label: "Metal"},
		{Key: "electronics", Label: "Electronics"},
		{Key: "auto", Label: "Auto parts"},
		{Key: "hazmat", Label: "Hazardous goods"},
		{Key: "oversized", Label: "Oversized cargo"},
		{Key: "documents", Label: "Documents / archives"},
	},
}

var regionsByLang = map[string][]dictItem{
	"ru": {
		{Key: "tashkent_city", Label: "Ташкент (город)"},
		{Key: "tashkent_region", Label: "Ташкентская область"},
		{Key: "andijan", Label: "Андижанская область"},
		{Key: "fergana", Label: "Ферганская область"},
		{Key: "namangan", Label: "Наманганская область"},
		{Key: "samarkand", Label: "Самаркандская область"},
		{Key: "bukhara", Label: "Бухарская область"},
		{Key: "kashkadarya", Label: "Кашкадарьинская область"},
		{Key: "surkhandarya", Label: "Сурхандарьинская область"},
		{Key: "navoi", Label: "Навоийская область"},
		{Key: "khorezm", Label: "Хорезмская область"},
		{Key: "syrdarya", Label: "Сырдарьинская область"},
		{Key: "jizzakh", Label: "Джизакская область"},
		{Key: "karakalpakstan", Label: "Республика Каракалпакстан"},
	},
	"uz": {
		{Key: "tashkent_city", Label: "Toshkent (shahar)"},
		{Key: "tashkent_region", Label: "Toshkent viloyati"},
		{Key: "andijan", Label: "Andijon viloyati"},
		{Key: "fergana", Label: "Farg'ona viloyati"},
		{Key: "namangan", Label: "Namangan viloyati"},
		{Key: "samarkand", Label: "Samarqand viloyati"},
		{Key: "bukhara", Label: "Buxoro viloyati"},
		{Key: "kashkadarya", Label: "Qashqadaryo viloyati"},
		{Key: "surkhandarya", Label: "Surxondaryo viloyati"},
		{Key: "navoi", Label: "Navoiy viloyati"},
		{Key: "khorezm", Label: "Xorazm viloyati"},
		{Key: "syrdarya", Label: "Sirdaryo viloyati"},
		{Key: "jizzakh", Label: "Jizzax viloyati"},
		{Key: "karakalpakstan", Label: "Qoraqalpog'iston Respublikasi"},
	},
	"en": {
		{Key: "tashkent_city", Label: "Tashkent (city)"},
		{Key: "tashkent_region", Label: "Tashkent region"},
		{Key: "andijan", Label: "Andijan region"},
		{Key: "fergana", Label: "Fergana region"},
		{Key: "namangan", Label: "Namangan region"},
		{Key: "samarkand", Label: "Samarkand region"},
		{Key: "bukhara", Label: "Bukhara region"},
		{Key: "kashkadarya", Label: "Kashkadarya region"},
		{Key: "surkhandarya", Label: "Surkhandarya region"},
		{Key: "navoi", Label: "Navoi region"},
		{Key: "khorezm", Label: "Khorezm region"},
		{Key: "syrdarya", Label: "Syrdarya region"},
		{Key: "jizzakh", Label: "Jizzakh region"},
		{Key: "karakalpakstan", Label: "Republic of Karakalpakstan"},
	},
}

func dictFor(m map[string][]dictItem, lang string) []dictItem {
	if v, ok := m[lang]; ok {
		return v
	}
	return m["ru"]
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
		"org_types":     dictFor(orgTypesByLang, lang),
		// cargo
		"quantity_units":     dictFor(quantityUnitsByLang, lang),
		"divisibility":       dictFor(divisibilityByLang, lang),
		"packaging_types":    dictFor(packagingByLang, lang),
		"body_types":         dictFor(bodyTypesByLang, lang),
		"loading_methods":    dictFor(loadingMethodsByLang, lang),
		"delivery_geography": dictFor(deliveryGeographyByLang, lang),
		"permits":            dictFor(permitsByLang, lang),
		"incoterms":          dictFor(incotermsByLang, lang),
		"adr_classes":        dictFor(adrClassesByLang, lang),
		// warehouse
		"warehouse_types":   dictFor(warehouseTypesByLang, lang),
		"heating_types":     dictFor(heatingTypesByLang, lang),
		"cold_chamber_types": dictFor(coldChamberTypesByLang, lang),
		"storage_types":     dictFor(storageTypesByLang, lang),
		"warehouse_services": dictFor(warehouseServicesByLang, lang),
		"infrastructure":    dictFor(infrastructureByLang, lang),
		"specialization":    dictFor(specializationByLang, lang),
		"regions":           dictFor(regionsByLang, lang),
	})
}
