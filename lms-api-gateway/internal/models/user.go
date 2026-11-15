package models

import "time"

// User mirrors the core service's users table. Never serialize password hashes.
type User struct {
	ID              uint      `gorm:"primaryKey" json:"id"`
	FirstName       string    `gorm:"type:varchar(100);not null" json:"first_name"`
	MiddleName      string    `gorm:"type:varchar(100)" json:"middle_name"`
	LastName        string    `gorm:"type:varchar(100);not null" json:"last_name"`
	Email           string    `gorm:"type:varchar(255);uniqueIndex;not null" json:"email"`
	Phone           string    `gorm:"type:varchar(20)" json:"phone"`
	PasswordHash    string    `gorm:"type:varchar(255);not null" json:"-"`
	ProfileImageUrl string    `gorm:"type:text" json:"profile_image_url"`
	EmailVerified   bool      `gorm:"default:false" json:"email_verified"`
	PhoneVerified   bool      `gorm:"default:false" json:"phone_verified"`
	IsActive        bool      `gorm:"default:true" json:"is_active"`
	Location        string    `gorm:"type:varchar(255)" json:"location"`
	RoleID          int       `gorm:"not null" json:"role_id"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// Only SHA-256 hashes of random, 256-bit session/reset tokens are stored.
type AuthSession struct {
	ID        string    `gorm:"primaryKey;size:64"`
	UserID    uint      `gorm:"not null;index"`
	ExpiresAt time.Time `gorm:"not null;index"`
	CreatedAt time.Time
}
type PasswordReset struct {
	ID        string    `gorm:"primaryKey;size:64"`
	UserID    uint      `gorm:"not null;index"`
	ExpiresAt time.Time `gorm:"not null;index"`
	CreatedAt time.Time
}
