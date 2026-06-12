package controllers

import (
	"cve-pro-license-api/database"
	"cve-pro-license-api/models"
	"cve-pro-license-api/utils"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"errors"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

const (
	msgRecuperacaoEnviada = "Se o e-mail estiver cadastrado, você receberá instruções para redefinir sua senha."
	intervaloMinSolicitacao = 2 * time.Minute
)

func duracaoTokenRecuperacao() time.Duration {
	minutos := 60
	if v := os.Getenv("PASSWORD_RESET_EXPIRY_MINUTES"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			minutos = n
		}
	}
	return time.Duration(minutos) * time.Minute
}

func urlRecuperacaoSenha() string {
	if url := strings.TrimSpace(os.Getenv("PASSWORD_RESET_URL")); url != "" {
		return strings.TrimRight(url, "/")
	}
	return "https://license.intelbras-cve-pro.com.br/auth/reset-password"
}

func gerarTokenRecuperacao() (string, string, error) {
	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		return "", "", err
	}
	token := hex.EncodeToString(raw)
	hash := sha256.Sum256([]byte(token))
	return token, hex.EncodeToString(hash[:]), nil
}

func hashTokenRecuperacao(token string) string {
	hash := sha256.Sum256([]byte(strings.TrimSpace(token)))
	return hex.EncodeToString(hash[:])
}

func tokensCoincidem(hashArmazenado, hashRecebido string) bool {
	return subtle.ConstantTimeCompare([]byte(hashArmazenado), []byte(hashRecebido)) == 1
}

func invalidarTokensPendentes(usuarioID uint) error {
	agora := time.Now()
	return database.DB.Model(&models.PasswordResetToken{}).
		Where("usuario_id = ? AND used_at IS NULL AND expires_at > ?", usuarioID, agora).
		Update("used_at", agora).Error
}

func ultimaSolicitacaoRecente(usuarioID uint) bool {
	var ultimo models.PasswordResetToken
	err := database.DB.
		Where("usuario_id = ?", usuarioID).
		Order("created_at desc").
		First(&ultimo).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return false
	}
	return time.Since(ultimo.CreatedAt) < intervaloMinSolicitacao
}

// SolicitarRecuperacaoSenha envia link de redefinição por e-mail (se a conta existir).
// @Summary Solicitar recuperação de senha
// @Description Envia e-mail com link seguro de redefinição. Resposta genérica para evitar enumeração de contas.
// @Tags Autenticação
// @Accept json
// @Produce json
// @Param request body object{email=string} true "E-mail da conta"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Router /solicitar-recuperacao-senha [post]
func SolicitarRecuperacaoSenha(c *gin.Context) {
	var req struct {
		Email string `json:"email" form:"email"`
	}
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"erro": "Dados inválidos"})
		return
	}

	email := normalizarEmail(req.Email)
	if email == "" {
		c.JSON(http.StatusBadRequest, gin.H{"erro": "E-mail é obrigatório"})
		return
	}

	var usuario models.Usuario
	err := database.DB.Where("LOWER(email) = ?", email).First(&usuario).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		c.JSON(http.StatusOK, gin.H{"mensagem": msgRecuperacaoEnviada})
		return
	}
	if err != nil {
		log.Printf("erro ao buscar usuário para recuperação de senha: %v", err)
		c.JSON(http.StatusOK, gin.H{"mensagem": msgRecuperacaoEnviada})
		return
	}

	if ultimaSolicitacaoRecente(usuario.ID) {
		c.JSON(http.StatusOK, gin.H{"mensagem": msgRecuperacaoEnviada})
		return
	}

	token, tokenHash, err := gerarTokenRecuperacao()
	if err != nil {
		log.Printf("erro ao gerar token de recuperação: %v", err)
		c.JSON(http.StatusOK, gin.H{"mensagem": msgRecuperacaoEnviada})
		return
	}

	if err := invalidarTokensPendentes(usuario.ID); err != nil {
		log.Printf("erro ao invalidar tokens pendentes: %v", err)
	}

	expiresAt := time.Now().Add(duracaoTokenRecuperacao())
	resetToken := models.PasswordResetToken{
		UsuarioID: usuario.ID,
		TokenHash: tokenHash,
		ExpiresAt: expiresAt,
	}
	if err := database.DB.Create(&resetToken).Error; err != nil {
		log.Printf("erro ao salvar token de recuperação: %v", err)
		c.JSON(http.StatusOK, gin.H{"mensagem": msgRecuperacaoEnviada})
		return
	}

	link := urlRecuperacaoSenha() + "?token=" + token
	if err := utils.EnviarEmailRecuperacaoSenha(usuario.Email, usuario.Nome, link, int(duracaoTokenRecuperacao().Minutes())); err != nil {
		log.Printf("erro ao enviar e-mail de recuperação (%s): %v", usuario.Email, err)
	}

	c.JSON(http.StatusOK, gin.H{"mensagem": msgRecuperacaoEnviada})
}

// RedefinirSenha redefine a senha usando token de uso único.
// @Summary Redefinir senha
// @Description Valida token enviado por e-mail e define nova senha.
// @Tags Autenticação
// @Accept json
// @Produce json
// @Param request body object{token=string,senha=string} true "Token e nova senha"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Router /redefinir-senha [post]
func RedefinirSenha(c *gin.Context) {
	var req struct {
		Token string `json:"token" form:"token"`
		Senha string `json:"senha" form:"senha"`
	}
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"erro": "Dados inválidos"})
		return
	}

	token := strings.TrimSpace(req.Token)
	senha := strings.TrimSpace(req.Senha)
	if token == "" || senha == "" {
		c.JSON(http.StatusBadRequest, gin.H{"erro": "Token e senha são obrigatórios"})
		return
	}

	if err := utils.ValidarSenha(senha); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"erro": err.Error()})
		return
	}

	tokenHash := hashTokenRecuperacao(token)
	agora := time.Now()

	var resetToken models.PasswordResetToken
	err := database.DB.
		Where("token_hash = ? AND used_at IS NULL AND expires_at > ?", tokenHash, agora).
		First(&resetToken).Error
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"erro": "Token inválido ou expirado"})
		return
	}

	if !tokensCoincidem(resetToken.TokenHash, tokenHash) {
		c.JSON(http.StatusBadRequest, gin.H{"erro": "Token inválido ou expirado"})
		return
	}

	var usuario models.Usuario
	if err := database.DB.First(&usuario, resetToken.UsuarioID).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"erro": "Token inválido ou expirado"})
		return
	}

	hash, err := HashSenha(senha)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"erro": "Erro ao criptografar senha"})
		return
	}

	tx := database.DB.Begin()
	if err := tx.Model(&resetToken).Update("used_at", agora).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"erro": "Erro ao redefinir senha"})
		return
	}
	if err := tx.Model(&usuario).Update("senha", hash).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"erro": "Erro ao redefinir senha"})
		return
	}
	if err := tx.Commit().Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"erro": "Erro ao redefinir senha"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"mensagem": "Senha redefinida com sucesso"})
}
