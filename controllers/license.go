package controllers

import (
	"cve-pro-license-api/database"
	"cve-pro-license-api/models"
	"cve-pro-license-api/utils"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// CriarLicenca cria uma ou mais licenças e envia por e-mail.
// @Summary Criar licença(s)
// @Description Gera uma ou mais licenças baseadas na compra e envia um único e-mail com os códigos.
// @Tags Licenças
// @Accept json
// @Produce json
// @Param request body models.LicenseRequest true "Dados da licença"
// @Success 201 "Licenças criadas com sucesso"
// @Failure 400 "Erro nos dados enviados"
// @Failure 401 "Não autorizado"
// @Failure 500 "Erro interno ao processar a licença"
// @Security BearerAuth
// @Router /criar-licenca [post]
func CriarLicenca(c *gin.Context) {
	var req models.LicenseRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Dados inválidos"})
		return
	}

	if req.Quantidade <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "A quantidade deve ser maior que 0"})
		return
	}

	var licencas []models.License
	var codigosGerados []string

	for i := 0; i < req.Quantidade; i++ {
		codigo := utils.GerarCodigo(req.Validade)
		statusInicial := models.StatusCriada
		validade := req.Validade

		if req.Coringa {
			codigo = utils.GerarCodigoCoringa()
			statusInicial = models.StatusCoringa
			validade = 0
		}

		if !models.IsStatusValido(statusInicial) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Status inválido"})
			return
		}

		licenca := models.License{
			Nome:         req.Nome,
			Email:        req.Email,
			CodigoCompra: req.CodigoCompra,
			Codigo:       codigo,
			Validade:     validade,
			Status:       statusInicial,
			Quantidade:   req.Quantidade,
			Coringa:      req.Coringa,
		}

		licencas = append(licencas, licenca)
		codigosGerados = append(codigosGerados, codigo)
	}

	if err := database.DB.Create(&licencas).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao salvar licenças"})
		return
	}

	ids := make([]uint, len(licencas))
	for i := range licencas {
		ids[i] = licencas[i].ID
	}
	utils.SaveAuditLog(database.DB, utils.ActorEmailFromGin(c), utils.AuditActionCreate, utils.AuditEntityLicense, nil, nil, map[string]interface{}{
		"codigos":       codigosGerados,
		"license_ids":   ids,
		"quantidade":    req.Quantidade,
		"nome":          req.Nome,
		"email":         req.Email,
		"codigo_compra": req.CodigoCompra,
		"validade":      req.Validade,
		"coringa":       req.Coringa,
	})

	err := utils.EnviarEmail(req.Email, req.Nome, codigosGerados, req.CodigoCompra)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao enviar e-mail", "detalhes": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": fmt.Sprintf("%d licenças criadas e enviadas com sucesso", req.Quantidade),
		"codigos": codigosGerados,
	})
}

// AtualizarStatusLicenca atualiza o status de uma licença existente.
// @Summary Atualizar status da licença
// @Description Atualiza o status de uma licença com base no código da licença.
// @Tags Licenças
// @Accept json
// @Produce json
// @Param request body object{codigo=string,status=string} true "Código da licença e novo status"
// @Success 200 {object} map[string]string "Status atualizado com sucesso"
// @Failure 400 {object} map[string]string "Erro nos dados enviados"
// @Failure 404 {object} map[string]string "Licença não encontrada"
// @Failure 500 {object} map[string]string "Erro interno"
// @Security BearerAuth
// @Router /atualizar-licenca [put]
func AtualizarStatusLicenca(c *gin.Context) {
	var req struct {
		Codigo string `json:"codigo"`
		Status string `json:"status"`
		Teste  bool   `json:"teste"` // opcional
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Dados inválidos"})
		return
	}
	req.Codigo = strings.TrimSpace(req.Codigo)

	var licenca models.License
	if err := database.DB.Where("UPPER(codigo) = UPPER(?)", req.Codigo).First(&licenca).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Licença não encontrada"})
		return
	}

	if !models.IsStatusValido(req.Status) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Status inválido"})
		return
	}

	// Licenca coringa (prefixo P3D) nao pode ter status alterado.
	codigoNormalizado := strings.ToUpper(strings.TrimSpace(licenca.Codigo))
	if licenca.Coringa || licenca.Status == models.StatusCoringa || strings.HasPrefix(codigoNormalizado, "P3D") {
		c.JSON(http.StatusOK, gin.H{
			"message": "Licença coringa validada com sucesso",
			"codigo":  licenca.Codigo,
		})
		return
	}

	// Bloqueia alteração caso a licença esteja expirada
	if licenca.Status == models.StatusExpirada {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Não é possível alterar o status de uma licença expirada"})
		return
	}

	// Bloqueia reativação de uma licença já ativada
	if req.Status == models.StatusAtivada && licenca.Status == models.StatusAtivada {
		c.JSON(http.StatusBadRequest, gin.H{"error": "A licença já está ativada"})
		return
	}

	antes := licenca
	licenca.Status = req.Status

	if req.Status == models.StatusAtivada && req.Teste {
		// Retrocede o UpdatedAt em 2 minutos para simulação de expiração
		licenca.UpdatedAt = time.Now().Add(-2 * time.Minute)
	}

	if err := database.DB.Save(&licenca).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao atualizar licença"})
		return
	}

	id := licenca.ID
	utils.SaveAuditLog(database.DB, utils.ActorEmailFromGin(c), utils.AuditActionUpdate, utils.AuditEntityLicense, &id, antes, licenca)

	c.JSON(http.StatusOK, gin.H{"message": "Status atualizado com sucesso"})
}

// ListarLicencas retorna todas as licenças ou filtra por código da compra ou código da licença.
// @Summary Lista licenças
// @Description Retorna todas as licenças cadastradas ou filtra por código da compra e/ou código da licença.
// @Tags Licenças
// @Produce json
// @Param codigo_compra query string false "Código da compra para filtrar"
// @Param codigo query string false "Código da licença para filtrar"
// @Success 200 {array} models.License
// @Failure 401 {object} map[string]string "Não autorizado"
// @Failure 500 {object} map[string]string "Erro interno"
// @Security BearerAuth
// @Router /licencas [get]
func ListarLicencas(c *gin.Context) {
	var licencas []models.License
	codigoCompra := c.Query("codigo_compra")
	codigoLicenca := c.Query("codigo")

	// Cria uma instância da query
	db := database.DB

	// Adiciona filtros dinamicamente
	if codigoCompra != "" {
		db = db.Where("codigo_compra = ?", codigoCompra)
	}
	if codigoLicenca != "" {
		db = db.Where("codigo = ?", codigoLicenca)
	}

	// Executa a busca
	if err := db.Find(&licencas).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao buscar licenças"})
		return
	}

	c.JSON(http.StatusOK, licencas)
}

// DeletarLicenca remove uma licenca pelo codigo.
// @Summary Deletar licença
// @Description Remove uma licença com base no código informado.
// @Tags Licenças
// @Produce json
// @Param codigo query string true "Código da licença"
// @Success 200 {object} map[string]string "Licença removida com sucesso"
// @Failure 400 {object} map[string]string "Código não informado"
// @Failure 404 {object} map[string]string "Licença não encontrada"
// @Failure 500 {object} map[string]string "Erro interno"
// @Security BearerAuth
// @Router /deletar-licenca [delete]
func DeletarLicenca(c *gin.Context) {
	codigo := strings.TrimSpace(c.Query("codigo"))
	if codigo == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Informe o código da licença"})
		return
	}

	var licenca models.License
	if err := database.DB.Where("UPPER(codigo) = UPPER(?)", codigo).First(&licenca).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Licença não encontrada"})
		return
	}

	antes := licenca
	id := licenca.ID
	if err := database.DB.Delete(&licenca).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao deletar licença"})
		return
	}

	utils.SaveAuditLog(database.DB, utils.ActorEmailFromGin(c), utils.AuditActionDelete, utils.AuditEntityLicense, &id, antes, map[string]interface{}{
		"soft_delete": true,
	})

	c.JSON(http.StatusOK, gin.H{"message": "Licença removida com sucesso"})
}
