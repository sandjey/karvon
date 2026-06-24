package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"ctm/internal/dto"
	"ctm/internal/model"
	"ctm/internal/service"
	"ctm/pkg/i18n"
)

type AdminHandler struct {
	svc *service.AdminService
}

func NewAdminHandler(svc *service.AdminService) *AdminHandler {
	return &AdminHandler{svc: svc}
}

// RegisterRoutes: login публичен, остальное — только super_admin.
func (h *AdminHandler) RegisterRoutes(rg *gin.RouterGroup, auth, superAdmin gin.HandlerFunc) {
	rg.POST("/admin/login", h.Login)
	rg.GET("/admin-free-mode", h.FreeModeUI)

	g := rg.Group("/admin", auth, superAdmin)
	g.GET("/free-mode", h.GetFreeMode)
	g.POST("/free-mode", h.SetFreeMode)
	g.GET("/dashboard", h.Dashboard)
	g.GET("/users", h.Users)
	g.GET("/users/:id", h.User)
	g.PATCH("/users/:id/block", h.Block)
	g.POST("/users/:id/topup", h.Topup)
	g.GET("/listings", h.Listings)
	g.DELETE("/listings/:type/:id", h.DeleteListing)
	g.PATCH("/listings/:type/:id/block", h.BlockListing)
	g.PATCH("/listings/:type/:id/unblock", h.UnblockListing)
	g.GET("/companies", h.Companies)
	g.GET("/payments", h.Payments)
	g.GET("/pricing", h.Pricing)
	g.POST("/pricing", h.CreatePricing)
	g.PATCH("/pricing/:key", h.UpdatePricing)
	g.DELETE("/pricing/:key", h.DeletePricing)
	g.GET("/moderators", h.ListModerators)
	g.POST("/moderators", h.CreateModerator)
	g.DELETE("/moderators/:id", h.DeleteModerator)
	g.GET("/categories", h.ListCategories)
	g.POST("/categories", h.CreateCategory)
	g.PATCH("/categories/:id", h.UpdateCategory)
	g.DELETE("/categories/:id", h.DeleteCategory)
}

func (h *AdminHandler) Login(c *gin.Context) {
	var req dto.AdminLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, "VALIDATION_ERROR", i18n.T(c, "VALIDATION_ERROR"))
		return
	}
	tokens, err := h.svc.Login(c.Request.Context(), req.Login, req.Password)
	if err != nil {
		FailCode(c, http.StatusUnauthorized, "INVALID_CREDENTIALS")
		return
	}
	OKMsg(c, tokens, "AUTH_SUCCESS")
}

func (h *AdminHandler) Dashboard(c *gin.Context) {
	m, err := h.svc.Dashboard(c.Request.Context(), c.DefaultQuery("period", "30d"))
	if err != nil {
		InternalError(c)
		return
	}
	OK(c, m)
}

func (h *AdminHandler) Users(c *gin.Context) {
	page, perPage := parsePagination(c)
	list, total, err := h.svc.Users(c.Request.Context(), c.Query("q"), c.Query("role"), (page-1)*perPage, perPage)
	if err != nil {
		InternalError(c)
		return
	}
	Paginated(c, list, int(total), page, perPage)
}

func (h *AdminHandler) User(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		FailCode(c, http.StatusBadRequest, "VALIDATION_ERROR")
		return
	}
	u, err := h.svc.User(c.Request.Context(), id)
	if err != nil {
		FailCode(c, http.StatusNotFound, "USER_NOT_FOUND")
		return
	}
	OK(c, u)
}

func (h *AdminHandler) Block(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		FailCode(c, http.StatusBadRequest, "VALIDATION_ERROR")
		return
	}
	var req dto.BlockRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, "VALIDATION_ERROR", i18n.T(c, "VALIDATION_ERROR"))
		return
	}
	if err := h.svc.SetBlocked(c.Request.Context(), id, req.Blocked, req.Reason); err != nil {
		InternalError(c)
		return
	}
	OKMsg(c, nil, "USER_BLOCK_UPDATED")
}

func (h *AdminHandler) Topup(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		FailCode(c, http.StatusBadRequest, "VALIDATION_ERROR")
		return
	}
	var req dto.TopupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, "VALIDATION_ERROR", i18n.T(c, "VALIDATION_ERROR"))
		return
	}
	adminID := c.MustGet("user_id").(uuid.UUID)
	if err := h.svc.TopupTokens(c.Request.Context(), id, req.Amount, adminID, req.Note); err != nil {
		FailCode(c, http.StatusNotFound, "USER_NOT_FOUND")
		return
	}
	OKMsg(c, nil, "TOKENS_TOPPED_UP")
}

func (h *AdminHandler) Listings(c *gin.Context) {
	lt := c.DefaultQuery("type", "cargo")
	if lt != "cargo" && lt != "warehouse" {
		FailCode(c, http.StatusBadRequest, "VALIDATION_ERROR")
		return
	}
	page, perPage := parsePagination(c)
	if lt == "warehouse" {
		list, total, err := h.svc.Warehouses(c.Request.Context(), (page-1)*perPage, perPage)
		if err != nil {
			InternalError(c)
			return
		}
		Paginated(c, list, int(total), page, perPage)
		return
	}
	list, total, err := h.svc.Cargo(c.Request.Context(), (page-1)*perPage, perPage)
	if err != nil {
		InternalError(c)
		return
	}
	Paginated(c, list, int(total), page, perPage)
}

func (h *AdminHandler) DeleteListing(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		FailCode(c, http.StatusBadRequest, "VALIDATION_ERROR")
		return
	}
	if err := h.svc.DeleteListing(c.Request.Context(), c.Param("type"), id); err != nil {
		FailCode(c, http.StatusNotFound, "LISTING_NOT_FOUND")
		return
	}
	OKMsg(c, nil, "LISTING_DELETED")
}

func (h *AdminHandler) BlockListing(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		FailCode(c, http.StatusBadRequest, "VALIDATION_ERROR")
		return
	}
	if err := h.svc.BlockListing(c.Request.Context(), c.Param("type"), id); err != nil {
		FailCode(c, http.StatusNotFound, "LISTING_NOT_FOUND")
		return
	}
	OKMsg(c, nil, "LISTING_BLOCK_UPDATED")
}

func (h *AdminHandler) UnblockListing(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		FailCode(c, http.StatusBadRequest, "VALIDATION_ERROR")
		return
	}
	if err := h.svc.UnblockListing(c.Request.Context(), c.Param("type"), id); err != nil {
		FailCode(c, http.StatusNotFound, "LISTING_NOT_FOUND")
		return
	}
	OKMsg(c, nil, "LISTING_BLOCK_UPDATED")
}

func (h *AdminHandler) Companies(c *gin.Context) {
	page, perPage := parsePagination(c)
	list, total, err := h.svc.Companies(c.Request.Context(), c.Query("status"), (page-1)*perPage, perPage)
	if err != nil {
		InternalError(c)
		return
	}
	Paginated(c, list, int(total), page, perPage)
}

func (h *AdminHandler) Payments(c *gin.Context) {
	page, perPage := parsePagination(c)
	list, total, err := h.svc.Payments(c.Request.Context(), (page-1)*perPage, perPage)
	if err != nil {
		InternalError(c)
		return
	}
	resp := make([]model.PaymentOrderResponse, len(list))
	for i := range list {
		resp[i] = model.ToPaymentResponse(&list[i])
	}
	Paginated(c, resp, int(total), page, perPage)
}

func (h *AdminHandler) Pricing(c *gin.Context) {
	list, err := h.svc.Pricing(c.Request.Context())
	if err != nil {
		InternalError(c)
		return
	}
	OK(c, list)
}

func (h *AdminHandler) UpdatePricing(c *gin.Context) {
	var req dto.UpdatePricingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, "VALIDATION_ERROR", i18n.T(c, "VALIDATION_ERROR"))
		return
	}
	fields := map[string]interface{}{}
	if req.Label != nil {
		fields["label"] = *req.Label
	}
	if req.PriceUZS != nil {
		fields["price_uzs"] = *req.PriceUZS
	}
	if req.PriceUSD != nil {
		fields["price_usd"] = *req.PriceUSD
	}
	if req.TokensAmount != nil {
		fields["tokens_amount"] = *req.TokensAmount
	}
	if req.DurationDays != nil {
		fields["duration_days"] = *req.DurationDays
	}
	if req.IsActive != nil {
		fields["is_active"] = *req.IsActive
	}
	adminID := c.MustGet("user_id").(uuid.UUID)
	if err := h.svc.UpdatePricing(c.Request.Context(), c.Param("key"), fields, adminID); err != nil {
		FailCode(c, http.StatusNotFound, "NOT_FOUND")
		return
	}
	OKMsg(c, nil, "PRICING_UPDATED")
}

func (h *AdminHandler) CreatePricing(c *gin.Context) {
	var req dto.CreatePricingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, "VALIDATION_ERROR", i18n.T(c, "VALIDATION_ERROR"))
		return
	}
	active := true
	if req.IsActive != nil {
		active = *req.IsActive
	}
	adminID := c.MustGet("user_id").(uuid.UUID)
	p := &model.PricingConfig{
		Key:          req.Key,
		Label:        req.Label,
		PriceUZS:     req.PriceUZS,
		PriceUSD:     req.PriceUSD,
		TokensAmount: req.TokensAmount,
		DurationDays: req.DurationDays,
		IsActive:     active,
	}
	if err := h.svc.CreatePricing(c.Request.Context(), p, adminID); err != nil {
		if isErr(err, service.ErrAlreadyExists) {
			FailCode(c, http.StatusConflict, "ALREADY_EXISTS")
			return
		}
		InternalError(c)
		return
	}
	CreatedMsg(c, p, "PRICING_CREATED")
}

func (h *AdminHandler) DeletePricing(c *gin.Context) {
	if err := h.svc.DeletePricing(c.Request.Context(), c.Param("key")); err != nil {
		InternalError(c)
		return
	}
	OKMsg(c, nil, "PRICING_DELETED")
}

func (h *AdminHandler) ListModerators(c *gin.Context) {
	page, perPage := parsePagination(c)
	list, total, err := h.svc.Users(c.Request.Context(), "", "moderator", (page-1)*perPage, perPage)
	if err != nil {
		InternalError(c)
		return
	}
	Paginated(c, list, int(total), page, perPage)
}

func (h *AdminHandler) CreateModerator(c *gin.Context) {
	var req dto.CreateModeratorRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, "VALIDATION_ERROR", i18n.T(c, "VALIDATION_ERROR"))
		return
	}
	u, err := h.svc.CreateModerator(c.Request.Context(), req.Phone, req.Name, req.Login, req.Password)
	if err != nil {
		InternalError(c)
		return
	}
	CreatedMsg(c, u, "MODERATOR_CREATED")
}

func (h *AdminHandler) DeleteModerator(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		FailCode(c, http.StatusBadRequest, "VALIDATION_ERROR")
		return
	}
	if err := h.svc.DeleteModerator(c.Request.Context(), id); err != nil {
		FailCode(c, http.StatusNotFound, "USER_NOT_FOUND")
		return
	}
	OKMsg(c, nil, "MODERATOR_DELETED")
}

func (h *AdminHandler) ListCategories(c *gin.Context) {
	list, err := h.svc.ListCategories(c.Request.Context())
	if err != nil {
		InternalError(c)
		return
	}
	OK(c, list)
}

func (h *AdminHandler) CreateCategory(c *gin.Context) {
	var req dto.CreateCategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, "VALIDATION_ERROR", i18n.T(c, "VALIDATION_ERROR"))
		return
	}
	cat, err := h.svc.CreateCategory(c.Request.Context(), req)
	if err != nil {
		if isErr(err, service.ErrAlreadyExists) {
			FailCode(c, http.StatusConflict, "ALREADY_EXISTS")
			return
		}
		InternalError(c)
		return
	}
	CreatedMsg(c, cat, "CATEGORY_CREATED")
}

func (h *AdminHandler) UpdateCategory(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		FailCode(c, http.StatusBadRequest, "VALIDATION_ERROR")
		return
	}
	var req dto.UpdateCategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, "VALIDATION_ERROR", i18n.T(c, "VALIDATION_ERROR"))
		return
	}
	if err := h.svc.UpdateCategory(c.Request.Context(), id, req); err != nil {
		InternalError(c)
		return
	}
	OKMsg(c, nil, "CATEGORY_UPDATED")
}

func (h *AdminHandler) DeleteCategory(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		FailCode(c, http.StatusBadRequest, "VALIDATION_ERROR")
		return
	}
	if err := h.svc.DeleteCategory(c.Request.Context(), id); err != nil {
		InternalError(c)
		return
	}
	OKMsg(c, nil, "CATEGORY_DELETED")
}

func (h *AdminHandler) GetFreeMode(c *gin.Context) {
	list, err := h.svc.Pricing(c.Request.Context())
	if err != nil {
		InternalError(c)
		return
	}
	flags := map[string]bool{
		"all":       false,
		"cargo":     false,
		"warehouse": false,
		"contacts":  false,
		"carriers":  false,
		"companies": false,
	}
	for _, p := range list {
		switch p.Key {
		case "system:free_mode":
			flags["all"] = p.TokensAmount > 0
		case "system:free_cargo":
			flags["cargo"] = p.TokensAmount > 0
		case "system:free_warehouse":
			flags["warehouse"] = p.TokensAmount > 0
		case "system:free_contacts":
			flags["contacts"] = p.TokensAmount > 0
		case "system:free_carriers":
			flags["carriers"] = p.TokensAmount > 0
		case "system:free_companies":
			flags["companies"] = p.TokensAmount > 0
		}
	}
	OK(c, flags)
}

func (h *AdminHandler) SetFreeMode(c *gin.Context) {
	var req struct {
		Key     string `json:"key"     binding:"required,oneof=all cargo warehouse contacts carriers companies"`
		Enabled bool   `json:"enabled"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, "VALIDATION_ERROR", i18n.T(c, "VALIDATION_ERROR"))
		return
	}
	adminID := c.MustGet("user_id").(uuid.UUID)
	val := 0
	if req.Enabled {
		val = 1
	}

	keyMap := map[string]string{
		"all":       "system:free_mode",
		"cargo":     "system:free_cargo",
		"warehouse": "system:free_warehouse",
		"contacts":  "system:free_contacts",
		"carriers":  "system:free_carriers",
		"companies": "system:free_companies",
	}

	pricingKey := keyMap[req.Key]
	if err := h.svc.UpdatePricing(c.Request.Context(), pricingKey, map[string]interface{}{"tokens_amount": val}, adminID); err != nil {
		InternalError(c)
		return
	}

	msgKey := "FREE_MODE_DISABLED"
	if req.Enabled {
		msgKey = "FREE_MODE_ENABLED"
	}
	OKMsg(c, gin.H{"key": req.Key, "enabled": req.Enabled}, msgKey)
}

func (h *AdminHandler) FreeModeUI(c *gin.Context) {
	c.Header("Content-Type", "text/html; charset=utf-8")
	c.String(200, `<!DOCTYPE html>
<html lang="ru">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>Бесплатный режим — CTM Admin</title>
<style>
  * { box-sizing: border-box; margin: 0; padding: 0; }
  body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', sans-serif; background: #f0f2f5; min-height: 100vh; display: flex; align-items: center; justify-content: center; }
  .card { background: white; border-radius: 16px; padding: 40px; width: 460px; box-shadow: 0 4px 24px rgba(0,0,0,0.10); }
  h1 { font-size: 22px; color: #1a1a2e; margin-bottom: 8px; }
  .subtitle { color: #666; font-size: 14px; margin-bottom: 28px; line-height: 1.5; }
  .section-title { font-size: 11px; font-weight: 700; color: #999; text-transform: uppercase; letter-spacing: 0.8px; margin: 20px 0 10px; }
  .toggle-row { display: flex; align-items: center; justify-content: space-between; padding: 14px 0; border-bottom: 1px solid #f5f5f5; }
  .toggle-row:last-child { border-bottom: none; }
  .toggle-label { font-size: 15px; font-weight: 600; color: #222; }
  .toggle-desc { font-size: 12px; color: #aaa; margin-top: 3px; }
  .switch { position: relative; display: inline-block; width: 52px; height: 28px; flex-shrink: 0; }
  .switch input { opacity: 0; width: 0; height: 0; }
  .slider { position: absolute; cursor: pointer; top: 0; left: 0; right: 0; bottom: 0; background: #ddd; transition: .25s; border-radius: 28px; }
  .slider:before { position: absolute; content: ""; height: 20px; width: 20px; left: 4px; bottom: 4px; background: white; transition: .25s; border-radius: 50%; box-shadow: 0 1px 3px rgba(0,0,0,0.2); }
  input:checked + .slider { background: #4f46e5; }
  input:checked + .slider.green { background: #10b981; }
  input:checked + .slider:before { transform: translateX(24px); }
  .divider { height: 1px; background: #f0f0f0; margin: 16px 0; }
  .all-row .toggle-label { color: #4f46e5; font-size: 16px; }
  .all-row .slider { background: #ddd; }
  .all-row input:checked + .slider { background: #4f46e5; }
  .toast { position: fixed; bottom: 24px; right: 24px; background: #1a1a2e; color: white; padding: 12px 20px; border-radius: 10px; font-size: 14px; opacity: 0; transition: opacity .3s; pointer-events: none; }
  .toast.show { opacity: 1; }
  .login-form input { width: 100%; padding: 10px 14px; border: 1px solid #e0e0e0; border-radius: 8px; font-size: 14px; margin-bottom: 10px; outline: none; }
  .login-form input:focus { border-color: #4f46e5; }
  .login-form button { width: 100%; padding: 12px; background: #4f46e5; color: white; border: none; border-radius: 8px; font-size: 15px; cursor: pointer; font-weight: 600; }
  .login-form button:hover { background: #4338ca; }
  #main { display: none; }
  #loginSection { }
  .badge { display: inline-block; font-size: 10px; background: #fef3c7; color: #92400e; padding: 2px 7px; border-radius: 20px; margin-left: 8px; font-weight: 600; vertical-align: middle; }
</style>
</head>
<body>
<div class="card">
  <h1>&#9881;&#65039; Управление ограничениями</h1>
  <p class="subtitle">Включайте или отключайте платёжные ограничения для каждого раздела отдельно</p>

  <div id="loginSection">
    <div class="login-form">
      <input type="text" id="login" placeholder="Логин администратора" />
      <input type="password" id="password" placeholder="Пароль" onkeydown="if(event.key==='Enter')doLogin()" />
      <button onclick="doLogin()">Войти</button>
    </div>
  </div>

  <div id="main">
    <div class="section-title">Глобальное управление</div>
    <div class="toggle-row all-row">
      <div>
        <div class="toggle-label">Всё бесплатно <span class="badge">МАСТЕР</span></div>
        <div class="toggle-desc">Включает/выключает все ограничения сразу</div>
      </div>
      <label class="switch">
        <input type="checkbox" id="toggle-all" onchange="setMode('all', this.checked)">
        <span class="slider"></span>
      </label>
    </div>

    <div class="divider"></div>
    <div class="section-title">По разделам</div>

    <div class="toggle-row">
      <div>
        <div class="toggle-label">&#128230; Грузы</div>
        <div class="toggle-desc">Снять лимит на создание грузовых объявлений</div>
      </div>
      <label class="switch">
        <input type="checkbox" id="toggle-cargo" onchange="setMode('cargo', this.checked)">
        <span class="slider"></span>
      </label>
    </div>

    <div class="toggle-row">
      <div>
        <div class="toggle-label">&#127981; Склады</div>
        <div class="toggle-desc">Снять лимит на создание складов</div>
      </div>
      <label class="switch">
        <input type="checkbox" id="toggle-warehouse" onchange="setMode('warehouse', this.checked)">
        <span class="slider"></span>
      </label>
    </div>

    <div class="toggle-row">
      <div>
        <div class="toggle-label">&#128065; Просмотр контактов</div>
        <div class="toggle-desc">Без списания токенов за просмотр</div>
      </div>
      <label class="switch">
        <input type="checkbox" id="toggle-contacts" onchange="setMode('contacts', this.checked)">
        <span class="slider"></span>
      </label>
    </div>

    <div class="toggle-row">
      <div>
        <div class="toggle-label">&#128666; Перевозчики</div>
        <div class="toggle-desc">Уже бесплатно — флаг для учёта</div>
      </div>
      <label class="switch">
        <input type="checkbox" id="toggle-carriers" onchange="setMode('carriers', this.checked)">
        <span class="slider"></span>
      </label>
    </div>

    <div class="toggle-row">
      <div>
        <div class="toggle-label">&#127970; Компании</div>
        <div class="toggle-desc">Создание компаний без модерации (сразу одобрено)</div>
      </div>
      <label class="switch">
        <input type="checkbox" id="toggle-companies" onchange="setMode('companies', this.checked)">
        <span class="slider"></span>
      </label>
    </div>
  </div>
</div>

<div class="toast" id="toast"></div>

<script>
let token = '';
const base = window.location.origin;

async function doLogin() {
  const login = document.getElementById('login').value.trim();
  const pass = document.getElementById('password').value;
  if (!login || !pass) { showToast('Введите логин и пароль'); return; }
  try {
    const r = await fetch(base + '/api/v1/admin/login', {
      method: 'POST', headers: {'Content-Type':'application/json'},
      body: JSON.stringify({login, password: pass})
    });
    const d = await r.json();
    if (d.success && d.data && d.data.access_token) {
      token = d.data.access_token;
      document.getElementById('loginSection').style.display = 'none';
      document.getElementById('main').style.display = 'block';
      loadStatus();
    } else { showToast('Неверный логин или пароль'); }
  } catch(e) { showToast('Ошибка подключения к серверу'); }
}

async function loadStatus() {
  try {
    const r = await fetch(base + '/api/v1/admin/free-mode', {
      headers: {'Authorization': 'Bearer ' + token}
    });
    const d = await r.json();
    if (d.success && d.data) {
      const f = d.data;
      setCheck('all', f.all);
      setCheck('cargo', f.cargo || f.all);
      setCheck('warehouse', f.warehouse || f.all);
      setCheck('contacts', f.contacts || f.all);
      setCheck('carriers', f.carriers || f.all);
      setCheck('companies', f.companies || false);
    }
  } catch(e) {}
}

function setCheck(key, val) {
  const el = document.getElementById('toggle-' + key);
  if (el) el.checked = !!val;
}

async function setMode(key, enabled) {
  try {
    const r = await fetch(base + '/api/v1/admin/free-mode', {
      method: 'POST',
      headers: {'Content-Type':'application/json','Authorization':'Bearer ' + token},
      body: JSON.stringify({key, enabled})
    });
    const d = await r.json();
    if (d.success) {
      showToast(enabled ? '&#10003; ' + labelFor(key) + ' — включено' : '&#128274; ' + labelFor(key) + ' — выключено');
      if (key === 'all') { loadStatus(); }
    } else { showToast('Ошибка: ' + (d.error && d.error.message || 'неизвестная')); }
  } catch(e) { showToast('Ошибка запроса'); }
}

function labelFor(key) {
  return {'all':'Всё','cargo':'Грузы','warehouse':'Склады','contacts':'Контакты','carriers':'Перевозчики','companies':'Компании'}[key] || key;
}

function showToast(msg) {
  const t = document.getElementById('toast');
  t.textContent = msg;
  t.classList.add('show');
  setTimeout(() => t.classList.remove('show'), 2800);
}
</script>
</body>
</html>`)
}
