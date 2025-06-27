package utils

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"log"
	"os"
	"strings"
)

func ValidarAssinaturaHMAC(secret string, corpo []byte, assinaturaRecebida string) bool {
	usePayload := os.Getenv("USE_HMAC_WITH_PAYLOAD") != "false"

	var mac []byte
	if usePayload {
		mac = gerarMAC([]byte(secret), corpo)
	} else {
		// Ignora o corpo, usa string vazia
		mac = gerarMAC([]byte(secret), []byte(""))
		log.Println("⚠️ Ignorando corpo do payload para HMAC (modo teste)")
	}

	expectedMAC := hex.EncodeToString(mac)

	// Logs para debug
	log.Println("🔐 HMAC esperado :", strings.ToLower(expectedMAC))
	log.Println("🔐 HMAC recebido :", strings.ToLower(assinaturaRecebida))

	return hmac.Equal(
		[]byte(strings.ToLower(expectedMAC)),
		[]byte(strings.ToLower(assinaturaRecebida)),
	)
}

func GerarAssinaturaHMAC(secret string, corpo []byte) string {
	usePayload := os.Getenv("USE_HMAC_WITH_PAYLOAD") != "false"

	if usePayload {
		return hex.EncodeToString(gerarMAC([]byte(secret), corpo))
	}

	log.Println("⚠️ Ignorando corpo do payload para gerar HMAC (modo teste)")
	return hex.EncodeToString(gerarMAC([]byte(secret), []byte("")))
}

func gerarMAC(secret, data []byte) []byte {
	mac := hmac.New(sha256.New, secret)
	mac.Write(data)
	return mac.Sum(nil)
}
