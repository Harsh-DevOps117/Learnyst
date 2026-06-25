package models

import (
	"time"
)

// =========================
// NOTIFICATIONS
// =========================

type Notification struct {
	ID            uint      `gorm:"primaryKey;autoIncrement"`
	TenantID      uint      `gorm:"not null"`
	UserID        uint      `gorm:"not null;index"`
	Type          string    `gorm:"type:varchar(50);not null"`
	Title         string    `gorm:"type:varchar(255);not null"`
	Message       string    `gorm:"type:text"`
	ReferenceType string    `gorm:"type:varchar(50)"`
	ReferenceID   *uint
	IsRead        bool      `gorm:"default:false"`
	CreatedAt     time.Time `gorm:"autoCreateTime"`

	Tenant Tenant `gorm:"foreignKey:TenantID"`
	User   User   `gorm:"foreignKey:UserID"`
}

// =========================
// CHAT
// =========================

type ChatThread struct {
	ID        uint      `gorm:"primaryKey;autoIncrement"`
	TenantID  uint      `gorm:"not null"`
	BatchID   *uint
	Type      string    `gorm:"type:varchar(20);default:'direct'"`
	CreatedAt time.Time `gorm:"autoCreateTime"`

	Tenant Tenant `gorm:"foreignKey:TenantID"`
	Batch  *Batch `gorm:"foreignKey:BatchID"`
}

type ChatParticipant struct {
	ID       uint      `gorm:"primaryKey;autoIncrement"`
	ThreadID uint      `gorm:"not null;uniqueIndex:idx_thread_user"`
	UserID   uint      `gorm:"not null;uniqueIndex:idx_thread_user"`
	JoinedAt time.Time `gorm:"autoCreateTime"`

	Thread ChatThread `gorm:"foreignKey:ThreadID"`
	User   User       `gorm:"foreignKey:UserID"`
}

type ChatMessage struct {
	ID            uint      `gorm:"primaryKey;autoIncrement"`
	ThreadID      uint      `gorm:"not null;index"`
	SenderID      uint      `gorm:"not null"`
	Message       string    `gorm:"type:text"`
	AttachmentUrl string    `gorm:"type:text"`
	IsRead        bool      `gorm:"default:false"`
	SentAt        time.Time `gorm:"autoCreateTime"`

	Thread ChatThread `gorm:"foreignKey:ThreadID"`
	Sender User       `gorm:"foreignKey:SenderID"`
}

// =========================
// AI DOUBT SOLVER
// =========================

type AiDoubtQuery struct {
	ID            uint      `gorm:"primaryKey;autoIncrement"`
	TenantID      uint      `gorm:"not null"`
	StudentID     uint      `gorm:"not null"`
	CourseID      *uint
	ReferenceType string    `gorm:"type:varchar(50)"`
	ReferenceID   *uint
	QuestionText  string    `gorm:"type:text;not null"`
	AiResponse    string    `gorm:"type:text"`
	CreatedAt     time.Time `gorm:"autoCreateTime"`

	Tenant  Tenant  `gorm:"foreignKey:TenantID"`
	Student User    `gorm:"foreignKey:StudentID"`
	Course  *Course `gorm:"foreignKey:CourseID"`
}
