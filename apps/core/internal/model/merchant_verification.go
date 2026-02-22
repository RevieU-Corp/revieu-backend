package model

import "time"

type MerchantVerification struct {
	ID               int64      `gorm:"primaryKey;autoIncrement" json:"id"`
	MerchantID       int64      `gorm:"not null;index" json:"merchant_id"`
	DocumentType     string     `gorm:"type:varchar(50);not null" json:"document_type"`
	DocumentURL      string     `gorm:"type:varchar(512);not null" json:"document_url"`
	BusinessLicense  string     `gorm:"type:varchar(255)" json:"business_license"`
	Status           string     `gorm:"type:varchar(20);default:'pending'" json:"status"`
	ReviewedBy       *int64     `json:"reviewed_by"`
	ReviewedAt       *time.Time `json:"reviewed_at"`
	RejectionReason  string     `gorm:"type:text" json:"rejection_reason"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`

	Merchant *Merchant `gorm:"foreignKey:MerchantID" json:"merchant,omitempty"`
}

func (mv *MerchantVerification) TableName() string { return "merchant_verifications" }
