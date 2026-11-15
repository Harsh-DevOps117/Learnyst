package models

import "time"

type AuthSession struct {
	ID        string    `gorm:"primaryKey;size:64"`
	UserID    uint      `gorm:"not null;index"`
	ExpiresAt time.Time `gorm:"not null;index"`
	CreatedAt time.Time
}
type LocalAsset struct {
	Name         string `gorm:"primaryKey;size:255"`
	UserID       uint   `gorm:"not null;index"`
	CourseID     uint   `gorm:"not null;index"`
	ContentType  string
	OriginalName string
	CreatedAt    time.Time
}

// WorkspacePreference remembers which of the user's active academy memberships is selected.
type WorkspacePreference struct {
	UserID    uint `gorm:"primaryKey;autoIncrement:false"`
	TenantID  uint `gorm:"not null"`
	UpdatedAt time.Time
}
