package models

import "gorm.io/gorm"

const (
	NivelSuperAdmin   = "superAdmin"
	NivelAdmin        = "admin"
	NivelVisualizador = "visualizador"
	NivelPendente     = "pendente"
)

func IsNivelValido(nivel string) bool {
	switch nivel {
	case NivelSuperAdmin, NivelAdmin, NivelVisualizador:
		return true
	default:
		return false
	}
}

func PodeEscrever(nivel string) bool {
	return nivel == NivelSuperAdmin || nivel == NivelAdmin
}

// @name Usuario
// @ignore
type Usuario struct {
	gorm.Model
	Nome        string `json:"nome"`
	Email       string `json:"email" gorm:"unique"`
	Senha       string `json:"senha"`
	NivelAcesso string `json:"nivel_acesso" gorm:"default:pendente"`
}

// UsuarioRequest representa os dados enviados para cadastrar um usuário.
// @name UsuarioRequest
type UsuarioRequest struct {
	Nome  string `json:"nome" form:"nome" example:"João Silva"`
	Email string `json:"email" form:"email" example:"teste@exemplo.com"`
	Senha string `json:"senha" form:"senha" example:"123456"`
}

// UsuarioUpdateRequest representa os dados para atualizar um usuário.
// @name UsuarioUpdateRequest
type UsuarioUpdateRequest struct {
	Nome        string `json:"nome" example:"João Silva"`
	NivelAcesso string `json:"nivel_acesso" example:"admin"`
}

// UsuarioResponse representa um usuário retornado pela API (sem senha).
// @name UsuarioResponse
type UsuarioResponse struct {
	ID          uint   `json:"id"`
	Nome        string `json:"nome"`
	Email       string `json:"email"`
	NivelAcesso string `json:"nivel_acesso"`
	CreatedAt   string `json:"created_at"`
}

func UsuarioParaResponse(u Usuario) UsuarioResponse {
	return UsuarioResponse{
		ID:          u.ID,
		Nome:        u.Nome,
		Email:       u.Email,
		NivelAcesso: u.NivelAcesso,
		CreatedAt:   u.CreatedAt.Format("2006-01-02 15:04:05"),
	}
}
