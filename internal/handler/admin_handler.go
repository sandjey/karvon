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
	enabled := false
	for _, p := range list {
		if p.Key == "system:free_mode" {
			enabled = p.TokensAmount > 0
			break
		}
	}
	OK(c, gin.H{"enabled": enabled})
}

func (h *AdminHandler) SetFreeMode(c *gin.Context) {
	var req struct {
		Enabled bool `json:"enabled"`
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
	if err := h.svc.UpdatePricing(c.Request.Context(), "system:free_mode", map[string]interface{}{"tokens_amount": val}, adminID); err != nil {
		InternalError(c)
		return
	}
	key := "FREE_MODE_DISABLED"
	if req.Enabled {
		key = "FREE_MODE_ENABLED"
	}
	OKMsg(c, gin.H{"enabled": req.Enabled}, key)
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
  .card { background: white; border-radius: 16px; padding: 40px; width: 420px; box-shadow: 0 4px 24px rgba(0,0,0,0.10); }
  h1 { font-size: 22px; color: #1a1a2e; margin-bottom: 8px; }
  .subtitle { color: #666; font-size: 14px; margin-bottom: 32px; line-height: 1.5; }
  .toggle-row { display: flex; align-items: center; justify-content: space-between; padding: 20px 0; border-top: 1px solid #f0f0f0; }
  .toggle-label { font-size: 16px; font-weight: 600; color: #333; }
  .toggle-desc { font-size: 13px; color: #999; margin-top: 4px; }
  .switch { position: relative; display: inline-block; width: 56px; height: 30px; }
  .switch input { opacity: 0; width: 0; height: 0; }
  .slider { position: absolute; cursor: pointer; top: 0; left: 0; right: 0; bottom: 0; background: #ccc; transition: .3s; border-radius: 30px; }
  .slider:before { position: absolute; content: ""; height: 22px; width: 22px; left: 4px; bottom: 4px; background: white; transition: .3s; border-radius: 50%; }
  input:checked + .slider { background: #4f46e5; }
  input:checked + .slider:before { transform: translateX(26px); }
  .status { margin-top: 24px; padding: 14px 18px; border-radius: 10px; font-size: 14px; font-weight: 500; display: none; }
  .status.on { background: #ecfdf5; color: #065f46; display: block; }
  .status.off { background: #fef2f2; color: #991b1b; display: block; }
  .login-form { margin-bottom: 24px; }
  .login-form input { width: 100%; padding: 10px 14px; border: 1px solid #ddd; border-radius: 8px; font-size: 14px; margin-bottom: 10px; }
  .login-form button { width: 100%; padding: 11px; background: #4f46e5; color: white; border: none; border-radius: 8px; font-size: 15px; cursor: pointer; font-weight: 600; }
  .login-form button:hover { background: #4338ca; }
  .features { background: #f8f9ff; border-radius: 10px; padding: 16px; margin-bottom: 24px; }
  .features p { font-size: 13px; color: #555; margin-bottom: 6px; }
  .features p:last-child { margin-bottom: 0; }
  .features span { color: #4f46e5; font-weight: 600; }
  #main { display: none; }
  #loginSection { display: block; }
</style>
</head>
<body>
<div class="card">
  <h1>Управление режимом</h1>
  <p class="subtitle">Переключатель бесплатного режима для платформы CTM</p>

  <div id="loginSection">
    <div class="login-form">
      <input type="text" id="login" placeholder="Логин администратора" />
      <input type="password" id="password" placeholder="Пароль" />
      <button onclick="doLogin()">Войти</button>
    </div>
  </div>

  <div id="main">
    <div class="features">
      <p>+ <span>Создание грузов</span> — без ограничений</p>
      <p>+ <span>Создание складов</span> — без ограничений</p>
      <p>+ <span>Создание перевозчиков</span> — без ограничений</p>
      <p>+ <span>Просмотр контактов</span> — без списания токенов</p>
    </div>

    <div class="toggle-row">
      <div>
        <div class="toggle-label">Бесплатный режим</div>
        <div class="toggle-desc">Отключает все платёжные ограничения</div>
      </div>
      <label class="switch">
        <input type="checkbox" id="freeToggle" onchange="setFreeMode(this.checked)">
        <span class="slider"></span>
      </label>
    </div>

    <div id="status" class="status"></div>
  </div>
</div>

<script>
let token = '';
const base = window.location.origin;

async function doLogin() {
  const login = document.getElementById('login').value;
  const pass = document.getElementById('password').value;
  try {
    const r = await fetch(base + '/api/v1/admin/login', {
      method: 'POST',
      headers: {'Content-Type':'application/json'},
      body: JSON.stringify({login, password: pass})
    });
    const d = await r.json();
    if (d.success && d.data && d.data.access_token) {
      token = d.data.access_token;
      document.getElementById('loginSection').style.display = 'none';
      document.getElementById('main').style.display = 'block';
      loadStatus();
    } else {
      alert('Неверный логин или пароль');
    }
  } catch(e) { alert('Ошибка подключения'); }
}

async function loadStatus() {
  const r = await fetch(base + '/api/v1/admin/free-mode', {
    headers: {'Authorization': 'Bearer ' + token}
  });
  const d = await r.json();
  if (d.success) {
    document.getElementById('freeToggle').checked = d.data.enabled;
    showStatus(d.data.enabled);
  }
}

async function setFreeMode(enabled) {
  const r = await fetch(base + '/api/v1/admin/free-mode', {
    method: 'POST',
    headers: {'Content-Type':'application/json','Authorization':'Bearer ' + token},
    body: JSON.stringify({enabled})
  });
  const d = await r.json();
  if (d.success) {
    showStatus(enabled);
  }
}

function showStatus(enabled) {
  const el = document.getElementById('status');
  if (enabled) {
    el.className = 'status on';
    el.textContent = 'Бесплатный режим ВКЛЮЧЁН — все ограничения сняты';
  } else {
    el.className = 'status off';
    el.textContent = 'Бесплатный режим ВЫКЛЮЧЕН — платёжные ограничения активны';
  }
}
</script>
</body>
</html>`)
}
