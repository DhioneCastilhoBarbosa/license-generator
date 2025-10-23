package models

import "time"

type Chave struct {
	ID        uint       `json:"id" gorm:"primaryKey"`
	Nome      string     `json:"nome"`
	Email     string     `json:"email"`
	CPF       string     `json:"cpf" gorm:"unique;not null"`
	Chave     string     `json:"chave" gorm:"unique"`
	Status    string     `json:"status" gorm:"default:'Criada'"`
	Conta     string     `json:"conta" gorm:"default:'Conta não informada'"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at,omitempty" gorm:"index"`
}


type ChaveRequest struct {
	Nome   string `json:"nome" example:"nome do usuário"`
	Email  string `json:"email" example:"email do usuário"`
	CPF    string `json:"cpf" example:"12345678901"`
	Chave  string `json:"chave" example:"chave-1234"`
	Status string `json:"status" example:"Criada"`
	Conta  string `json:"conta" example:"Conta de exemplo"`
}

type AtualizarChaveRequest struct {
	Chave  string `json:"chave" example:"chave-1234"`
	Status string `json:"status" example:"Ativada"`
	Conta  string `json:"conta" example:"Conta de exemplo"`
}

type CriarChaveRequest struct {
	Nome  string `json:"nome" example:"nome do usuário"`
	Email string `json:"email" example:"email do usuário"`
	CPF   string `json:"cpf" example:"12345678901"`
}
