package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/russo2642/renti_kz/internal/domain"
	"github.com/russo2642/renti_kz/internal/services"
	"github.com/russo2642/renti_kz/internal/utils"
	"github.com/russo2642/renti_kz/pkg/auth"
)

type Middleware struct {
	tokenManager     auth.TokenManager
	authUseCase      domain.AuthUseCase
	userCacheService *services.UserCacheService
}

func NewMiddleware(tokenManager auth.TokenManager, authUseCase domain.AuthUseCase, userCacheService *services.UserCacheService) *Middleware {
	return &Middleware{
		tokenManager:     tokenManager,
		authUseCase:      authUseCase,
		userCacheService: userCacheService,
	}
}

func (m *Middleware) AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		accessToken, ok := utils.ExtractBearerToken(c)
		if !ok {
			c.Abort()
			return
		}

		if m.userCacheService != nil {
			if cachedUser, err := m.userCacheService.GetCachedTokenValidation(accessToken); err == nil && cachedUser != nil {
				if !cachedUser.IsActive {
					utils.AbortWithUnauthorized(c, "аккаунт заблокирован")
					return
				}
				utils.SetUserContext(c, cachedUser.ID, cachedUser.Role)
				c.Next()
				return
			}
		}

		_, err := m.tokenManager.ParseAccessToken(accessToken)
		if err != nil {
			utils.AbortWithUnauthorized(c, "недействительный токен")
			return
		}

		userID, err := m.authUseCase.GetUserFromToken(accessToken)
		if err != nil {
			utils.AbortWithUnauthorized(c, "недействительный токен")
			return
		}

		if !userID.IsActive {
			utils.AbortWithUnauthorized(c, "аккаунт заблокирован")
			return
		}

		if m.userCacheService != nil {
			_ = m.userCacheService.CacheTokenValidation(accessToken, userID)
		}

		utils.SetUserContext(c, userID.ID, userID.Role)
		c.Next()
	}
}

func (m *Middleware) OptionalAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		accessToken, ok := utils.TryExtractBearerToken(c)
		if !ok {
			c.Next()
			return
		}

		if m.userCacheService != nil {
			if cachedUser, err := m.userCacheService.GetCachedTokenValidation(accessToken); err == nil && cachedUser != nil {
				if !cachedUser.IsActive {
					c.Next()
					return
				}
				utils.SetUserContext(c, cachedUser.ID, cachedUser.Role)
				c.Next()
				return
			}
		}

		_, err := m.tokenManager.ParseAccessToken(accessToken)
		if err != nil {
			c.Next()
			return
		}

		userID, err := m.authUseCase.GetUserFromToken(accessToken)
		if err != nil {
			c.Next()
			return
		}

		if !userID.IsActive {
			c.Next()
			return
		}

		if m.userCacheService != nil {
			_ = m.userCacheService.CacheTokenValidation(accessToken, userID)
		}

		utils.SetUserContext(c, userID.ID, userID.Role)
		c.Next()
	}
}

func (m *Middleware) RoleMiddleware(roles ...domain.UserRole) gin.HandlerFunc {
	return func(c *gin.Context) {

		if utils.CheckUserRole(c, roles...) {
			c.Next()
		} else {
			c.Abort()
		}
	}
}

func (m *Middleware) CORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		utils.SetCORSHeaders(c)

		if c.Request.Method == "OPTIONS" {
			c.Header("Access-Control-Max-Age", "86400")
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}
