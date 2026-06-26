package models
import (
	"time"
)
type Quiz struct {
	ID              uint      `gorm:"primaryKey;autoIncrement"`
	TenantID        uint      `gorm:"not null"`
	CourseID        uint      `gorm:"not null"`
	ModuleID        *uint
	TeacherID       uint      `gorm:"not null"`
	Title           string    `gorm:"type:varchar(255);not null"`
	Description     string    `gorm:"type:text"`
	TotalMarks      int       `gorm:"default:0"`
	DurationMinutes int
	IsAiGenerated   bool      `gorm:"default:false"`
	AiPrompt        string    `gorm:"type:text"`
	CreatedAt       time.Time `gorm:"autoCreateTime"`
	Tenant  Tenant        `gorm:"foreignKey:TenantID"`
	Course  Course        `gorm:"foreignKey:CourseID"`
	Module  *CourseModule `gorm:"foreignKey:ModuleID"`
	Teacher TenantStaff   `gorm:"foreignKey:TeacherID"`
}
type QuizQuestion struct {
	ID            uint   `gorm:"primaryKey;autoIncrement"`
	QuizID        uint   `gorm:"not null"`
	QuestionText  string `gorm:"type:text;not null"`
	QuestionType  string `gorm:"type:varchar(20);default:'mcq'"`
	Options       string `gorm:"type:json"`
	CorrectAnswer string `gorm:"type:text"`
	Marks         int    `gorm:"default:1"`
	Position      int    `gorm:"default:0"`
	IsAiGenerated bool   `gorm:"default:false"`
	Quiz Quiz `gorm:"foreignKey:QuizID"`
}
type QuizAttempt struct {
	ID          uint       `gorm:"primaryKey;autoIncrement"`
	QuizID      uint       `gorm:"not null"`
	StudentID   uint       `gorm:"not null"`
	StartedAt   time.Time  `gorm:"autoCreateTime"`
	SubmittedAt *time.Time
	Score       *int
	Status      string     `gorm:"type:varchar(20);default:'in_progress'"`
	Quiz    Quiz `gorm:"foreignKey:QuizID"`
	Student User `gorm:"foreignKey:StudentID"`
}
type QuizAnswer struct {
	ID            uint   `gorm:"primaryKey;autoIncrement"`
	AttemptID     uint   `gorm:"not null;uniqueIndex:idx_attempt_question"`
	QuestionID    uint   `gorm:"not null;uniqueIndex:idx_attempt_question"`
	AnswerText    string `gorm:"type:text"`
	IsCorrect     *bool
	MarksObtained int    `gorm:"default:0"`
	Attempt  QuizAttempt  `gorm:"foreignKey:AttemptID"`
	Question QuizQuestion `gorm:"foreignKey:QuestionID"`
}
