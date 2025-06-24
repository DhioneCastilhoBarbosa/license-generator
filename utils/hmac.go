package utils

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"log"
	"strings"
)

func ValidarAssinaturaHMAC(secret string, corpo []byte, assinaturaRecebida string) bool {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(corpo)
	expectedMAC := hex.EncodeToString(mac.Sum(nil))

	// Logs para debug
	log.Println("🔐 HMAC esperado :", strings.ToLower(expectedMAC))
	log.Println("🔐 HMAC recebido :", strings.ToLower(assinaturaRecebida))

	// Normaliza para lowercase antes de comparar
	return hmac.Equal(
		[]byte(strings.ToLower(expectedMAC)),
		[]byte(strings.ToLower(assinaturaRecebida)),
	)
}

func GerarAssinaturaHMAC(secret string, corpo []byte) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(corpo)
	return hex.EncodeToString(mac.Sum(nil))
}
