package models
import (
	"time"
)
type User struct {
	ID              uint      `gorm:"primaryKey;autoIncrement" json:"id"`
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
	CreatedAt       time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt       time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}
type RefreshToken struct {
	ID          uint       `gorm:"primaryKey;autoIncrement"`
	UserID      uint       `gorm:"not null;index"`
	User        User       `gorm:"foreignKey:UserID"`
	TokenHash   string     `gorm:"type:varchar(255);uniqueIndex;not null"`
	DeviceName  string     `gorm:"type:varchar(100)"`
	IPAddress   string     `gorm:"type:varchar(45)"`
	UserAgent   string     `gorm:"type:text"`
	ExpiresAt   time.Time  `gorm:"not null"`
	RevokedAt   *time.Time
	CreatedAt   time.Time  `gorm:"autoCreateTime"`
}
type Tenant struct {
	ID               uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	OwnerUserID      uint      `gorm:"not null" json:"owner_user_id"`
	Owner            User      `gorm:"foreignKey:OwnerUserID" json:"owner"`
	AcademyName      string    `gorm:"type:varchar(255);not null" json:"academy_name"`
	AcademySlug      string    `gorm:"type:varchar(100);uniqueIndex;not null" json:"academy_slug"`
	Description      string    `gorm:"type:text" json:"description"`
	LogoUrl          string    `gorm:"type:text" json:"logo_url"`
	BannerUrl        string    `gorm:"type:text" json:"banner_url"`
	CustomDomain     string    `gorm:"type:varchar(255)" json:"custom_domain"`
	SSLEnabled       bool      `gorm:"default:false" json:"ssl_enabled"`
	SubscriptionPlan string    `gorm:"type:varchar(50)" json:"subscription_plan"`
	RazorpayAccountID string   `gorm:"type:varchar(150)" json:"razorpay_account_id"`
	UpiID            string    `gorm:"type:varchar(150)" json:"upi_id"`
	IsActive         bool      `gorm:"default:true" json:"is_active"`
	IsSuspended      bool      `gorm:"default:false" json:"is_suspended"`
	CreatedAt        time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt        time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}
type Role struct {
	ID          uint      `gorm:"primaryKey;autoIncrement"`
	Name        string    `gorm:"type:varchar(50);uniqueIndex;not null"`
	Description string    `gorm:"type:text"`
	CreatedAt   time.Time `gorm:"autoCreateTime"`
}
type Permission struct {
	ID          uint   `gorm:"primaryKey;autoIncrement"`
	Name        string `gorm:"type:varchar(100);uniqueIndex;not null"`
	Description string `gorm:"type:text"`
}
type RolePermission struct {
	RoleID       uint       `gorm:"primaryKey;autoIncrement:false"`
	PermissionID uint       `gorm:"primaryKey;autoIncrement:false"`
	Role         Role       `gorm:"foreignKey:RoleID"`
	Permission   Permission `gorm:"foreignKey:PermissionID"`
}
type TenantStaff struct {
	ID        uint      `gorm:"primaryKey;autoIncrement"`
	TenantID  uint      `gorm:"not null;uniqueIndex:idx_tenant_user"`
	UserID    uint      `gorm:"not null;uniqueIndex:idx_tenant_user"`
	RoleID    uint      `gorm:"not null"`
	InvitedBy *uint
	IsActive  bool      `gorm:"default:true"`
	UpiID     string    `gorm:"type:varchar(150)"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
	Tenant     Tenant `gorm:"foreignKey:TenantID"`
	User       User   `gorm:"foreignKey:UserID"`
	Role       Role   `gorm:"foreignKey:RoleID"`
	InvitedByU *User  `gorm:"foreignKey:InvitedBy"`
}
type Invitation struct {
	ID          uint      `gorm:"primaryKey;autoIncrement"`
	TenantID    uint      `gorm:"not null"`
	Email       string    `gorm:"type:varchar(255);not null"`
	RoleID      uint      `gorm:"not null"`
	CourseID    *uint
	InviteToken string    `gorm:"type:varchar(255);not null"`
	ExpiresAt   *time.Time
	Accepted    bool      `gorm:"default:false"`
	CreatedAt   time.Time `gorm:"autoCreateTime"`
	Tenant Tenant  `gorm:"foreignKey:TenantID"`
	Role   Role    `gorm:"foreignKey:RoleID"`
	Course *Course `gorm:"foreignKey:CourseID"`
}
