package jobs

import (
	"log"
	"strings"
	"time"

	"cve-pro-license-api/database"
	"cve-pro-license-api/models"
	"cve-pro-license-api/utils"
)

func VerificarLicencasExpiradas() {
	var licencas []models.License

	if err := database.DB.Where("status = ?", "Ativada").Find(&licencas).Error; err != nil {
		log.Println("Erro ao buscar licenças:", err)
		return
	}

	now := time.Now()

	for _, licenca := range licencas {
		// TESTE: expira em 1 minuto
		if strings.ToUpper(licenca.CodigoCompra) == "TESTE" {
			dataExpiracao := licenca.UpdatedAt.Add(1 * time.Minute)
			segundosParaExpirar := int(dataExpiracao.Sub(now).Seconds())

			// Aviso 30 segundos antes
			if segundosParaExpirar <= 30 && !licenca.AvisoExpiracaoEnviado {
				utils.EnviarAvisoRenovacao(licenca.Email, licenca.Nome, licenca.Codigo)
				licenca.AvisoExpiracaoEnviado = true
				database.DB.Save(&licenca)
				log.Printf("Aviso de expiração (TESTE) enviado para %s", licenca.Codigo)
			}

			// Expiração
			if now.After(dataExpiracao) {
				licenca.Status = "Expirada"

				if !licenca.UltimoAvisoRenovacao {
					utils.EnviarAvisoExpiracao(licenca.Email, licenca.Nome, licenca.Codigo)
					licenca.UltimoAvisoRenovacao = true
				}

				if err := database.DB.Save(&licenca).Error; err != nil {
					log.Printf("Erro ao atualizar licença TESTE %s: %v\n", licenca.Codigo, err)
				} else {
					log.Printf("Licença TESTE %s expirada (modo de teste)", licenca.Codigo)
				}
			}
			continue
		}

		// Licenças normais
		dataExpiracao := licenca.UpdatedAt.AddDate(0, licenca.Validade, 0)
		diasParaExpirar := int(dataExpiracao.Sub(now).Hours() / 24)

		// Aviso 3 dias antes
		if diasParaExpirar == 3 && !licenca.AvisoExpiracaoEnviado {
			utils.EnviarAvisoRenovacao(licenca.Email, licenca.Nome, licenca.Codigo)
			licenca.AvisoExpiracaoEnviado = true
			database.DB.Save(&licenca)
			log.Printf("Aviso de expiração enviado para %s", licenca.Codigo)
		}

		// Expirada
		if now.After(dataExpiracao) {
			licenca.Status = "Expirada"

			if !licenca.UltimoAvisoRenovacao {
				utils.EnviarAvisoExpiracao(licenca.Email, licenca.Nome, licenca.Codigo)
				licenca.UltimoAvisoRenovacao = true
			}

			if err := database.DB.Save(&licenca).Error; err != nil {
				log.Printf("Erro ao atualizar licença %s: %v\n", licenca.Codigo, err)
			} else {
				log.Printf("Licença %s expirada", licenca.Codigo)
			}
		}
	}
}
