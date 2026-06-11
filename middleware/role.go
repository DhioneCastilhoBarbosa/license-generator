package middleware

import (
	"cve-pro-license-api/models"
	"net/http"

	"github.com/gin-gonic/gin"
)

func RequireNivel(niveis ...string) gin.HandlerFunc {
	allowed := make(map[string]struct{}, len(niveis))
	for _, n := range niveis {
		allowed[n] = struct{}{}
	}

	return func(c *gin.Context) {
		nivel, ok := c.Get("user_nivel")
		if !ok {
			c.JSON(http.StatusForbidden, gin.H{"error": "Acesso negado"})
			c.Abort()
			return
		}

		n, ok := nivel.(string)
		if !ok || !models.IsNivelValido(n) {
			c.JSON(http.StatusForbidden, gin.H{"error": "Acesso negado"})
			c.Abort()
			return
		}

		if _, ok := allowed[n]; !ok {
			c.JSON(http.StatusForbidden, gin.H{"error": "Acesso negado"})
			c.Abort()
			return
		}

		c.Next()
	}
}

func RequireEscrita() gin.HandlerFunc {
	return RequireNivel(models.NivelAdmin, models.NivelSuperAdmin)
}
