package models

import (
	"time"
)

// =========================
// CERTIFICATES
// =========================

type Certificate struct {
	ID              uint      `gorm:"primaryKey;autoIncrement"`
	TenantID        uint      `gorm:"not null"`
	StudentID       uint      `gorm:"not null;uniqueIndex:idx_cert_student_course"`
	CourseID        uint      `gorm:"not null;uniqueIndex:idx_cert_student_course"`
	CertificateCode string    `gorm:"type:varchar(50);uniqueIndex;not null"`
	CertificateUrl  string    `gorm:"type:text"`
	IssuedAt        time.Time `gorm:"autoCreateTime"`

	Tenant  Tenant `gorm:"foreignKey:TenantID"`
	Student User   `gorm:"foreignKey:StudentID"`
	Course  Course `gorm:"foreignKey:CourseID"`
}

// =========================
// PROGRESS ANALYSIS
// =========================

type VideoProgress struct {
	ID                 uint       `gorm:"primaryKey;autoIncrement"`
	StudentID          uint       `gorm:"not null;uniqueIndex:idx_prog_student_video"`
	VideoID            uint       `gorm:"not null;uniqueIndex:idx_prog_student_video"`
	WatchTimeSeconds   int        `gorm:"default:0"`
	CompletionPercent  int        `gorm:"default:0"`
	LastWatchedAt      *time.Time

	Student User        `gorm:"foreignKey:StudentID"`
	Video   CourseVideo `gorm:"foreignKey:VideoID"`
}

type TopicPerformance struct {
	ID               uint      `gorm:"primaryKey;autoIncrement"`
	StudentID        uint      `gorm:"not null"`
	CourseID         uint      `gorm:"not null"`
	TopicName        string    `gorm:"type:varchar(150);not null"`
	AccuracyPercent  float64   `gorm:"type:decimal(5,2);default:0"`
	AttemptsCount    int       `gorm:"default:0"`
	LastUpdated      time.Time `gorm:"autoUpdateTime"`

	Student User   `gorm:"foreignKey:StudentID"`
	Course  Course `gorm:"foreignKey:CourseID"`
}

// =========================
// GRADES / PERFORMANCE SUMMARY
// =========================

type StudentPerformanceSummary struct {
	ID                 uint      `gorm:"primaryKey;autoIncrement"`
	StudentID          uint      `gorm:"not null;uniqueIndex:idx_sum_student_course"`
	CourseID           uint      `gorm:"not null;uniqueIndex:idx_sum_student_course"`
	AssignmentAvg      float64   `gorm:"type:decimal(5,2);default:0"`
	QuizAvg            float64   `gorm:"type:decimal(5,2);default:0"`
	AttendancePercent  float64   `gorm:"type:decimal(5,2);default:0"`
	OverallGrade       string    `gorm:"type:varchar(10)"`
	UpdatedAt          time.Time `gorm:"autoUpdateTime"`

	Student User   `gorm:"foreignKey:StudentID"`
	Course  Course `gorm:"foreignKey:CourseID"`
}
