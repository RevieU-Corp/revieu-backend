package model

import "time"

type Report struct {
	ID           int64      `gorm:"primaryKey;autoIncrement" json:"id"`
	ReporterID   int64      `gorm:"not null;index" json:"reporter_id"`
	TargetType   string     `gorm:"type:varchar(20);not null" json:"target_type"`
	TargetID     int64      `gorm:"not null" json:"target_id"`
	Reason       string     `gorm:"type:varchar(50);not null" json:"reason"`
	Description  string     `gorm:"type:text" json:"description"`
	Status       string     `gorm:"type:varchar(20);default:'pending'" json:"status"`
	ReviewedBy   *int64     `json:"reviewed_by"`
	ReviewedAt   *time.Time `json:"reviewed_at"`
	Resolution   string     `gorm:"type:text" json:"resolution"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

func (r *Report) TableName() string { return "reports" }

type AdminAuditLog struct {
	ID         int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	AdminID    int64     `gorm:"not null;index" json:"admin_id"`
	Action     string    `gorm:"type:varchar(50);not null" json:"action"`
	TargetType string    `gorm:"type:varchar(20)" json:"target_type"`
	TargetID   int64     `json:"target_id"`
	Details    string    `gorm:"type:jsonb;default:'{}'" json:"details"`
	CreatedAt  time.Time `json:"created_at"`
}

func (a *AdminAuditLog) TableName() string { return "admin_audit_logs" }
