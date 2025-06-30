package controllers

import (
	"bytes"
	"cve-pro-license-api/models"
	"cve-pro-license-api/utils"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// Struct recebida no webhook
type VtexOrderEvent struct {
	OrderId      string `json:"orderId"`
	CurrentState string `json:"currentState"`
	LastChange   string `json:"lastChange"`
	SalesChannel string `json:"salesChannel"`
}

// Fila em memória para os OrderIDs recebidos
var orderQueue = make(chan string, 100) // pode ajustar tamanho

var once sync.Once

// Inicia o worker uma única vez
func StartOrderWorker() {
	once.Do(func() {
		go func() {
			for orderID := range orderQueue {
				log.Println("🚚 Processando pedido:", orderID)
				err := fetchOrderDetails(orderID)
				if err != nil {
					log.Println("❌ Erro ao buscar pedido:", err)
				}
				time.Sleep(500 * time.Millisecond) // evita sobrecarga na VTEX
			}
		}()
	})
}

// Função para consumir dados da VTEX
func fetchOrderDetails(orderID string) error {
	appKey := os.Getenv("VTEX_APP_KEY")
	appToken := os.Getenv("VTEX_APP_TOKEN")

	url := fmt.Sprintf("https://lojaintelbras.myvtex.com/api/oms/pvt/orders/%s", orderID)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("X-VTEX-API-AppKey", appKey)
	req.Header.Set("X-VTEX-API-AppToken", appToken)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("erro na requisição: %s", string(body))
	}

	var result struct {
		OrderId           string `json:"orderId"`
		ClientProfileData struct {
			Email     string `json:"email"`
			FirstName string `json:"firstName"`
			LastName  string `json:"lastName"`
		} `json:"clientProfileData"`
		Items []struct {
			ID       string `json:"id"`
			Quantity int    `json:"quantity"`
		} `json:"items"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return err
	}

	log.Printf("📄 Detalhes do pedido %s: %+v\n", orderID, result)

	// Mapeamento de validade por ID
	validadePorID := map[int]int{
		1991061: 12,
		1991062: 24,
		1991063: 36,
		1991064: 48,
		1991065: 60,
		//4820091: 1, // compra teste sensor de presença
	}

	// Verificar se existe ao menos 1 item válido
	for _, item := range result.Items {
		idInt, err := strconv.Atoi(item.ID)
		if err != nil {
			//log.Printf("⚠️ ID do item não é número: %s", item.ID)
			continue
		}

		if validade, ok := validadePorID[idInt]; ok {
			// Dados da licença
			nome := fmt.Sprintf("%s %s", result.ClientProfileData.FirstName, result.ClientProfileData.LastName)
			email := limparEmail(result.ClientProfileData.Email)
			codigoCompra := result.OrderId
			quantidade := item.Quantity

			reqLicenca := models.LicenseRequest{
				Nome:         nome,
				Email:        email,
				CodigoCompra: codigoCompra,
				Quantidade:   quantidade,
				Validade:     validade,
			}

			if err := utils.CriarLicencaAutomatica(reqLicenca); err != nil {
				log.Println("❌ Erro ao criar licença automaticamente:", err)
			} else {
				log.Printf("✅ Licença criada para item ID %d com validade %d meses\n", item.ID, validade)
			}
		}
	}

	return nil
}

// Função para limpar o e-mail recebido do webhook
func limparEmail(email string) string {
	sufixo := ".ct.vtex.com.br"
	if idx := strings.Index(email, sufixo); idx > 0 {
		parte := email[:idx] // até antes de ".ct.vtex.com.br"
		// Remove a parte final do tipo "-hash"
		if dashIdx := strings.LastIndex(parte, "-"); dashIdx > 0 {
			return parte[:dashIdx]
		}
	}
	return email
}

// VtexWebhook godoc
// @Summary Webhook da VTEX para pedidos
// @Description Recebe eventos da VTEX com dados do pedido e inicia o processo de geração de licença automática
// @Tags Webhook
// @Accept json
// @Produce json
// @Param X-VTEX-HMAC-SHA256 header string true "Assinatura HMAC do corpo da requisição"
// @Param payload body VtexOrderEvent true "Evento de pedido VTEX"
// @Success 200 {string} string "ok"
// @Failure 400 {object} models.ErrorResponse "Erro ao ler body ou JSON inválido"
// @Failure 401 {object} models.ErrorResponse "Assinatura inválida"
// @Router /webhook/vtex-vendas [post]
func VtexWebhook(c *gin.Context) {
	secret := os.Getenv("VTEX_WEBHOOK_SECRET")
	bodyBytes, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "erro ao ler body"})
		return
	}

	signature := c.GetHeader("X-VTEX-HMAC-SHA256")
	if !utils.ValidarAssinaturaHMAC(secret, bodyBytes, signature) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "assinatura inválida"})
		return
	}

	c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

	var payload VtexOrderEvent
	if err := c.BindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "JSON inválido"})
		return
	}

	log.Printf("📬 Evento recebido: %+v\n", payload)

	// Enfileira o OrderId
	orderQueue <- payload.OrderId

	c.Status(http.StatusOK)
}
