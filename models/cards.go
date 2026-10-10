package models

import (
	"time"
	"gorm.io/gorm"
)

type Card struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	TenantID  string         `gorm:"index;not null" json:"tenant_id"` // Clave para multi-tenant isolation
	CardholderName string    `gorm:"size:255;not null" json:"cardholder_name"`
	Last4     string         `gorm:"size:4;not null" json:"last4"`     // Solo los últimos 4 dígitos (PCI Compliance)
	PanHash   string         `gorm:"size:64;not null;index" json:"-"`  // Hash seguro para búsquedas sin exponer el PAN
	Status    string         `gorm:"size:50;default:'ISSUED'" json:"status"` // ISSUED, ACTIVE, SUSPENDED, CLOSED
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}