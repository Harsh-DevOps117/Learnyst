package database

import (
	"fmt"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"lms-core-service/internal/models"
	"lms-core-service/internal/storage"
	"os"
	"path/filepath"
	"time"
)

func SeedLocal(db *gorm.DB) error {
	for _, r := range []models.Role{{ID: 1, Name: "Tenant"}, {ID: 2, Name: "Teacher"}, {ID: 3, Name: "Student"}, {ID: 4, Name: "SuperAdmin"}} {
		role := r
		if err := db.Where("id = ?", role.ID).FirstOrCreate(&role).Error; err != nil {
			return err
		}
	}
	if os.Getenv("ENV") != "development" || os.Getenv("SEED_DEMO") != "true" {
		return nil
	}
	return db.Transaction(func(tx *gorm.DB) error {
		users := []struct {
			First, Last, Email, Password string
			Role                         int
		}{
			{"Alex", "Morgan", "teacher@learnyst.local", "LocalTeacher123!", 2},
			{"Jamie", "Singh", "student@learnyst.local", "LocalStudent123!", 3},
		}
		saved := []models.User{}
		for _, u := range users {
			var user models.User
			err := tx.Where("email = ?", u.Email).First(&user).Error
			if err == gorm.ErrRecordNotFound {
				hash, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
				if err != nil {
					return err
				}
				user = models.User{FirstName: u.First, LastName: u.Last, Email: u.Email, PasswordHash: string(hash), RoleID: u.Role, IsActive: true, Location: "Bengaluru, India"}
				if err = tx.Create(&user).Error; err != nil {
					return err
				}
			} else if err != nil {
				return err
			}
			saved = append(saved, user)
		}
		teacher, student := saved[0], saved[1]
		tenant := models.Tenant{OwnerUserID: teacher.ID, AcademyName: "The Learning Studio", AcademySlug: "learning-studio", Description: "A little curiosity. A lot of possibility. Practical courses to help you build what comes next."}
		if err := tx.Where("academy_slug = ?", tenant.AcademySlug).FirstOrCreate(&tenant).Error; err != nil {
			return err
		}
		staff := models.TenantStaff{TenantID: tenant.ID, UserID: teacher.ID, RoleID: 2, IsActive: true}
		if err := tx.Where("tenant_id = ? AND user_id = ?", tenant.ID, teacher.ID).FirstOrCreate(&staff).Error; err != nil {
			return err
		}
		samples := []struct {
			Slug, Title, Description, Category, Level, Lesson string
			Published, Enrolled                               bool
		}{
			{"demo-react", "React, from the ground up", "Build a strong foundation in components, state and the ideas behind modern web applications.", "Development", "Beginner", "Preview: Your first learning session", true, true},
			{"demo-design", "A fresh eye for visual design", "Find your visual voice through typography, colour and thoughtful layouts.", "Design", "Beginner", "Preview: Getting started with the studio", true, false},
			{"demo-docker", "Docker for everyday developers", "Make your development environment reproducible. Explore containers, images and local services.", "DevOps", "Intermediate", "Preview: Explore your local environment", true, true},
			{"demo-typescript", "A practical guide to TypeScript", "Turn JavaScript experience into safer, more expressive code. Your next course starts here.", "Development", "Intermediate", "", false, false},
		}
		samplePath := filepath.Join(storage.Root(), "sample.mp4")
		for _, s := range samples {
			course := models.Course{TenantID: tenant.ID, CreatedBy: staff.ID, Slug: s.Slug, Title: s.Title, Description: s.Description, Category: s.Category, Level: s.Level, IsPublished: s.Published}
			var existing models.Course
			if err := tx.Where("tenant_id = ? AND slug = ?", tenant.ID, s.Slug).First(&existing).Error; err == nil {
				continue
			} else if err != gorm.ErrRecordNotFound {
				return err
			}
			if err := tx.Create(&course).Error; err != nil {
				return err
			}
			if s.Published {
				content, err := os.ReadFile(samplePath)
				if err != nil {
					return fmt.Errorf("sample video missing; run npm run dev: %w", err)
				}
				name := fmt.Sprintf("demo-%d.mp4", course.ID)
				if err = os.WriteFile(filepath.Join(storage.Root(), name), content, 0600); err != nil {
					return err
				}
				asset := models.LocalAsset{Name: name, UserID: teacher.ID, CourseID: course.ID, ContentType: "video/mp4", OriginalName: "Learnyst preview.mp4"}
				if err = tx.Create(&asset).Error; err != nil {
					return err
				}
				video := models.CourseVideo{TenantID: tenant.ID, CourseID: course.ID, TeacherID: staff.ID, Title: s.Lesson, Description: "This short sample shows that video playback works locally. Teachers can upload their own MP4 or WebM lessons from the course editor.", VideoUrl: "/uploads/" + name, DurationSeconds: 18, Position: 1, IsPublished: true}
				if err = tx.Create(&video).Error; err != nil {
					return err
				}
				due := time.Now().Add(7 * 24 * time.Hour)
				assignment := models.Assignment{TenantID: tenant.ID, CourseID: course.ID, TeacherID: staff.ID, Title: "Put your learning into words", Description: "Describe one thing you want to build with the skills from this course. Outline the first three steps you would take and one question you would like to explore.", MaxMarks: 10, DueDate: &due}
				if err = tx.Create(&assignment).Error; err != nil {
					return err
				}
				if s.Enrolled {
					batch := models.Batch{TenantID: tenant.ID, CourseID: course.ID, Name: "Self-paced", IsActive: true}
					if err = tx.Create(&batch).Error; err != nil {
						return err
					}
					enrollment := models.Enrollment{TenantID: tenant.ID, BatchID: batch.ID, StudentID: student.ID, Status: "active"}
					if err = tx.Create(&enrollment).Error; err != nil {
						return err
					}
				}
			}
		}
		return nil
	})
}
