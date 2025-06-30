package controllers

import (
	"cve-pro-license-api/database"
	"cve-pro-license-api/models"
	"net/http"
	"os"
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
// @Description Cria um novo usuário no banco de dados.
// @Tags Autenticação
// @Accept json
// @Produce json
// @Param request body models.UsuarioRequest true "Dados do usuário"
// @Success 200 {object} map[string]string "Usuário cadastrado com sucesso"
// @Failure 400 {object} map[string]string "Erro nos dados enviados"
// @Router /cadastrar-usuario [post]
func CadastrarUsuario(c *gin.Context) {
	var usuario models.Usuario
	if err := c.ShouldBindJSON(&usuario); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"erro": "Dados inválidos"})
		return
	}

	// Define permissão padrão
	usuario.TemPermissao = false

	// Hash da senha
	hash, err := HashSenha(usuario.Senha)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"erro": "Erro ao criptografar senha"})
		return
	}
	usuario.Senha = hash

	// Salva no banco
	if err := database.DB.Create(&usuario).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"erro": "Erro ao cadastrar usuário"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"mensagem": "Usuário cadastrado com sucesso"})
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

	// Verifica se usuário existe e senha está correta
	if usuario.Email == "" || !verificarSenha(usuario.Senha, credenciais.Senha) {
		c.JSON(http.StatusUnauthorized, gin.H{"erro": "Credenciais inválidas"})
		return
	}

	// Verifica permissão
	if !usuario.TemPermissao {
		c.JSON(http.StatusForbidden, gin.H{"erro": "Usuário sem permissão de acesso"})
		return
	}

	// Gera token JWT
	token, err := gerarToken(usuario.Email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"erro": "Erro ao gerar token"})
		return
	}

	//println("Token gerado:", token)

	c.JSON(http.StatusOK, gin.H{"token": token})
}
