package models

import (
	"time"
)

type License struct {
	ID                    uint       `json:"id" gorm:"primaryKey"`
	CreatedAt             time.Time  `json:"created_at"`
	UpdatedAt             time.Time  `json:"updated_at"`
	DeletedAt             *time.Time `json:"deleted_at,omitempty" gorm:"index"`
	Nome                  string     `json:"nome"`
	Email                 string     `json:"email"`
	CodigoCompra          string     `json:"codigo_compra"`
	Codigo                string     `json:"codigo" gorm:"unique"`
	Validade              int        `json:"validade"`
	Status                string     `json:"status" gorm:"default:'Criada'"`
	Quantidade            int        `json:"quantidade"` // Novo campo para múltiplas licenças
	UltimoAvisoRenovacao  bool       `gorm:"column:ultimo_aviso_renovacao"`
	AvisoExpiracaoEnviado bool       `gorm:"column:aviso_expiracao_enviado"`
}

type LicenseRequest struct {
	Nome         string `json:"nome" example:"nome do usuário"`
	Email        string `json:"email" example:"email do usuário"`
	CodigoCompra string `json:"codigo_compra" example:"1234abef"`
	Validade     int    `json:"validade" example:"36"`
	Quantidade   int    `json:"quantidade"` // Novo campo para múltiplas licenças
	Teste        bool   `json:"teste"`      // opcional
}

const (
	StatusCriada   = "Criada"
	StatusAtivada  = "Ativada"
	StatusExpirada = "Expirada"
)

func IsStatusValido(status string) bool {
	switch status {
	case StatusCriada, StatusAtivada, StatusExpirada:
		return true
	default:
		return false
	}
}
