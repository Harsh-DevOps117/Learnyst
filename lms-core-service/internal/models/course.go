package models

import (
	"time"
)

// =========================
// COURSE CATALOG & BATCHES
// =========================

type Course struct {
	ID           uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	TenantID     uint      `gorm:"not null;uniqueIndex:idx_tenant_slug" json:"tenant_id"`
	Title        string    `gorm:"type:varchar(255);not null" json:"title"`
	Slug         string    `gorm:"type:varchar(150);not null;uniqueIndex:idx_tenant_slug" json:"slug"`
	Description  string    `gorm:"type:text" json:"description"`
	ThumbnailUrl string    `gorm:"type:text" json:"thumbnail_url"`
	Category     string    `gorm:"type:varchar(100)" json:"category"`
	Level        string    `gorm:"type:varchar(30)" json:"level"`
	Price        float64   `gorm:"type:decimal(10,2);default:0" json:"price"`
	IsPublished  bool      `gorm:"default:false" json:"is_published"`
	CreatedBy    uint      `gorm:"not null" json:"created_by"`
	CreatedAt    time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt    time.Time `gorm:"autoUpdateTime" json:"updated_at"`

	Tenant    Tenant      `gorm:"foreignKey:TenantID" json:"tenant,omitempty"`
	Creator   TenantStaff `gorm:"foreignKey:CreatedBy" json:"creator,omitempty"`
}

type CourseModule struct {
	ID        uint      `gorm:"primaryKey;autoIncrement"`
	CourseID  uint      `gorm:"not null"`
	Title     string    `gorm:"type:varchar(255);not null"`
	Position  int       `gorm:"default:0"`
	CreatedAt time.Time `gorm:"autoCreateTime"`

	Course Course `gorm:"foreignKey:CourseID"`
}

type CourseTeacher struct {
	ID         uint      `gorm:"primaryKey;autoIncrement"`
	TenantID   uint      `gorm:"not null"`
	CourseID   uint      `gorm:"not null;uniqueIndex:idx_course_teacher"`
	TeacherID  uint      `gorm:"not null;uniqueIndex:idx_course_teacher"`
	AssignedBy *uint
	AssignedAt time.Time `gorm:"autoCreateTime"`

	Tenant       Tenant      `gorm:"foreignKey:TenantID"`
	Course       Course      `gorm:"foreignKey:CourseID"`
	Teacher      TenantStaff `gorm:"foreignKey:TeacherID"`
	AssignedByU  *TenantStaff `gorm:"foreignKey:AssignedBy"`
}

type Batch struct {
	ID        uint       `gorm:"primaryKey;autoIncrement"`
	TenantID  uint       `gorm:"not null"`
	CourseID  uint       `gorm:"not null"`
	Name      string     `gorm:"type:varchar(150);not null"`
	StartDate *time.Time `gorm:"type:date"`
	EndDate   *time.Time `gorm:"type:date"`
	IsActive  bool       `gorm:"default:true"`
	CreatedAt time.Time  `gorm:"autoCreateTime"`

	Tenant Tenant `gorm:"foreignKey:TenantID"`
	Course Course `gorm:"foreignKey:CourseID"`
}

type Enrollment struct {
	ID         uint      `gorm:"primaryKey;autoIncrement"`
	TenantID   uint      `gorm:"not null"`
	BatchID    uint      `gorm:"not null;uniqueIndex:idx_batch_student"`
	StudentID  uint      `gorm:"not null;uniqueIndex:idx_batch_student"`
	Status     string    `gorm:"type:varchar(20);default:'active'"`
	EnrolledAt time.Time `gorm:"autoCreateTime"`

	Tenant  Tenant `gorm:"foreignKey:TenantID"`
	Batch   Batch  `gorm:"foreignKey:BatchID"`
	Student User   `gorm:"foreignKey:StudentID"`
}

type Coupon struct {
	ID                uint      `gorm:"primaryKey;autoIncrement"`
	TenantID          uint      `gorm:"not null"`
	CourseID          *uint     // Null means it applies to all courses for this tenant
	Code              string    `gorm:"type:varchar(50);not null"`
	DiscountPercent   int       `gorm:"default:0"`
	MaxDiscountAmount float64   `gorm:"type:decimal(10,2);default:0"`
	ValidUntil        *time.Time
	IsActive          bool      `gorm:"default:true"`
	CreatedAt         time.Time `gorm:"autoCreateTime"`

	Tenant Tenant  `gorm:"foreignKey:TenantID"`
	Course *Course `gorm:"foreignKey:CourseID"`
}

// =========================
// VIDEO CONTENT
// =========================

type CourseVideo struct {
	ID              uint      `gorm:"primaryKey;autoIncrement"`
	TenantID        uint      `gorm:"not null"`
	CourseID        uint      `gorm:"not null"`
	ModuleID        *uint
	TeacherID       uint      `gorm:"not null"`
	Title           string    `gorm:"type:varchar(255);not null"`
	Description     string    `gorm:"type:text"`
	VideoUrl        string    `gorm:"type:text;not null"`
	DurationSeconds int
	Position        int       `gorm:"default:0"`
	IsPublished     bool      `gorm:"default:false"`
	CreatedAt       time.Time `gorm:"autoCreateTime"`
	UpdatedAt       time.Time `gorm:"autoUpdateTime"`

	Tenant  Tenant        `gorm:"foreignKey:TenantID"`
	Course  Course        `gorm:"foreignKey:CourseID"`
	Module  *CourseModule `gorm:"foreignKey:ModuleID"`
	Teacher TenantStaff   `gorm:"foreignKey:TeacherID"`
}

// =========================
// CLASS SCHEDULE & ATTENDANCE
// =========================

type Class struct {
	ID              uint      `gorm:"primaryKey;autoIncrement"`
	TenantID        uint      `gorm:"not null"`
	BatchID         uint      `gorm:"not null"`
	TeacherID       uint      `gorm:"not null"`
	VideoID         *uint
	Title           string    `gorm:"type:varchar(255);not null"`
	ScheduledAt     time.Time `gorm:"not null"`
	DurationMinutes int
	Status          string    `gorm:"type:varchar(20);default:'upcoming'"`
	CreatedAt       time.Time `gorm:"autoCreateTime"`

	Tenant  Tenant       `gorm:"foreignKey:TenantID"`
	Batch   Batch        `gorm:"foreignKey:BatchID"`
	Teacher TenantStaff  `gorm:"foreignKey:TeacherID"`
	Video   *CourseVideo `gorm:"foreignKey:VideoID"`
}

type ClassAttendance struct {
	ID        uint      `gorm:"primaryKey;autoIncrement"`
	ClassID   uint      `gorm:"not null;uniqueIndex:idx_class_student"`
	StudentID uint      `gorm:"not null;uniqueIndex:idx_class_student"`
	Status    string    `gorm:"type:varchar(20);default:'absent'"`
	MarkedBy  *uint
	MarkedAt  time.Time `gorm:"autoCreateTime"`

	Class    Class        `gorm:"foreignKey:ClassID"`
	Student  User         `gorm:"foreignKey:StudentID"`
	Marker   *TenantStaff `gorm:"foreignKey:MarkedBy"`
}

// =========================
// NOTES & ASSIGNMENTS
// =========================

type Note struct {
	ID         uint      `gorm:"primaryKey;autoIncrement"`
	TenantID   uint      `gorm:"not null"`
	CourseID   uint      `gorm:"not null"`
	ModuleID   *uint
	TeacherID  uint      `gorm:"not null"`
	Title      string    `gorm:"type:varchar(255);not null"`
	FileUrl    string    `gorm:"type:text;not null"`
	CreatedAt  time.Time `gorm:"autoCreateTime"`

	Tenant  Tenant        `gorm:"foreignKey:TenantID"`
	Course  Course        `gorm:"foreignKey:CourseID"`
	Module  *CourseModule `gorm:"foreignKey:ModuleID"`
	Teacher TenantStaff   `gorm:"foreignKey:TeacherID"`
}

type Assignment struct {
	ID            uint       `gorm:"primaryKey;autoIncrement"`
	TenantID      uint       `gorm:"not null"`
	CourseID      uint       `gorm:"not null"`
	ModuleID      *uint
	TeacherID     uint       `gorm:"not null"`
	Title         string     `gorm:"type:varchar(255);not null"`
	Description   string     `gorm:"type:text"`
	AttachmentUrl string     `gorm:"type:text"`
	DueDate       *time.Time
	MaxMarks      int        `gorm:"default:100"`
	CreatedAt     time.Time  `gorm:"autoCreateTime"`

	Tenant  Tenant        `gorm:"foreignKey:TenantID"`
	Course  Course        `gorm:"foreignKey:CourseID"`
	Module  *CourseModule `gorm:"foreignKey:ModuleID"`
	Teacher TenantStaff   `gorm:"foreignKey:TeacherID"`
}

type AssignmentSubmission struct {
	ID            uint       `gorm:"primaryKey;autoIncrement"`
	AssignmentID  uint       `gorm:"not null;uniqueIndex:idx_assignment_student"`
	StudentID     uint       `gorm:"not null;uniqueIndex:idx_assignment_student"`
	SubmissionUrl string     `gorm:"type:text"`
	SubmissionTxt string     `gorm:"type:text"`
	SubmittedAt   *time.Time
	MarksObtained *int
	Feedback      string     `gorm:"type:text"`
	GradedBy      *uint
	GradedAt      *time.Time
	Status        string     `gorm:"type:varchar(20);default:'pending'"`

	Assignment Assignment   `gorm:"foreignKey:AssignmentID"`
	Student    User         `gorm:"foreignKey:StudentID"`
	Grader     *TenantStaff `gorm:"foreignKey:GradedBy"`
}
