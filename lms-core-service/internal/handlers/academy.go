package handlers

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"lms-core-service/internal/database"
	"lms-core-service/internal/mail"
	m "lms-core-service/internal/models"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"
)

func listWorkspaces(c *gin.Context) {
	u := currentUser(c)
	if !staffUser(u) {
		fail(c, 403, "Only academy staff have workspaces.")
		return
	}
	current, err := workspace(u)
	if err != nil {
		fail(c, 500, "Could not load your workspace.")
		return
	}
	var rows []m.TenantStaff
	if database.DB.Preload("Tenant").Where("user_id = ? AND is_active = ?", u.ID, true).Order("id ASC").Find(&rows).Error != nil {
		fail(c, 500, "Could not load your workspaces.")
		return
	}
	result := []gin.H{}
	for _, row := range rows {
		if row.Tenant.IsActive && !row.Tenant.IsSuspended {
			result = append(result, gin.H{"id": row.TenantID, "name": row.Tenant.AcademyName, "slug": row.Tenant.AcademySlug, "role_id": row.RoleID, "current": row.TenantID == current.TenantID})
		}
	}
	c.JSON(200, result)
}
func selectWorkspace(c *gin.Context) {
	u := currentUser(c)
	if !staffUser(u) {
		fail(c, 403, "Staff access required.")
		return
	}
	var req struct {
		TenantID uint `json:"tenant_id" binding:"required"`
	}
	if c.ShouldBindJSON(&req) != nil {
		fail(c, 400, "Choose a valid academy.")
		return
	}
	var staff m.TenantStaff
	if database.DB.Preload("Tenant").Where("user_id = ? AND tenant_id = ? AND is_active = ?", u.ID, req.TenantID, true).First(&staff).Error != nil || !staff.Tenant.IsActive || staff.Tenant.IsSuspended {
		fail(c, 403, "You are not active staff in this academy.")
		return
	}
	pref := m.WorkspacePreference{UserID: u.ID, TenantID: req.TenantID}
	if database.DB.Clauses(clause.OnConflict{Columns: []clause.Column{{Name: "user_id"}}, DoUpdates: clause.AssignmentColumns([]string{"tenant_id", "updated_at"})}).Create(&pref).Error != nil {
		fail(c, 500, "Could not select this academy.")
		return
	}
	c.JSON(200, gin.H{"message": "Workspace selected."})
}
func ownerAcademy(c *gin.Context) (*m.Tenant, bool) {
	u := currentUser(c)
	if u.RoleID != 1 {
		fail(c, 403, "Only the academy owner can manage academy settings and staff.")
		return nil, false
	}
	staff, err := workspace(u)
	if err != nil {
		fail(c, 500, "Could not load academy.")
		return nil, false
	}
	var tenant m.Tenant
	if database.DB.Where("id = ? AND owner_user_id = ?", staff.TenantID, u.ID).First(&tenant).Error != nil {
		fail(c, 403, "You do not own this academy.")
		return nil, false
	}
	return &tenant, true
}
func getAcademy(c *gin.Context) {
	tenant, ok := ownerAcademy(c)
	if !ok {
		return
	}
	var rows []m.TenantStaff
	if database.DB.Preload("User").Where("tenant_id = ? AND role_id = ?", tenant.ID, 2).Order("id DESC").Find(&rows).Error != nil {
		fail(c, 500, "Could not load teachers.")
		return
	}
	teachers := []gin.H{}
	active := 0
	for _, s := range rows {
		var n int64
		database.DB.Model(&m.Course{}).Where("created_by = ? AND tenant_id = ?", s.ID, tenant.ID).Count(&n)
		if s.IsActive {
			active++
		}
		teachers = append(teachers, gin.H{"id": s.ID, "user_id": s.UserID, "name": s.User.FirstName + " " + s.User.LastName, "email": s.User.Email, "active": s.IsActive, "course_count": n})
	}
	var invitations []m.Invitation
	database.DB.Where("tenant_id = ? AND role_id = ?", tenant.ID, 2).Order("id DESC").Find(&invitations)
	invites := []gin.H{}
	for _, i := range invitations {
		invites = append(invites, gin.H{"id": i.ID, "email": i.Email, "accepted": i.Accepted, "expires_at": i.ExpiresAt})
	}
	var courses, published, students int64
	database.DB.Model(&m.Course{}).Where("tenant_id = ?", tenant.ID).Count(&courses)
	database.DB.Model(&m.Course{}).Where("tenant_id = ? AND is_published = ?", tenant.ID, true).Count(&published)
	database.DB.Table("enrollments").Where("tenant_id = ? AND status = ?", tenant.ID, "active").Distinct("student_id").Count(&students)
	c.JSON(200, gin.H{"academy": gin.H{"id": tenant.ID, "name": tenant.AcademyName, "slug": tenant.AcademySlug, "description": tenant.Description}, "teachers": teachers, "invitations": invites, "stats": gin.H{"courses": courses, "published": published, "teachers": active, "students": students}})
}

var academySlug = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)

func updateAcademy(c *gin.Context) {
	tenant, ok := ownerAcademy(c)
	if !ok {
		return
	}
	var req struct {
		Name        string `json:"name" binding:"required,max=255"`
		Slug        string `json:"slug" binding:"required,max=100"`
		Description string `json:"description" binding:"max=5000"`
	}
	if c.ShouldBindJSON(&req) != nil || strings.TrimSpace(req.Name) == "" || !academySlug.MatchString(req.Slug) {
		fail(c, 400, "Enter an academy name and a lowercase address using letters, numbers and hyphens.")
		return
	}
	for _, reserved := range []string{"dashboard", "api", "login", "signup", "forgot-password", "reset-password", "invite", "uploads", "_next"} {
		if req.Slug == reserved {
			fail(c, 400, "This address is reserved. Choose another academy address.")
			return
		}
	}
	err := database.DB.Model(tenant).Updates(map[string]interface{}{"academy_name": strings.TrimSpace(req.Name), "academy_slug": req.Slug, "description": strings.TrimSpace(req.Description)}).Error
	if err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			fail(c, 409, "Another academy is already using that address.")
		} else {
			fail(c, 500, "Could not save academy.")
		}
		return
	}
	getAcademy(c)
}
func inviteAcademyTeacher(c *gin.Context) {
	tenant, ok := ownerAcademy(c)
	if !ok {
		return
	}
	var req struct {
		Email string `json:"email" binding:"required,email,max=255"`
	}
	if c.ShouldBindJSON(&req) != nil {
		fail(c, 400, "Enter a valid teacher email.")
		return
	}
	req.Email = strings.ToLower(strings.TrimSpace(req.Email))
	var existing m.User
	if database.DB.Where("email = ?", req.Email).First(&existing).Error == nil && existing.RoleID != 2 {
		fail(c, 400, "Invite a teacher account. Existing owners and students cannot be changed to teachers through an invitation.")
		return
	}
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		fail(c, 500, "Could not create invitation.")
		return
	}
	token := hex.EncodeToString(b)
	sum := sha256.Sum256([]byte(token))
	expiry := time.Now().Add(7 * 24 * time.Hour)
	invitation := m.Invitation{TenantID: tenant.ID, Email: req.Email, RoleID: 2, InviteToken: hex.EncodeToString(sum[:]), ExpiresAt: &expiry}
	if database.DB.Create(&invitation).Error != nil {
		fail(c, 500, "Could not save invitation.")
		return
	}
	if mail.SendAcademyInvitationEmail(req.Email, tenant.AcademyName, token) != nil {
		database.DB.Delete(&invitation)
		fail(c, http.StatusServiceUnavailable, "Local email inbox is unavailable. Please try again.")
		return
	}
	c.JSON(201, gin.H{"message": "Teacher invitation sent. Check the local inbox."})
}
func acceptAcademyTeacher(c *gin.Context) {
	u := currentUser(c)
	var req struct {
		Token string `json:"token" binding:"required,len=64,hexadecimal"`
	}
	if c.ShouldBindJSON(&req) != nil {
		fail(c, 400, "This invitation link is incomplete.")
		return
	}
	sum := sha256.Sum256([]byte(req.Token))
	var tenantID uint
	err := database.DB.Transaction(func(tx *gorm.DB) error {
		var invite m.Invitation
		if tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("invite_token = ? AND role_id = ? AND accepted = ? AND expires_at > ?", hex.EncodeToString(sum[:]), 2, false, time.Now()).First(&invite).Error != nil {
			return errors.New("This teacher invitation has expired or already been used.")
		}
		if !strings.EqualFold(invite.Email, u.Email) {
			return errors.New("Sign in with the teacher email this invitation was sent to.")
		}
		var tenant m.Tenant
		if tx.Where("id = ? AND is_active = ? AND is_suspended = ?", invite.TenantID, true, false).First(&tenant).Error != nil {
			return errors.New("This academy is unavailable.")
		}
		staff := m.TenantStaff{TenantID: tenant.ID, UserID: u.ID, RoleID: 2, IsActive: true, InvitedBy: &tenant.OwnerUserID}
		if err := tx.Where("tenant_id = ? AND user_id = ?", tenant.ID, u.ID).FirstOrCreate(&staff).Error; err != nil {
			return err
		}
		if err := tx.Model(&staff).Update("is_active", true).Error; err != nil {
			return err
		}
		pref := m.WorkspacePreference{UserID: u.ID, TenantID: tenant.ID}
		if err := tx.Clauses(clause.OnConflict{Columns: []clause.Column{{Name: "user_id"}}, DoUpdates: clause.AssignmentColumns([]string{"tenant_id", "updated_at"})}).Create(&pref).Error; err != nil {
			return err
		}
		tenantID = tenant.ID
		return tx.Model(&invite).Update("accepted", true).Error
	})
	if err != nil {
		fail(c, 400, err.Error())
		return
	}
	c.JSON(200, gin.H{"message": "You have joined the academy as a teacher.", "tenant_id": tenantID, "redirect": "/dashboard"})
}
func updateAcademyTeacher(c *gin.Context) {
	tenant, ok := ownerAcademy(c)
	if !ok {
		return
	}
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil || id == 0 {
		fail(c, 404, "Teacher not found.")
		return
	}
	var req struct {
		Active *bool `json:"active" binding:"required"`
	}
	if c.ShouldBindJSON(&req) != nil {
		fail(c, 400, "Choose active or inactive status.")
		return
	}
	result := database.DB.Model(&m.TenantStaff{}).Where("id = ? AND tenant_id = ? AND role_id = ?", uint(id), tenant.ID, 2).Update("is_active", *req.Active)
	if result.Error != nil {
		fail(c, 500, "Could not update teacher access.")
		return
	}
	if result.RowsAffected == 0 {
		fail(c, 404, "Teacher not found in your academy.")
		return
	}
	c.JSON(200, gin.H{"message": "Teacher access updated."})
}
