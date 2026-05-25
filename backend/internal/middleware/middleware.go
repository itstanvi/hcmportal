package middleware

import (
	"net/http"
	"strings"

	"hcm-backend/internal/auth"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const (
	CtxUserID   = "userID"
	CtxTenantID = "tenantID"
	CtxRole     = "role"
	CtxEmail    = "email"
)

// JWTAuth verifies the bearer token and stores claims in the gin context.
func JWTAuth(secret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if header == "" || !strings.HasPrefix(strings.ToLower(header), "bearer ") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing or invalid authorization header"})
			return
		}
		token := strings.TrimSpace(header[len("Bearer "):])
		claims, err := auth.ParseToken(secret, token)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired token"})
			return
		}

		c.Set(CtxUserID, claims.UserID)
		c.Set(CtxTenantID, claims.TenantID)
		c.Set(CtxRole, claims.Role)
		c.Set(CtxEmail, claims.Email)
		c.Next()
	}
}

// RequireRole allows only the specified roles. Use after JWTAuth.
func RequireRole(roles ...string) gin.HandlerFunc {
	allowed := make(map[string]struct{}, len(roles))
	for _, r := range roles {
		allowed[r] = struct{}{}
	}
	return func(c *gin.Context) {
		role, _ := c.Get(CtxRole)
		if _, ok := allowed[role.(string)]; !ok {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "insufficient permissions"})
			return
		}
		c.Next()
	}
}

// TenantID extracts the tenant id from context. Always trust this; never trust the client.
func TenantID(c *gin.Context) uuid.UUID {
	v, _ := c.Get(CtxTenantID)
	if id, ok := v.(uuid.UUID); ok {
		return id
	}
	return uuid.Nil
}

// UserID extracts the user id from context.
func UserID(c *gin.Context) uuid.UUID {
	v, _ := c.Get(CtxUserID)
	if id, ok := v.(uuid.UUID); ok {
		return id
	}
	return uuid.Nil
}

// Role extracts the role from context.
func Role(c *gin.Context) string {
	v, _ := c.Get(CtxRole)
	if r, ok := v.(string); ok {
		return r
	}
	return ""
}
