package middleware

import (
	"cve-pro-license-api/database"
	"cve-pro-license-api/models"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
)

var jwtSecret = []byte(os.Getenv("JWT_SECRET"))

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString := c.GetHeader("Authorization")

		if tokenString == "" || !strings.HasPrefix(tokenString, "Bearer ") {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Token ausente"})
			c.Abort()
			return
		}

		tokenString = strings.TrimPrefix(tokenString, "Bearer ")

		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return jwtSecret, nil
		})

		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Token inválido"})
			c.Abort()
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Token inválido"})
			c.Abort()
			return
		}

		email, ok := claims["email"].(string)
		if !ok || email == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Token inválido"})
			c.Abort()
			return
		}

		var usuario models.Usuario
		if err := database.DB.Where("email = ?", email).First(&usuario).Error; err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Usuário não encontrado"})
			c.Abort()
			return
		}

		if usuario.NivelAcesso == models.NivelPendente || !models.IsNivelValido(usuario.NivelAcesso) {
			c.JSON(http.StatusForbidden, gin.H{"error": "Conta aguardando aprovação do administrador"})
			c.Abort()
			return
		}

		c.Set("user_email", usuario.Email)
		c.Set("user_id", usuario.ID)
		c.Set("user_nome", usuario.Nome)
		c.Set("user_nivel", usuario.NivelAcesso)

		c.Next()
	}
}
