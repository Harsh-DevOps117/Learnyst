package dtos

type SignupRequest struct {
	FirstName string `json:"first_name" binding:"required,max=100"`
	LastName  string `json:"last_name" binding:"required,max=100"`
	Email     string `json:"email" binding:"required,email,max=255"`
	Password  string `json:"password" binding:"required,min=8,max=72"`
	RoleID    int    `json:"role_id" binding:"required,oneof=1 2 3"`
}
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email,max=255"`
	Password string `json:"password" binding:"required,max=72"`
}
type ResetPasswordRequest struct {
	Email string `json:"email" binding:"required,email,max=255"`
}
type CompleteResetPasswordRequest struct {
	Token       string `json:"token" binding:"required,len=64,hexadecimal"`
	NewPassword string `json:"new_password" binding:"required,min=8,max=72"`
}
type ChangePasswordRequest struct {
	OldPassword string `json:"old_password" binding:"required,max=72"`
	NewPassword string `json:"new_password" binding:"required,min=8,max=72"`
}
