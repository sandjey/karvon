package geo

import (
	_ "embed"
	"bufio"
	"bytes"
	"sort"
	"strconv"
	"strings"
	"unicode"
)

//go:embed data/countryInfo.txt
var countryInfoData []byte

//go:embed data/admin1Codes.txt
var admin1CodesData []byte

//go:embed data/cities15000.txt
var citiesData []byte

// Country holds country metadata.
type Country struct {
	Code      string  `json:"code"`
	NameRu    string  `json:"name_ru"`
	NameEn    string  `json:"name_en"`
	Continent string  `json:"continent,omitempty"`
	Lat       float64 `json:"lat"`
	Lng       float64 `json:"lng"`
}

// City holds city / settlement metadata.
type City struct {
	NameRu      string  `json:"name_ru"`
	NameEn      string  `json:"name_en"`
	CountryCode string  `json:"country_code"`
	CountryName string  `json:"country"`
	Region      string  `json:"region,omitempty"`
	Lat         float64 `json:"lat"`
	Lng         float64 `json:"lng"`
	Population  int     `json:"population"`

	// cyrAlts holds all pure-Russian Cyrillic tokens from alternateNames.
	// Not exported; used only for search to handle all Russian spelling variants.
	cyrAlts []string
}

var (
	Countries []Country
	Cities    []City

	countryByCode map[string]Country
	admin1Names   map[string]string // "CC.code" → name
)

func init() {
	loadAdmin1()
	loadCountries()
	loadCities()
	loadCountryCoords()
}

// loadCountryCoords sets Lat/Lng on each Country using the most-populous city
// in that country (countryInfo.txt has no coordinates).
func loadCountryCoords() {
	maxPop := make(map[string]int)
	coords := make(map[string][2]float64)
	for _, city := range Cities {
		if city.Population > maxPop[city.CountryCode] {
			maxPop[city.CountryCode] = city.Population
			coords[city.CountryCode] = [2]float64{city.Lat, city.Lng}
		}
	}
	for i := range Countries {
		if c, ok := coords[Countries[i].Code]; ok {
			Countries[i].Lat = c[0]
			Countries[i].Lng = c[1]
		}
	}
	// rebuild countryByCode with updated coords
	for _, c := range Countries {
		countryByCode[c.Code] = c
	}
}

func loadAdmin1() {
	admin1Names = make(map[string]string)
	sc := bufio.NewScanner(bytes.NewReader(admin1CodesData))
	for sc.Scan() {
		line := sc.Text()
		if line == "" {
			continue
		}
		f := strings.Split(line, "\t")
		if len(f) < 2 {
			continue
		}
		admin1Names[f[0]] = f[1]
	}
}

func loadCountries() {
	countryByCode = make(map[string]Country)
	sc := bufio.NewScanner(bytes.NewReader(countryInfoData))
	for sc.Scan() {
		line := sc.Text()
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		f := strings.Split(line, "\t")
		if len(f) < 9 {
			continue
		}
		code := f[0]
		nameEn := f[4]
		continent := f[8]

		c := Country{
			Code:      code,
			NameEn:    nameEn,
			NameRu:    ruCountryName(code, nameEn),
			Continent: continent,
		}
		Countries = append(Countries, c)
		countryByCode[code] = c
	}
	sort.Slice(Countries, func(i, j int) bool {
		return Countries[i].NameRu < Countries[j].NameRu
	})
}

func loadCities() {
	sc := bufio.NewScanner(bytes.NewReader(citiesData))
	sc.Buffer(make([]byte, 1024*1024), 1024*1024)
	for sc.Scan() {
		line := sc.Text()
		if line == "" {
			continue
		}
		f := strings.Split(line, "\t")
		if len(f) < 15 {
			continue
		}
		lat, err1 := strconv.ParseFloat(f[4], 64)
		lng, err2 := strconv.ParseFloat(f[5], 64)
		pop, err3 := strconv.Atoi(f[14])
		if err1 != nil || err2 != nil || err3 != nil {
			continue
		}

		nameEn := f[1]
		asciiName := f[2]
		altNames := f[3]
		cc := f[8]
		admin1Key := cc + "." + f[10]

		cyrAlts := collectRuTokens(altNames)
		nameRu := pickDisplayName(cyrAlts, asciiName)
		if nameRu == "" {
			nameRu = nameEn
		}

		country := countryByCode[cc]
		region := admin1Names[admin1Key]

		Cities = append(Cities, City{
			NameEn:      nameEn,
			NameRu:      nameRu,
			CountryCode: cc,
			CountryName: country.NameRu,
			Region:      region,
			Lat:         lat,
			Lng:         lng,
			Population:  pop,
			cyrAlts:     cyrAlts,
		})
	}
}

// ruCyrillicLetters is the set of standard Russian Cyrillic letters (33 letters).
// We exclude non-Russian Cyrillic letters used in Kazakh (Ә,Ғ,Қ,Ң,Ө,Ұ,Ү),
// Ukrainian (Ї,І,Є,Ґ), Belarusian (Ў), and other scripts.
var ruCyrillicLetters = func() map[rune]bool {
	const ru = "АБВГДЕЁЖЗИЙКЛМНОПРСТУФХЦЧШЩЪЫЬЭЮЯабвгдеёжзийклмнопрстуфхцчшщъыьэюя"
	m := make(map[rune]bool, 66)
	for _, r := range ru {
		m[r] = true
	}
	return m
}()

// isPureRussian returns true if every letter in s is a standard Russian Cyrillic letter.
func isPureRussian(s string) bool {
	if s == "" {
		return false
	}
	for _, r := range s {
		if unicode.IsLetter(r) && !ruCyrillicLetters[r] {
			return false
		}
	}
	return true
}

// collectRuTokens returns all pure-Russian Cyrillic tokens from the GeoNames alternateNames field.
// Single-word tokens (no spaces, no hyphens) are included; multi-word tokens starting
// with a capital letter are also included. Duplicates are skipped.
func collectRuTokens(altNames string) []string {
	if altNames == "" {
		return nil
	}
	seen := make(map[string]bool)
	var out []string
	for _, token := range strings.Split(altNames, ",") {
		token = strings.TrimSpace(token)
		if token == "" || seen[token] || !isPureRussian(token) {
			continue
		}
		seen[token] = true
		out = append(out, token)
	}
	return out
}

// ruTranslit maps a lowercase ASCII byte to the expected Russian Cyrillic rune.
// 'y' deliberately maps to 'ы' — this correctly prefers "Алматы" over "Алмати".
var ruTranslit = map[byte]rune{
	'a': 'а', 'b': 'б', 'c': 'к', 'd': 'д', 'e': 'е',
	'f': 'ф', 'g': 'г', 'h': 'х', 'i': 'и', 'j': 'ж',
	'k': 'к', 'l': 'л', 'm': 'м', 'n': 'н', 'o': 'о',
	'p': 'п', 'r': 'р', 's': 'с', 't': 'т', 'u': 'у',
	'v': 'в', 'w': 'в', 'x': 'к', 'y': 'ы', 'z': 'з',
}

// translitScore counts absolute character matches between the Cyrillic token and
// the expected transliteration of asciiName (no normalization by length).
// Longer correct matches beat shorter ones; length proximity is a separate tiebreaker.
func translitScore(token, asciiName string) int {
	cyrRunes := []rune(strings.ToLower(strings.ReplaceAll(strings.ReplaceAll(token, " ", ""), "-", "")))
	asciiBytes := []byte(strings.ToLower(strings.ReplaceAll(asciiName, " ", "")))
	n := len(cyrRunes)
	if len(asciiBytes) < n {
		n = len(asciiBytes)
	}
	matches := 0
	for i := 0; i < n; i++ {
		if expected, ok := ruTranslit[asciiBytes[i]]; ok {
			if cyrRunes[i] == expected {
				matches++
			}
		}
	}
	return matches
}

// pickDisplayName chooses the best Russian display name from a list of pure-Russian tokens.
//
// Scoring (single-word tokens without hyphens):
//  1. translitScore: how many of the first ≤6 chars match ASCII transliteration (higher = better).
//  2. On score tie: token whose rune-count is closest to asciiName (no-space) length wins.
//  3. On further tie: first found wins (stable, no relying on GeoNames list order).
//
// Multi-word tokens compete on the same scoring; single-word wins on equal total score.
func pickDisplayName(tokens []string, asciiName string) string {
	refLen := len([]rune(strings.ReplaceAll(asciiName, " ", "")))

	type candidate struct {
		text  string
		ts    int // translitScore
		diff  int // abs(runeCount - refLen)
		multi bool
	}

	var best *candidate
	for _, t := range tokens {
		nosp := strings.ReplaceAll(t, " ", "")
		runeLen := len([]rune(nosp))
		diff := runeLen - refLen
		if diff < 0 {
			diff = -diff
		}
		ts := translitScore(t, asciiName)
		multi := strings.ContainsRune(t, ' ') || strings.ContainsRune(t, '-')

		c := &candidate{text: t, ts: ts, diff: diff, multi: multi}

		if best == nil {
			best = c
			continue
		}
		// Compare: higher translitScore wins; same score → smaller diff wins;
		// same diff → single-word beats multi; same everything → keep first.
		switch {
		case c.ts > best.ts:
			best = c
		case c.ts == best.ts && c.diff < best.diff:
			best = c
		case c.ts == best.ts && c.diff == best.diff && !c.multi && best.multi:
			best = c
		}
	}

	if best == nil {
		return ""
	}
	return best.text
}

func hasCyrillic(s string) bool {
	for _, r := range s {
		if unicode.In(r, unicode.Cyrillic) {
			return true
		}
	}
	return false
}

func isMostlyCyrillic(s string) bool {
	if s == "" {
		return false
	}
	total, cyr := 0, 0
	for _, r := range s {
		if unicode.IsLetter(r) {
			total++
			if unicode.In(r, unicode.Cyrillic) {
				cyr++
			}
		}
	}
	return total > 0 && float64(cyr)/float64(total) >= 0.6
}

func lower(s string) string { return strings.ToLower(s) }

func score(name, q string) int {
	n := lower(name)
	if n == q {
		return 4
	}
	if strings.HasPrefix(n, q) {
		return 3
	}
	if strings.Contains(n, " "+q) {
		return 2
	}
	if strings.Contains(n, q) {
		return 1
	}
	return 0
}

func bestCityScore(city City, q string) int {
	s := score(city.NameRu, q)
	if e := score(city.NameEn, q); e > s {
		s = e
	}
	// also search all Cyrillic alternate names so users can type any known Russian form
	for _, alt := range city.cyrAlts {
		if a := score(alt, q); a > s {
			s = a
		}
	}
	return s
}

func bestCountryScore(c Country, q string) int {
	s := score(c.NameRu, q)
	if e := score(c.NameEn, q); e > s {
		s = e
	}
	return s
}

// SearchCountries returns countries matching q.
// Empty q returns all countries sorted alphabetically by Russian name.
func SearchCountries(q string, limit int) []Country {
	if limit <= 0 {
		limit = 250
	}
	q = strings.TrimSpace(lower(q))
	if q == "" {
		out := make([]Country, len(Countries))
		copy(out, Countries)
		if limit < len(out) {
			out = out[:limit]
		}
		return out
	}

	type sc struct {
		c Country
		s int
	}
	var res []sc
	for _, c := range Countries {
		if s := bestCountryScore(c, q); s > 0 {
			res = append(res, sc{c, s})
		}
	}
	sort.SliceStable(res, func(i, j int) bool { return res[i].s > res[j].s })
	out := make([]Country, 0, len(res))
	for _, r := range res {
		out = append(out, r.c)
	}
	if limit < len(out) {
		out = out[:limit]
	}
	return out
}

// SearchCities returns cities matching q, optionally filtered by countryCode.
// Results ordered: score DESC, population DESC.
func SearchCities(q, countryCode string, limit int) []City {
	if limit <= 0 || limit > 100 {
		limit = 10
	}
	q = strings.TrimSpace(lower(q))
	ccUp := strings.ToUpper(countryCode)

	type sc struct {
		c City
		s int
		p int
	}
	var res []sc
	for _, city := range Cities {
		if ccUp != "" && city.CountryCode != ccUp {
			continue
		}
		var s int
		if q == "" {
			s = 1
		} else {
			s = bestCityScore(city, q)
		}
		if s > 0 {
			res = append(res, sc{city, s, city.Population})
		}
	}
	sort.SliceStable(res, func(i, j int) bool {
		if res[i].s != res[j].s {
			return res[i].s > res[j].s
		}
		return res[i].p > res[j].p
	})
	out := make([]City, 0, limit)
	for i, r := range res {
		if i >= limit {
			break
		}
		out = append(out, r.c)
	}
	return out
}

// ruCountryName returns the Russian name for a country ISO code.
// Falls back to the English name if not found.
func ruCountryName(code, fallback string) string {
	if name, ok := ruNames[code]; ok {
		return name
	}
	return fallback
}

// ruNames maps ISO 3166-1 alpha-2 codes to standard Russian country names.
var ruNames = map[string]string{
	"AF": "Афганистан",
	"AX": "Аландские острова",
	"AL": "Албания",
	"DZ": "Алжир",
	"AS": "Американское Самоа",
	"AD": "Андорра",
	"AO": "Ангола",
	"AI": "Ангилья",
	"AQ": "Антарктида",
	"AG": "Антигуа и Барбуда",
	"AR": "Аргентина",
	"AM": "Армения",
	"AW": "Аруба",
	"AU": "Австралия",
	"AT": "Австрия",
	"AZ": "Азербайджан",
	"BS": "Багамы",
	"BH": "Бахрейн",
	"BD": "Бангладеш",
	"BB": "Барбадос",
	"BY": "Беларусь",
	"BE": "Бельгия",
	"BZ": "Белиз",
	"BJ": "Бенин",
	"BM": "Бермуды",
	"BT": "Бутан",
	"BO": "Боливия",
	"BQ": "Бонэйр, Синт-Эстатиус и Саба",
	"BA": "Босния и Герцеговина",
	"BW": "Ботсвана",
	"BV": "Остров Буве",
	"BR": "Бразилия",
	"IO": "Британская территория в Индийском океане",
	"BN": "Бруней",
	"BG": "Болгария",
	"BF": "Буркина-Фасо",
	"BI": "Бурунди",
	"CV": "Кабо-Верде",
	"KH": "Камбоджа",
	"CM": "Камерун",
	"CA": "Канада",
	"KY": "Острова Кайман",
	"CF": "Центральноафриканская Республика",
	"TD": "Чад",
	"CL": "Чили",
	"CN": "Китай",
	"CX": "Остров Рождества",
	"CC": "Кокосовые острова",
	"CO": "Колумбия",
	"KM": "Коморы",
	"CG": "Республика Конго",
	"CD": "ДР Конго",
	"CK": "Острова Кука",
	"CR": "Коста-Рика",
	"CI": "Кот-д'Ивуар",
	"HR": "Хорватия",
	"CU": "Куба",
	"CW": "Кюрасао",
	"CY": "Кипр",
	"CZ": "Чехия",
	"DK": "Дания",
	"DJ": "Джибути",
	"DM": "Доминика",
	"DO": "Доминиканская Республика",
	"EC": "Эквадор",
	"EG": "Египет",
	"SV": "Сальвадор",
	"GQ": "Экваториальная Гвинея",
	"ER": "Эритрея",
	"EE": "Эстония",
	"SZ": "Эсватини",
	"ET": "Эфиопия",
	"FK": "Фолклендские острова",
	"FO": "Фарерские острова",
	"FJ": "Фиджи",
	"FI": "Финляндия",
	"FR": "Франция",
	"GF": "Французская Гвиана",
	"PF": "Французская Полинезия",
	"TF": "Французские Южные территории",
	"GA": "Габон",
	"GM": "Гамбия",
	"GE": "Грузия",
	"DE": "Германия",
	"GH": "Гана",
	"GI": "Гибралтар",
	"GR": "Греция",
	"GL": "Гренландия",
	"GD": "Гренада",
	"GP": "Гваделупа",
	"GU": "Гуам",
	"GT": "Гватемала",
	"GG": "Гернси",
	"GN": "Гвинея",
	"GW": "Гвинея-Бисау",
	"GY": "Гайана",
	"HT": "Гаити",
	"HM": "Остров Херд и острова Макдональд",
	"VA": "Ватикан",
	"HN": "Гондурас",
	"HK": "Гонконг",
	"HU": "Венгрия",
	"IS": "Исландия",
	"IN": "Индия",
	"ID": "Индонезия",
	"IR": "Иран",
	"IQ": "Ирак",
	"IE": "Ирландия",
	"IM": "Остров Мэн",
	"IL": "Израиль",
	"IT": "Италия",
	"JM": "Ямайка",
	"JP": "Япония",
	"JE": "Джерси",
	"JO": "Иордания",
	"KZ": "Казахстан",
	"KE": "Кения",
	"KI": "Кирибати",
	"KP": "Северная Корея",
	"KR": "Южная Корея",
	"XK": "Косово",
	"KW": "Кувейт",
	"KG": "Кыргызстан",
	"LA": "Лаос",
	"LV": "Латвия",
	"LB": "Ливан",
	"LS": "Лесото",
	"LR": "Либерия",
	"LY": "Ливия",
	"LI": "Лихтенштейн",
	"LT": "Литва",
	"LU": "Люксембург",
	"MO": "Макао",
	"MG": "Мадагаскар",
	"MW": "Малави",
	"MY": "Малайзия",
	"MV": "Мальдивы",
	"ML": "Мали",
	"MT": "Мальта",
	"MH": "Маршалловы острова",
	"MQ": "Мартиника",
	"MR": "Мавритания",
	"MU": "Маврикий",
	"YT": "Майотта",
	"MX": "Мексика",
	"FM": "Микронезия",
	"MD": "Молдова",
	"MC": "Монако",
	"MN": "Монголия",
	"ME": "Черногория",
	"MS": "Монтсеррат",
	"MA": "Марокко",
	"MZ": "Мозамбик",
	"MM": "Мьянма",
	"NA": "Намибия",
	"NR": "Науру",
	"NP": "Непал",
	"NL": "Нидерланды",
	"NC": "Новая Каледония",
	"NZ": "Новая Зеландия",
	"NI": "Никарагуа",
	"NE": "Нигер",
	"NG": "Нигерия",
	"NU": "Ниуэ",
	"NF": "Остров Норфолк",
	"MK": "Северная Македония",
	"MP": "Северные Марианские острова",
	"NO": "Норвегия",
	"OM": "Оман",
	"PK": "Пакистан",
	"PW": "Палау",
	"PS": "Палестина",
	"PA": "Панама",
	"PG": "Папуа — Новая Гвинея",
	"PY": "Парагвай",
	"PE": "Перу",
	"PH": "Филиппины",
	"PN": "Острова Питкэрн",
	"PL": "Польша",
	"PT": "Португалия",
	"PR": "Пуэрто-Рико",
	"QA": "Катар",
	"RE": "Реюньон",
	"RO": "Румыния",
	"RU": "Россия",
	"RW": "Руанда",
	"BL": "Сен-Бартелеми",
	"SH": "Острова Святой Елены",
	"KN": "Сент-Китс и Невис",
	"LC": "Сент-Люсия",
	"MF": "Сен-Мартен",
	"PM": "Сен-Пьер и Микелон",
	"VC": "Сент-Винсент и Гренадины",
	"WS": "Самоа",
	"SM": "Сан-Марино",
	"ST": "Сан-Томе и Принсипи",
	"SA": "Саудовская Аравия",
	"SN": "Сенегал",
	"RS": "Сербия",
	"SC": "Сейшелы",
	"SL": "Сьерра-Леоне",
	"SG": "Сингапур",
	"SX": "Синт-Мартен",
	"SK": "Словакия",
	"SI": "Словения",
	"SB": "Соломоновы острова",
	"SO": "Сомали",
	"ZA": "ЮАР",
	"GS": "Южная Георгия",
	"SS": "Южный Судан",
	"ES": "Испания",
	"LK": "Шри-Ланка",
	"SD": "Судан",
	"SR": "Суринам",
	"SJ": "Шпицберген и Ян-Майен",
	"SE": "Швеция",
	"CH": "Швейцария",
	"SY": "Сирия",
	"TW": "Тайвань",
	"TJ": "Таджикистан",
	"TZ": "Танзания",
	"TH": "Таиланд",
	"TL": "Восточный Тимор",
	"TG": "Того",
	"TK": "Токелау",
	"TO": "Тонга",
	"TT": "Тринидад и Тобаго",
	"TN": "Тунис",
	"TR": "Турция",
	"TM": "Туркменистан",
	"TC": "Острова Тёркс и Кайкос",
	"TV": "Тувалу",
	"UG": "Уганда",
	"UA": "Украина",
	"AE": "ОАЭ",
	"GB": "Великобритания",
	"US": "США",
	"UM": "Внешние острова США",
	"UY": "Уругвай",
	"UZ": "Узбекистан",
	"VU": "Вануату",
	"VE": "Венесуэла",
	"VN": "Вьетнам",
	"VG": "Британские Виргинские острова",
	"VI": "Виргинские острова США",
	"WF": "Уоллис и Футуна",
	"EH": "Западная Сахара",
	"YE": "Йемен",
	"ZM": "Замбия",
	"ZW": "Зимбабве",
}
