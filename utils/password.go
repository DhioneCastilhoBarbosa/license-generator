package utils

import (
	"errors"
	"unicode"
)

var ErrSenhaInvalida = errors.New("Mínimo 8 caracteres, com letra maiúscula, minúscula, número e caractere especial")

func ValidarSenha(senha string) error {
	if len(senha) < 8 {
		return ErrSenhaInvalida
	}

	var temMaiuscula, temMinuscula, temNumero, temEspecial bool
	for _, r := range senha {
		switch {
		case unicode.IsUpper(r):
			temMaiuscula = true
		case unicode.IsLower(r):
			temMinuscula = true
		case unicode.IsDigit(r):
			temNumero = true
		case unicode.IsPunct(r) || unicode.IsSymbol(r):
			temEspecial = true
		}
	}

	if temMaiuscula && temMinuscula && temNumero && temEspecial {
		return nil
	}
	return ErrSenhaInvalida
}

func MensagemRequisitosSenha() string {
	return ErrSenhaInvalida.Error()
}
