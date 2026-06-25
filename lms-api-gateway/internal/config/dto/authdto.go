package dtos

type SignupRequest struct {
	FirstName  string `json:"first_name" binding:"required,max=50"`
	MiddleName string `json:"middle_name" binding:"omitempty,max=50"`
	LastName   string `json:"last_name" binding:"required,max=50"`

	Email      string `json:"email" binding:"required,email"`
	Phone      string `json:"phone" binding:"required,e164"`
	Password   string `json:"password" binding:"required,min=8"`

	Location   string `json:"location" binding:"required"`
	RoleID     int    `json:"role_id" binding:"required,gt=0"`
}

type LoginRequest struct {
	Email      string `json:"email" binding:"required,email"`
	Password   string `json:"password" binding:"required"`
	DeviceName string `json:"device_name" binding:"required"`
}

type OTPRequest struct {
	Target  string `json:"target" binding:"required"`
	Channel string `json:"channel" binding:"required,oneof=sms email"`
}

type OTPVerifyRequest struct {
	Target string `json:"target" binding:"required"`
	Code   string `json:"code" binding:"required,len=6,numeric"`
}

type AuthResponse struct {
	Message      string `json:"message"`
	UserID       uint   `json:"user_id,omitempty"`
	RoleID       uint   `json:"role_id,omitempty"`
	AccessToken  string `json:"access_token,omitempty"`
	RefreshToken string `json:"refresh_token,omitempty"`
	ExpiresAt    int64  `json:"expires_at,omitempty"`
}

type CreateTenantRequest struct {
	AcademyName  string `json:"academy_name" binding:"required,max=100"`
	AcademySlug  string `json:"academy_slug" binding:"required,max=100,alphanum"`
	Description  string `json:"description" binding:"omitempty"`
	LogoUrl      string `json:"logo_url" binding:"omitempty,url"`
	BannerUrl    string `json:"banner_url" binding:"omitempty,url"`
	CustomDomain string `json:"custom_domain" binding:"omitempty,fqdn"`
}

type UpdateTenantRequest struct {
	AcademyName  string `json:"academy_name" binding:"omitempty,max=100"`
	Description  string `json:"description" binding:"omitempty"`
	LogoUrl      string `json:"logo_url" binding:"omitempty,url"`
	BannerUrl    string `json:"banner_url" binding:"omitempty,url"`
	IsActive     *bool  `json:"is_active" binding:"omitempty"`
}

type CreateSubscriptionRequest struct {
	TenantID     uint    `json:"tenant_id" binding:"required,gt=0"`
	PlanName     string  `json:"plan_name" binding:"required,max=50"`
	PlanType     string  `json:"plan_type" binding:"required,max=50"`
	PlanStatus   string  `json:"plan_status" binding:"required,oneof=active trial cancelled"`
	PlanCost     float64 `json:"plan_cost" binding:"required,min=0"`
	PlanDuration string  `json:"plan_duration" binding:"required,max=50"`
}

type CreateRoleRequest struct {
	RoleName string `json:"role_name" binding:"required,max=50"`
}

type AddPermissionRequest struct {
	RoleID         int    `json:"role_id" binding:"required,gt=0"`
	PermissionType string `json:"permission_type" binding:"required,max=50"`
	PermissionName string `json:"permission_name" binding:"required,max=50"`
}

type ResetPasswordRequest struct {
	Email string `json:"email" binding:"required,email"`
}

type CompleteResetPasswordRequest struct {
	Token       string `json:"token" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=8"`
}

type ChangePasswordRequest struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=8"`
}

type ChangeEmailRequest struct {
	NewEmail string `json:"new_email" binding:"required,email"`
}

type VerifyChangeEmailRequest struct {
	Token string `json:"token" binding:"required"`
}
