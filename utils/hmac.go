package utils

import (
	"crypto/sha256"
	"encoding/hex"
	"log"
	"os"
	"strings"
)

// Agora usamos apenas SHA256 da secret
func ValidarAssinaturaHMAC(secret string, _ []byte, assinaturaRecebida string) bool {
	useSHA256 := os.Getenv("USE_SHA256_SECRET_ONLY") == "true"

	var expectedHash string
	if useSHA256 {
		hash := sha256.Sum256([]byte(secret))
		expectedHash = hex.EncodeToString(hash[:])
		//log.Println("⚠️ Usando SHA256 puro da secret (modo inseguro)")
	} else {
		// Se não estiver ativado, retorna false
		log.Println("🚫 Modo SHA256 puro da secret não ativado")
		return false
	}

	// Logs para debug
	//log.Println("🔐 SHA256 esperado :", strings.ToLower(expectedHash))
	//log.Println("🔐 SHA256 recebido :", strings.ToLower(assinaturaRecebida))

	return strings.EqualFold(expectedHash, assinaturaRecebida)
}

func GerarAssinaturaHMAC(secret string, _ []byte) string {
	hash := sha256.Sum256([]byte(secret))
	return hex.EncodeToString(hash[:])
}
