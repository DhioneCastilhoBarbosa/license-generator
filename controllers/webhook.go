package controllers

import (
	"bytes"
	"cve-pro-license-api/utils"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

type VtexOrderEvent struct {
	OrderId      string `json:"orderId"`
	CurrentState string `json:"currentState"`
	LastChange   string `json:"lastChange"`
	SalesChannel string `json:"salesChannel"`
}

func VtexWebhook(c *gin.Context) {
	secret := os.Getenv("VTEX_WEBHOOK_SECRET")
	log.Println("🔑 Segredo do webhook:", secret)

	// ✅ 1. Lê o corpo bruto antes de tudo
	bodyBytes, err := io.ReadAll(c.Request.Body)
	if err != nil {
		log.Println("❌ Erro ao ler body:", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Erro ao ler body"})
		return
	}

	log.Println("📦 Corpo recebido bruto:", string(bodyBytes))

	// ✅ 2. Valida assinatura
	signature := c.GetHeader("X-VTEX-HMAC-SHA256")
	if !utils.ValidarAssinaturaHMAC(secret, bodyBytes, signature) {
		log.Println("🔐 HMAC esperado :", utils.GerarAssinaturaHMAC(secret, bodyBytes))
		log.Println("🔐 HMAC recebido :", signature)
		log.Println("🚫 Assinatura HMAC inválida", signature)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "assinatura inválida"})
		return
	}

	// ✅ 3. Reconstrói o corpo para uso com BindJSON
	c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

	// ✅ 4. Parse para struct
	var payload VtexOrderEvent
	if err := c.BindJSON(&payload); err != nil {
		log.Println("❌ JSON inválido:", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "JSON inválido"})
		return
	}

	log.Printf("📬 Evento VTEX recebido: %+v\n", payload)
	c.Status(http.StatusOK)
}
