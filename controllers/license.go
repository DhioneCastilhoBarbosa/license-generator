package controllers

import (
	"cve-pro-license-api/database"
	"cve-pro-license-api/models"
	"cve-pro-license-api/utils"
	"fmt"
	"net/http"
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

		if !models.IsStatusValido(statusInicial) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Status inválido"})
			return
		}

		licenca := models.License{
			Nome:         req.Nome,
			Email:        req.Email,
			CodigoCompra: req.CodigoCompra,
			Codigo:       codigo,
			Validade:     req.Validade,
			Status:       statusInicial,
			Quantidade:   req.Quantidade,
		}

		licencas = append(licencas, licenca)
		codigosGerados = append(codigosGerados, codigo)
	}

	if err := database.DB.Create(&licencas).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao salvar licenças"})
		return
	}

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

	var licenca models.License
	if err := database.DB.Where("codigo = ?", req.Codigo).First(&licenca).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Licença não encontrada"})
		return
	}

	if !models.IsStatusValido(req.Status) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Status inválido"})
		return
	}

	licenca.Status = req.Status

	if req.Status == models.StatusAtivada && req.Teste {
		// Retrocede o UpdatedAt em 2 minutos para simulação de expiração
		licenca.UpdatedAt = time.Now().Add(-2 * time.Minute)
	}

	if err := database.DB.Save(&licenca).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao atualizar licença"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Status atualizado com sucesso"})
}

// ListarLicencas retorna todas as licenças ou filtra por código da compra.
// @Summary Lista licenças
// @Description Retorna todas as licenças cadastradas ou filtra por código da compra.
// @Tags Licenças
// @Produce json
// @Param codigo_compra query string false "Código da compra para filtrar"
// @Success 200 {array} models.License
// @Failure 401 {object} map[string]string "Não autorizado"
// @Failure 500 {object} map[string]string "Erro interno"
// @Security BearerAuth
// @Router /licencas [get]
func ListarLicencas(c *gin.Context) {
	var licencas []models.License
	codigoCompra := c.Query("codigo_compra") // Pega o parâmetro opcional da query

	query := database.DB

	// Se o parâmetro "codigo_compra" for passado, filtra as licenças
	if codigoCompra != "" {
		query = query.Where("codigo_compra = ?", codigoCompra)
	}

	// Busca as licenças no banco de dados
	if err := query.Find(&licencas).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao buscar licenças"})
		return
	}

	c.JSON(http.StatusOK, licencas)
}
