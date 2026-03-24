package utils

import (
	"cve-pro-license-api/database"
	"cve-pro-license-api/models"
	"fmt"
)

func CriarLicencaAutomatica(req models.LicenseRequest) error {
	var licencas []models.License
	var codigosGerados []string

	for i := 0; i < req.Quantidade; i++ {
		codigo := GerarCodigo(req.Validade)
		status := models.StatusCriada
		validade := req.Validade
		if req.Coringa {
			codigo = GerarCodigoCoringa()
			status = models.StatusCoringa
			validade = 0
		}

		licenca := models.License{
			Nome:         req.Nome,
			Email:        req.Email,
			CodigoCompra: req.CodigoCompra,
			Codigo:       codigo,
			Validade:     validade,
			Status:       status,
			Quantidade:   req.Quantidade,
			Coringa:      req.Coringa,
		}

		licencas = append(licencas, licenca)
		codigosGerados = append(codigosGerados, codigo)
	}

	if err := database.DB.Create(&licencas).Error; err != nil {
		return fmt.Errorf("erro ao salvar licenças: %w", err)
	}

	if err := EnviarEmail(req.Email, req.Nome, codigosGerados, req.CodigoCompra); err != nil {
		return fmt.Errorf("erro ao enviar email: %w", err)
	}

	return nil
}
