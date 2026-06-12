package utils

import "testing"

func TestValidarSenha(t *testing.T) {
	validas := []string{"Senha@123", "Abcdef1!", "Intelbras#9"}
	for _, s := range validas {
		if err := ValidarSenha(s); err != nil {
			t.Fatalf("senha válida rejeitada (%q): %v", s, err)
		}
	}

	invalidas := []string{"", "curta1!", "semnumero!", "SEMNMIN1!", "SemEspecial1", "SoLetras", "12345678"}
	for _, s := range invalidas {
		if err := ValidarSenha(s); err == nil {
			t.Fatalf("senha inválida aceita (%q)", s)
		}
	}
}
