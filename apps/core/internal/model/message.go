package model

import "time"

type Message struct {
	ID             int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	ConversationID int64     `gorm:"not null;index" json:"conversation_id"`
	SenderID       int64     `gorm:"not null;index" json:"sender_id"`
	Content        string    `gorm:"type:text;not null" json:"content"`
	MessageType    string    `gorm:"type:varchar(20);default:'text'" json:"message_type"`
	Attachments    string    `gorm:"type:jsonb;default:'[]'" json:"attachments"`
	IsRead         bool      `gorm:"default:false" json:"is_read"`
	CreatedAt      time.Time `json:"created_at"`

	Sender *User `gorm:"foreignKey:SenderID" json:"sender,omitempty"`
}

func (m *Message) TableName() string { return "messages" }
