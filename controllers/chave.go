package controllers

import (
	"cve-pro-license-api/database"
	"cve-pro-license-api/models"
	"cve-pro-license-api/utils"
	"net/http"

	"github.com/gin-gonic/gin"
)

// CriarChave cria uma nova chave de acesso e envia por e-mail.
// @Summary Criar chave de acesso
// @Description Gera uma chave de acesso única e envia por e-mail, caso o CPF ainda não esteja cadastrado.
// @Tags Chaves de Acesso
// @Accept json
// @Produce json
// @Param request body models.CriarChaveRequest true "Dados da chave de acesso"
// @Success 201 {object} map[string]interface{} "Chave criada com sucesso"
// @Failure 400 {object} map[string]string "CPF já cadastrado"
// @Failure 500 {object} map[string]string "Erro interno ao processar a chave de acesso ou enviar o e-mail"
// @Router /criar-chave [post]
func CriarChave(c *gin.Context) {
	var req models.ChaveRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Dados inválidos"})
		return
	}

	// Verifica se já existe uma chave com o mesmo CPF
	var existente models.Chave
	if err := database.DB.Where("cpf = ?", req.CPF).First(&existente).Error; err == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Já existe um cadastro com este CPF"})
		return
	}

	// Gerar chave e status
	chaveGerada := utils.GerarChave()
	statusInicial := "Criada"

	chave := models.Chave{
		Nome:   req.Nome,
		Email:  req.Email,
		CPF:    req.CPF,
		Chave:  chaveGerada,
		Status: statusInicial,
	}

	if err := database.DB.Create(&chave).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao criar chave"})
		return
	}

	if err := utils.EnviarEmailChave(req.Email, req.Nome, chaveGerada); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao enviar e-mail", "detalhes": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Chave criada com sucesso",
		"chave":   chave.Chave,
		"status":  chave.Status,
	})
}

// AtualizarStatusChave atualiza o status de uma chave de acesso existente.
// @Summary Atualizar status da chave de acesso
// @Description Atualiza o status de uma chave de acesso existente.
// @Tags Chaves de Acesso
// @Accept json
// @Produce json
// @Param request body models.AtualizarChaveRequest true "Dados da chave de acesso"
// @Success 200 "Status atualizado com sucesso"
// @Failure 400 "Erro nos dados enviados"
// @Failure 404 "Chave não encontrada"
// @Failure 500 "Erro interno ao atualizar status da chave de acesso"
// @Router /atualizar-status-chave [put]
// @Security BearerAuth
func AtualizarStatusChave(c *gin.Context) {
	var req struct {
		Chave  string `json:"chave"`
		Status string `json:"status"`
		Conta  string `json:"conta"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Dados inválidos"})
		return
	}

	var chave models.Chave
	if err := database.DB.Where("chave = ?", req.Chave).First(&chave).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Chave não encontrada"})
		return
	}
	if !models.IsStatusValido(req.Status) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Status inválido"})
		return
	}

	// Bloqueia alteração caso a chave esteja expirada
	if req.Status == models.StatusExpirada {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Não é possível alterar o status de uma chave de acesso expirada"})
		return
	}

	// Bloqueia reativação de uma chave já ativada
	if chave.Status == models.StatusAtivada && req.Status == models.StatusAtivada {
		c.JSON(http.StatusBadRequest, gin.H{"error": "A chave de acesso já está ativada"})
		return
	}

	chave.Status = req.Status
	chave.Conta = req.Conta

	if err := database.DB.Save(&chave).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao atualizar status da chave de acesso"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "Status da chave de acesso atualizado com sucesso",
		"chave":   chave.Chave,
		"status":  chave.Status,
		"conta":   chave.Conta,
	})
}

// ListarChaves lista todas as chaves de acesso, podendo filtrar por email ou CPF.
// @Summary Listar chaves de acesso
// @Description Lista todas as chaves de acesso cadastradas, com opção de filtrar por email ou CPF.
// @Tags Chaves de Acesso
// @Accept json
// @Produce json
// @Param email query string false "Filtrar por email"
// @Param cpf query string false "Filtrar por CPF"
// @Success 200 {array} models.Chave "Lista de chaves de acesso"
// @Failure 500 "Erro interno ao buscar chaves de acesso"
// @Failure 404 "Nenhuma chave de acesso encontrada"
// @Router /chaves [get]
// @Security BearerAuth
func ListarChaves(c *gin.Context) {
	var chaves []models.Chave
	Email := c.Query("email")
	CPF := c.Query("cpf")

	db := database.DB

	if Email != "" {
		db = db.Where("email=?", Email)
	}
	if CPF != "" {
		db = db.Where("cpf=?", CPF)
	}

	if err := db.Find(&chaves).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao buscar chaves de acesso"})
		return
	}

	if len(chaves) == 0 {
		c.JSON(http.StatusNotFound, gin.H{"message": "Nenhuma chave de acesso encontrada"})
		return
	}

	c.JSON(http.StatusOK, chaves)
}

// RecuperarChaves busca chaves de acesso por email e envia a chave gerada por e-mail.
// @Summary Recuperar chaves de acesso
// @Description Busca chaves de acesso por email e envia a chave gerada por e-mail
// @Tags Chaves de Acesso
// @Accept json
// @Produce json
// @Param email query string true "Email do usuário"
// @Success 200 "Chave de acesso enviada com sucesso"
// @Failure 400 "Email inválido"
// @Failure 404 "Nenhuma chave de acesso encontrada"
// @Failure 500 "Erro interno ao enviar e-mail"
// @Router /recuperar-chave [get]
func RecuperarChaves(c *gin.Context) {
	var chaves []models.Chave
	Email := c.Query("email")
	db := database.DB

	if Email != "" {
		db = db.Where("email=?", Email)
	}

	if err := db.Find(&chaves).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao buscar chaves de acesso"})
		return
	}

	if len(chaves) == 0 {
		c.JSON(http.StatusNotFound, gin.H{"message": "Nenhuma chave de acesso encontrada"})
		return
	}

	nome := chaves[0].Nome
	email := chaves[0].Email
	chaveGerada := chaves[0].Chave

	if err := utils.EnviarEmailChave(email, nome, chaveGerada); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao enviar e-mail", "detalhes": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Chave de acesso enviada com sucesso"})

}
