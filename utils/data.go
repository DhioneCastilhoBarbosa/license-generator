package utils

import (
	"strings"
	"time"
)

func CalcularValidade(prazo string) time.Time {
	now := time.Now()

	switch strings.ToUpper(prazo) {
	case "1MIN":
		return now.Add(time.Minute) // válido por 1 minuto apenas para teste
	case "1M":
		return now.AddDate(0, 1, 0)
	case "12M":
		return now.AddDate(0, 12, 0)
	case "25M":
		return now.AddDate(0, 24, 0)
	case "36M":
		return now.AddDate(0, 36, 0)
	case "48M":
		return now.AddDate(0, 48, 0)
	case "60M":
		return now.AddDate(0, 60, 0)
	default:
		return now.AddDate(0, 1, 0) // padrão: 1 mês
	}
}
