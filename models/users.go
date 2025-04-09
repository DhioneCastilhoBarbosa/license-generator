package models

import "gorm.io/gorm"

type Usuario struct {
	gorm.Model
	Email        string `json: "email" gorm:"unique"`
	Senha        string `json: "senha"`
	TemPermissao bool   `json: "-" gorm:"default:false"`
}
