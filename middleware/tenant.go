package middleware

import (
	"net/http"
	"github.com/gin-gonic/gin"
)

func TenantMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		tenantID := c.GetHeader("X-Tenant-ID")
		if tenantID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "El header X-Tenant-ID es obligatorio"})
			c.Abort()
			return
		}
		// Guardo tenant_id en el contexto para usarlo en los controladores
		c.Set("tenant_id", tenantID)
		c.Next()
	}
}