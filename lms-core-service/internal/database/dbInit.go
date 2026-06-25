package database

import (
	"log"
	"os"

	"lms-core-service/internal/models"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/plugin/dbresolver"
)

var DB *gorm.DB

func InitDatabase() (*gorm.DB, error) {
	masterUri := os.Getenv("DB_URI")
	if masterUri == "" {
		masterUri = "postgresql://adminFutre123:adminFutreKanye123@localhost:5432/adminFutreMetro123?sslmode=disable"
	}

	replicaUri := os.Getenv("REPLICA_DB_URI")
	if replicaUri == "" {
		replicaUri = "postgresql://adminFutre123:adminFutreKanye123@localhost:5433/adminFutreMetro123?sslmode=disable"
	}

	var err error
	DB, err = gorm.Open(postgres.Open(masterUri), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	return DB, nil
}

func InitDBResolver(db *gorm.DB) {
	masterUri := os.Getenv("DB_URI")
	if masterUri == "" {
		masterUri = "postgresql://adminFutre123:adminFutreKanye123@localhost:5432/adminFutreMetro123?sslmode=disable"
	}
	replicaUri := os.Getenv("REPLICA_DB_URI")
	if replicaUri == "" {
		replicaUri = "postgresql://adminFutre123:adminFutreKanye123@localhost:5433/adminFutreMetro123?sslmode=disable"
	}

	err := db.Use(dbresolver.Register(dbresolver.Config{
		Sources:  []gorm.Dialector{postgres.Open(masterUri)},
		Replicas: []gorm.Dialector{postgres.Open(replicaUri)},
		Policy:   dbresolver.RandomPolicy{},
	}))
	if err != nil {
		log.Println("WARNING: Failed to initialize dbresolver:", err)
	} else {
		log.Println("Master-Slave DB Resolver initialized successfully")
	}
}

func AutoMigrate(db *gorm.DB) error {
	log.Println("Running Automigrations for LMS Core Service...")
	err := db.AutoMigrate(
		// Identity
		&models.User{},
		&models.RefreshToken{},
		&models.Tenant{},
		&models.Role{},
		&models.Permission{},
		&models.RolePermission{},
		&models.TenantStaff{},
		&models.Invitation{},

		// Course
		&models.Course{},
		&models.CourseModule{},
		&models.CourseTeacher{},
		&models.Batch{},
		&models.Enrollment{},
		&models.Coupon{},
		&models.CourseVideo{},
		&models.Class{},
		&models.ClassAttendance{},
		&models.Note{},
		&models.Assignment{},
		&models.AssignmentSubmission{},

		// Quiz
		&models.Quiz{},
		&models.QuizQuestion{},
		&models.QuizAttempt{},
		&models.QuizAnswer{},

		// Communication
		&models.Notification{},
		&models.ChatThread{},
		&models.ChatParticipant{},
		&models.ChatMessage{},
		&models.AiDoubtQuery{},

		// Performance
		&models.Certificate{},
		&models.VideoProgress{},
		&models.TopicPerformance{},
		&models.StudentPerformanceSummary{},
	)

	if err != nil {
		log.Println("Migration failed:", err)
		return err
	}
	
	log.Println("Database Migrated Successfully.")
	return nil
}
