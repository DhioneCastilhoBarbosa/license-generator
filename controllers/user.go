package controllers

import (
	"cve-pro-license-api/database"
	"cve-pro-license-api/models"
	"cve-pro-license-api/utils"
	"errors"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"golang.org/x/crypto/bcrypt"
)

var jwtSecret = []byte(os.Getenv("JWT_SECRET"))

func HashSenha(senha string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(senha), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

func verificarSenha(hash string, senha string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(senha))
	return err == nil
}

func gerarToken(email string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"email": email,
		"exp":   time.Now().Add(time.Hour * 24).Unix(),
	})
	return token.SignedString(jwtSecret)
}

// CadastrarUsuario cadastra um novo usuário.
// @Summary Cadastro de usuário
// @Description Cria um novo usuário aguardando aprovação (sem acesso até o superAdmin definir o nível).
// @Tags Autenticação
// @Accept json
// @Produce json
// @Param request body models.UsuarioRequest true "Dados do usuário"
// @Success 200 {object} map[string]string "Usuário cadastrado com sucesso"
// @Failure 400 {object} map[string]string "Erro nos dados enviados"
// @Router /cadastrar-usuario [post]
func CadastrarUsuario(c *gin.Context) {
	var req models.UsuarioRequest
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"erro": "Dados inválidos"})
		return
	}

	req.Nome = strings.TrimSpace(req.Nome)
	req.Email = strings.TrimSpace(req.Email)
	req.Senha = strings.TrimSpace(req.Senha)

	if req.Email == "" || req.Senha == "" {
		c.JSON(http.StatusBadRequest, gin.H{"erro": "E-mail e senha são obrigatórios"})
		return
	}

	if req.Nome == "" {
		if at := strings.Index(req.Email, "@"); at > 0 {
			req.Nome = req.Email[:at]
		} else {
			req.Nome = req.Email
		}
	}

	hash, err := HashSenha(req.Senha)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"erro": "Erro ao criptografar senha"})
		return
	}

	usuario := models.Usuario{
		Nome:        req.Nome,
		Email:       req.Email,
		Senha:       hash,
		NivelAcesso: models.NivelPendente,
	}

	if err := database.DB.Create(&usuario).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"erro": "Erro ao cadastrar usuário"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"mensagem": "Usuário cadastrado com sucesso. Aguarde aprovação do administrador para acessar o sistema.",
	})
}

// Login autentica um usuário e retorna um token JWT.
// @Summary Login do usuário
// @Description Autentica o usuário e retorna um token JWT para acesso às rotas protegidas.
// @Tags Autenticação
// @Accept json
// @Produce json
// @Param request body object{email=string,senha=string} true "Credenciais do usuário"
// @Success 200 {object} map[string]string "Token JWT"
// @Failure 400 {object} map[string]string "Erro nos dados enviados"
// @Failure 401 {object} map[string]string "Credenciais inválidas"
// @Failure 403 {object} map[string]string "Usuário sem permissão"
// @Router /login [post]
func Login(c *gin.Context) {
	var credenciais struct {
		Email string `json:"email"`
		Senha string `json:"senha"`
	}

	if err := c.ShouldBindJSON(&credenciais); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"erro": "Dados inválidos"})
		return
	}

	var usuario models.Usuario
	database.DB.Where("email = ?", credenciais.Email).First(&usuario)

	if usuario.Email == "" || !verificarSenha(usuario.Senha, credenciais.Senha) {
		c.JSON(http.StatusUnauthorized, gin.H{"erro": "Credenciais inválidas"})
		return
	}

	if usuario.NivelAcesso == models.NivelPendente || !models.IsNivelValido(usuario.NivelAcesso) {
		c.JSON(http.StatusForbidden, gin.H{"erro": "Conta aguardando aprovação do administrador"})
		return
	}

	token, err := gerarToken(usuario.Email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"erro": "Erro ao gerar token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"token":        token,
		"nome":         usuario.Nome,
		"nivel_acesso": usuario.NivelAcesso,
	})
}

// ListarUsuarios retorna todos os usuários cadastrados.
// @Summary Listar usuários
// @Description Lista todos os usuários. Acesso exclusivo de superAdmin.
// @Tags Usuários
// @Produce json
// @Success 200 {array} models.UsuarioResponse
// @Failure 401 {object} map[string]string "Não autorizado"
// @Failure 403 {object} map[string]string "Acesso negado"
// @Security BearerAuth
// @Router /usuarios [get]
func ListarUsuarios(c *gin.Context) {
	var usuarios []models.Usuario
	if err := database.DB.Order("created_at desc").Find(&usuarios).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"erro": "Erro ao listar usuários"})
		return
	}

	resposta := make([]models.UsuarioResponse, len(usuarios))
	for i, u := range usuarios {
		resposta[i] = models.UsuarioParaResponse(u)
	}

	c.JSON(http.StatusOK, resposta)
}

// AtualizarUsuario atualiza nome e nível de acesso de um usuário.
// @Summary Atualizar usuário
// @Description Atualiza nome e nível de acesso. Acesso exclusivo de superAdmin.
// @Tags Usuários
// @Accept json
// @Produce json
// @Param id path int true "ID do usuário"
// @Param request body models.UsuarioUpdateRequest true "Dados para atualização"
// @Success 200 {object} models.UsuarioResponse
// @Failure 400 {object} map[string]string "Erro nos dados enviados"
// @Failure 404 {object} map[string]string "Usuário não encontrado"
// @Security BearerAuth
// @Router /usuarios/{id} [put]
func AtualizarUsuario(c *gin.Context) {
	id := c.Param("id")

	var usuario models.Usuario
	if err := database.DB.First(&usuario, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"erro": "Usuário não encontrado"})
		return
	}

	var req models.UsuarioUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"erro": "Dados inválidos"})
		return
	}

	antes := models.UsuarioParaResponse(usuario)

	if nome := strings.TrimSpace(req.Nome); nome != "" {
		usuario.Nome = nome
	}

	if req.NivelAcesso != "" {
		if req.NivelAcesso == models.NivelPendente || !models.IsNivelValido(req.NivelAcesso) {
			c.JSON(http.StatusBadRequest, gin.H{"erro": "Nível de acesso inválido. Use: superAdmin, admin ou visualizador"})
			return
		}

		if usuario.NivelAcesso == models.NivelSuperAdmin && req.NivelAcesso != models.NivelSuperAdmin {
			if err := garantirOutroSuperAdmin(usuario.ID); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"erro": err.Error()})
				return
			}
		}

		usuario.NivelAcesso = req.NivelAcesso
	}

	if err := database.DB.Save(&usuario).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"erro": "Erro ao atualizar usuário"})
		return
	}

	depois := models.UsuarioParaResponse(usuario)
	entityID := usuario.ID
	utils.SaveAuditLog(database.DB, utils.ActorEmailFromGin(c), utils.AuditActionUpdate, utils.AuditEntityUsuario, &entityID, antes, depois)

	c.JSON(http.StatusOK, depois)
}

// DeletarUsuario remove um usuário do sistema.
// @Summary Deletar usuário
// @Description Remove um usuário. Acesso exclusivo de superAdmin.
// @Tags Usuários
// @Produce json
// @Param id path int true "ID do usuário"
// @Success 200 {object} map[string]string "Usuário removido com sucesso"
// @Failure 400 {object} map[string]string "Operação não permitida"
// @Failure 404 {object} map[string]string "Usuário não encontrado"
// @Security BearerAuth
// @Router /usuarios/{id} [delete]
func DeletarUsuario(c *gin.Context) {
	id := c.Param("id")

	var usuario models.Usuario
	if err := database.DB.First(&usuario, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"erro": "Usuário não encontrado"})
		return
	}

	actorID, _ := c.Get("user_id")
	if actorID == usuario.ID {
		c.JSON(http.StatusBadRequest, gin.H{"erro": "Não é possível excluir o próprio usuário"})
		return
	}

	if usuario.NivelAcesso == models.NivelSuperAdmin {
		if err := garantirOutroSuperAdmin(usuario.ID); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"erro": err.Error()})
			return
		}
	}

	antes := models.UsuarioParaResponse(usuario)
	entityID := usuario.ID

	if err := database.DB.Delete(&usuario).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"erro": "Erro ao deletar usuário"})
		return
	}

	utils.SaveAuditLog(database.DB, utils.ActorEmailFromGin(c), utils.AuditActionDelete, utils.AuditEntityUsuario, &entityID, antes, nil)

	c.JSON(http.StatusOK, gin.H{"mensagem": "Usuário removido com sucesso"})
}

func garantirOutroSuperAdmin(excluirID uint) error {
	var count int64
	if err := database.DB.Model(&models.Usuario{}).
		Where("nivel_acesso = ? AND id <> ?", models.NivelSuperAdmin, excluirID).
		Count(&count).Error; err != nil {
		return err
	}
	if count == 0 {
		return errors.New("deve existir ao menos um superAdmin no sistema")
	}
	return nil
}
