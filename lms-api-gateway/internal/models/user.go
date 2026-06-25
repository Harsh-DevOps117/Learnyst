package models

import (
	"time"

	"gorm.io/gorm"
)

type User struct {
	ID             uint           `gorm:"primaryKey;autoIncrement"`
	CreatedAt      time.Time
	UpdatedAt      time.Time
	DeletedAt      gorm.DeletedAt `gorm:"index"`
	First_Name     string         `gorm:"type:varchar(50);not null"`
	Middle_Name    string         `gorm:"type:varchar(50)"`
	Last_Name      string         `gorm:"type:varchar(50);not null"`
	Email          string         `gorm:"type:varchar(255);uniqueIndex;not null"`
	Phone          string         `gorm:"type:varchar(20);uniqueIndex;not null"`
	Email_Verified bool           `gorm:"default:false"`
	Phone_Verified bool           `gorm:"default:false"`
	Is_Active      bool           `gorm:"default:true"`
	Location       string         `gorm:"type:text;not null"`

	Password_Hash  string         `gorm:"type:varchar(255);not null"`

	RoleID        int            `gorm:"not null"`
	Role          Role           `gorm:"foreignKey:RoleID"`
	Subscriptions []Subscription `gorm:"foreignKey:OwnerUserID"`
}

type Refresh_Token struct {
	ID          uint           `gorm:"primaryKey;autoIncrement"`
	Token       string         `gorm:"type:text;not null;uniqueIndex"`
	UserID      uint           `gorm:"not null;index"`
	User        User           `gorm:"foreignKey:UserID"`
	Device_Name string         `gorm:"type:varchar(50);not null"`
	IP_Address  string         `gorm:"type:varchar(50);not null"`
	CreatedAt   time.Time      `gorm:"autoCreateTime"`
	ExpiresAt   time.Time      `gorm:"not null"`
	RevokedAt   *time.Time     `gorm:"default:null"`
}

type Role struct {
	ID          int          `gorm:"primaryKey;autoIncrement"`
	Name        string       `gorm:"type:varchar(50);not null;uniqueIndex"`
	Permissions []Permission `gorm:"foreignKey:RoleID"`
}

type Permission struct {
	ID              int    `gorm:"primaryKey;autoIncrement"`
	RoleID          int    `gorm:"not null;index"`
	Permission_Type string `gorm:"type:varchar(50);not null"`
	Permission_Name string `gorm:"type:varchar(50);not null"`
}

type Subscription struct {
	gorm.Model
	OwnerUserID uint      `gorm:"not null;index"`
	TenantID    uint      `gorm:"not null;index"`
	Plan_Name   string    `gorm:"type:varchar(50);not null"`
	Plan_Type   string    `gorm:"type:varchar(50);not null"`
	Plan_Status string    `gorm:"type:varchar(50);not null"`
	Plan_Started time.Time `gorm:"not null"`
	Plan_Ended   time.Time `gorm:"not null"`
	Plan_Cost    float64   `gorm:"type:decimal(10,2);not null"`
	Plan_Duration string   `gorm:"type:varchar(50);not null"`
}

type Tenant struct {
	gorm.Model
	OwnerUserID       uint           `gorm:"not null;index"`
	User              User           `gorm:"foreignKey:OwnerUserID"`
	Academy_Name      string         `gorm:"type:varchar(100);not null"`
	Academy_Slug      string         `gorm:"type:varchar(100);not null;uniqueIndex"`
	Description       string         `gorm:"type:text"`
	Logo_Url          string         `gorm:"type:text"`
	Banner_Url        string         `gorm:"type:text"`
	Custom_Domain     string         `gorm:"type:varchar(100)"`
	Subscription_Plan []Subscription `gorm:"foreignKey:TenantID"`
	Is_Active         bool           `gorm:"default:true"`
	Is_Suspended      bool           `gorm:"default:false"`
}
