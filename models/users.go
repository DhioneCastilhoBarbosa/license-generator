package models

import "gorm.io/gorm"

// @name Usuario
// @ignore
type Usuario struct {
	gorm.Model
	Email        string `json:"email" gorm:"unique"`
	Senha        string `json:"senha"`
	TemPermissao bool   `json:"-" gorm:"default:false"`
}

// UsuarioRequest representa os dados enviados para cadastrar um usuário
// @name UsuarioRequest
type UsuarioRequest struct {
	Email string `json:"email" example:"teste@exemplo.com"`
	Senha string `json:"senha" example:"123456"`
}
