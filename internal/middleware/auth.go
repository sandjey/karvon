package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"ctm/internal/handler"
	"ctm/internal/repository"
	jwtpkg "ctm/pkg/jwt"
)

// Auth проверяет JWT и кладёт user_id + role в контекст Gin.
func Auth(jwtMgr *jwtpkg.Manager, userRepo *repository.UserRepo) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if !strings.HasPrefix(header, "Bearer ") {
			handler.Unauthorized(c)
			c.Abort()
			return
		}

		tokenStr := strings.TrimPrefix(header, "Bearer ")
		claims, err := jwtMgr.Parse(tokenStr)
		if err != nil {
			handler.Unauthorized(c)
			c.Abort()
			return
		}

		userID, err := uuid.Parse(claims.UserID)
		if err != nil {
			handler.Unauthorized(c)
			c.Abort()
			return
		}

		user, err := userRepo.FindByID(c.Request.Context(), userID)
		if err != nil || user == nil {
			handler.Unauthorized(c)
			c.Abort()
			return
		}

		if user.IsBlocked {
			handler.FailCode(c, 403, "USER_BLOCKED")
			c.Abort()
			return
		}

		if claims.TokenVersion != user.TokenVersion {
			handler.Unauthorized(c)
			c.Abort()
			return
		}

		c.Set("user_id", userID)
		c.Set("role", user.Role)
		c.Set("user", user)
		c.Next()
	}
}

// CompanyVerified проверяет что у пользователя есть хотя бы одна одобренная компания.
func CompanyVerified(companyRepo *repository.CompanyRepo) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.MustGet("user_id").(uuid.UUID)
		companies, err := companyRepo.FindByUserID(c.Request.Context(), userID)
		if err != nil {
			handler.InternalError(c)
			c.Abort()
			return
		}
		for _, comp := range companies {
			if comp.Status == "approved" {
				c.Next()
				return
			}
		}
		handler.FailCode(c, 403, "COMPANY_NOT_VERIFIED")
		c.Abort()
	}
}

// Role проверяет что роль пользователя входит в разрешённый список.
func Role(roles ...string) gin.HandlerFunc {
	allowed := make(map[string]bool, len(roles))
	for _, r := range roles {
		allowed[r] = true
	}
	return func(c *gin.Context) {
		role, _ := c.Get("role")
		if !allowed[role.(string)] {
			handler.Forbidden(c)
			c.Abort()
			return
		}
		c.Next()
	}
}
