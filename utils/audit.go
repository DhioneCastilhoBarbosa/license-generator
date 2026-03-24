package utils

import (
	"cve-pro-license-api/models"
	"encoding/json"
	"log"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

const (
	AuditActionCreate = "create"
	AuditActionUpdate = "update"
	AuditActionDelete = "delete"

	AuditEntityLicense = "license"
	AuditEntityChave   = "chave"
)

// ActorEmailFromGin retorna o email do JWT (middleware deve popular "user_email").
func ActorEmailFromGin(c *gin.Context) string {
	if v, ok := c.Get("user_email"); ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

func ptrJSON(v interface{}) (*string, error) {
	if v == nil {
		return nil, nil
	}
	b, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}
	s := string(b)
	return &s, nil
}

// SaveAuditLog persiste um registro de auditoria (nao interrompe a operacao principal em caso de erro).
func SaveAuditLog(db *gorm.DB, actorEmail, action, entityType string, entityID *uint, before, after interface{}) {
	beforePtr, err := ptrJSON(before)
	if err != nil {
		log.Printf("audit: marshal before: %v", err)
		beforePtr = nil
	}
	afterPtr, err := ptrJSON(after)
	if err != nil {
		log.Printf("audit: marshal after: %v", err)
		afterPtr = nil
	}

	entry := models.AuditLog{
		ActorEmail:  actorEmail,
		Action:      action,
		EntityType:  entityType,
		EntityID:    entityID,
		BeforeState: beforePtr,
		AfterState:  afterPtr,
	}
	if err := db.Create(&entry).Error; err != nil {
		log.Printf("audit: falha ao gravar log: %v", err)
	}
}
