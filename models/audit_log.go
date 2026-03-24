package models

import "time"

// AuditLog registra criacoes, atualizacoes e exclusoes (soft delete) feitas pela API.
type AuditLog struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	CreatedAt   time.Time `json:"created_at"`
	ActorEmail  string    `json:"actor_email" gorm:"index"`
	Action      string    `json:"action" gorm:"index"`      // create, update, delete
	EntityType  string    `json:"entity_type" gorm:"index"` // license, chave
	EntityID    *uint     `json:"entity_id,omitempty"`
	BeforeState *string   `json:"before_state,omitempty" gorm:"type:jsonb"`
	AfterState  *string   `json:"after_state,omitempty" gorm:"type:jsonb"`
}
