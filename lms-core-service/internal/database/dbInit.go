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
		log.Fatal("DB_URI environment variable is required")
	}
	replicaUri := os.Getenv("REPLICA_DB_URI")
	if replicaUri == "" {
		log.Println("REPLICA_DB_URI environment variable not set, using master DB_URI")
		replicaUri = masterUri
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
		log.Fatal("DB_URI environment variable is required")
	}
	replicaUri := os.Getenv("REPLICA_DB_URI")
	if replicaUri == "" {
		replicaUri = masterUri
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
		&models.User{},
		&models.RefreshToken{},
		&models.Tenant{},
		&models.Role{},
		&models.Permission{},
		&models.RolePermission{},
		&models.TenantStaff{},
		&models.Invitation{},
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
		&models.Quiz{},
		&models.QuizQuestion{},
		&models.QuizAttempt{},
		&models.QuizAnswer{},
		&models.Notification{},
		&models.ChatThread{},
		&models.ChatParticipant{},
		&models.ChatMessage{},
		&models.AiDoubtQuery{},
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
