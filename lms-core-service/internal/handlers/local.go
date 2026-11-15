package handlers

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"io"
	"lms-core-service/internal/database"
	"lms-core-service/internal/mail"
	m "lms-core-service/internal/models"
	"lms-core-service/internal/storage"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

func LocalRouter() *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Recovery())
	_ = r.SetTrustedProxies(nil)
	r.Use(func(c *gin.Context) {
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("Cache-Control", "no-store")
		if c.Request.Method != "GET" && c.Request.Method != "HEAD" {
			if origin := c.GetHeader("Origin"); origin != "" && origin != strings.TrimRight(os.Getenv("FRONTEND_URL"), "/") {
				c.AbortWithStatusJSON(403, gin.H{"error": "Request origin is not allowed."})
				return
			}
			limit := int64(1 << 20)
			if c.Request.URL.Path == "/api/v1/upload" {
				limit = 256 << 20
			}
			c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, limit)
		}
		c.Next()
	})
	r.GET("/health", func(c *gin.Context) { c.JSON(200, gin.H{"status": "ok", "service": "core", "storage": "local"}) })
	r.GET("/uploads/:name", serveLocalAsset)
	r.HEAD("/uploads/:name", serveLocalAsset)
	// Public academy pages receive metadata only; draft lessons are never exposed.
	r.GET("/api/v1/tenants/slug/:slug", publicAcademy)
	v := r.Group("/api/v1", requireSession)
	v.GET("/me", localProfile)
	v.GET("/workspaces", listWorkspaces)
	v.PUT("/workspaces/current", selectWorkspace)
	v.GET("/academy", getAcademy)
	v.PUT("/academy", updateAcademy)
	v.POST("/academy/invitations", inviteAcademyTeacher)
	v.PUT("/academy/teachers/:id", updateAcademyTeacher)
	v.PUT("/me", updateLocalProfile)
	v.GET("/courses", listLocalCourses)
	v.POST("/courses", createLocalCourse)
	v.GET("/courses/:id", getLocalCourse)
	v.PUT("/courses/:id", updateLocalCourse)
	v.DELETE("/courses/:id", deleteLocalCourse)
	v.POST("/courses/:id/enroll", enrollLocalCourse)
	v.POST("/courses/:id/invitations", sendLocalInvitation)
	v.POST("/invitations/accept", acceptLocalInvitation)
	v.POST("/courses/:id/videos", addLocalLesson)
	v.POST("/courses/:id/assignments", addLocalAssignment)
	v.POST("/upload", uploadLocalAsset)
	v.POST("/progress/video", saveLocalProgress)
	v.GET("/assignments", listLocalAssignments)
	v.POST("/assignments/:id/submit", submitLocalAssignment)
	v.PUT("/submissions/:id/grade", gradeLocalSubmission)
	v.GET("/students", listLocalStudents)
	return r
}
func sessionUser(c *gin.Context) (*m.User, error) {
	token, err := c.Cookie("learnyst_session")
	if err != nil || len(token) != 64 {
		return nil, errors.New("no session")
	}
	sum := sha256.Sum256([]byte(token))
	var session m.AuthSession
	if database.DB.Where("id = ? AND expires_at > ?", hex.EncodeToString(sum[:]), time.Now()).First(&session).Error != nil {
		return nil, errors.New("expired session")
	}
	var user m.User
	if database.DB.Where("id = ? AND is_active = ?", session.UserID, true).First(&user).Error != nil {
		return nil, errors.New("inactive user")
	}
	return &user, nil
}
func requireSession(c *gin.Context) {
	u, err := sessionUser(c)
	if err != nil {
		c.AbortWithStatusJSON(401, gin.H{"error": "Your session has expired. Please sign in again."})
		return
	}
	c.Set("local_user", u)
	c.Next()
}
func currentUser(c *gin.Context) *m.User              { return c.MustGet("local_user").(*m.User) }
func fail(c *gin.Context, status int, message string) { c.JSON(status, gin.H{"error": message}) }
func staffUser(u *m.User) bool                        { return u.RoleID == 1 || u.RoleID == 2 }
func workspace(u *m.User) (*m.TenantStaff, error) {
	if !staffUser(u) {
		return nil, errors.New("teacher access required")
	}
	var staff m.TenantStaff
	var preferred m.WorkspacePreference
	if database.DB.First(&preferred, "user_id = ?", u.ID).Error == nil {
		if database.DB.Where("user_id = ? AND tenant_id = ? AND is_active = ?", u.ID, preferred.TenantID, true).First(&staff).Error == nil {
			return &staff, nil
		}
	}
	if database.DB.Where("user_id = ? AND is_active = ?", u.ID, true).First(&staff).Error == nil {
		return &staff, nil
	}
	err := database.DB.Transaction(func(tx *gorm.DB) error {
		// Dashboard requests arrive together on a new account. Serialize its
		// initial academy creation, then recheck membership under the lock.
		var locked m.User
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&locked, u.ID).Error; err != nil {
			return err
		}
		if err := tx.Where("user_id = ? AND is_active = ?", u.ID, true).First(&staff).Error; err == nil {
			return nil
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
		tenant := m.Tenant{OwnerUserID: u.ID, AcademyName: u.FirstName + "’s Academy", AcademySlug: fmt.Sprintf("academy-%d", u.ID)}
		if err := tx.Where("academy_slug = ?", tenant.AcademySlug).FirstOrCreate(&tenant).Error; err != nil {
			return err
		}
		staff = m.TenantStaff{TenantID: tenant.ID, UserID: u.ID, RoleID: uint(u.RoleID), IsActive: true}
		return tx.Where("tenant_id = ? AND user_id = ?", tenant.ID, u.ID).FirstOrCreate(&staff).Error
	})
	return &staff, err
}
func manages(u *m.User, course *m.Course) bool {
	if !staffUser(u) {
		return false
	}
	var tenant m.Tenant
	if database.DB.First(&tenant, course.TenantID).Error == nil && tenant.OwnerUserID == u.ID {
		return true
	}
	var n int64
	database.DB.Model(&m.TenantStaff{}).Where("id = ? AND user_id = ? AND tenant_id = ? AND is_active = ?", course.CreatedBy, u.ID, course.TenantID, true).Count(&n)
	return n > 0
}
func enrolled(userID, courseID uint) bool {
	var n int64
	database.DB.Table("enrollments").Joins("JOIN batches ON batches.id = enrollments.batch_id").Where("batches.course_id = ? AND enrollments.student_id = ? AND enrollments.status = ?", courseID, userID, "active").Count(&n)
	return n > 0
}
func readable(u *m.User, course *m.Course) bool {
	return manages(u, course) || (u.RoleID == 3 && course.IsPublished && enrolled(u.ID, course.ID))
}
func courseFromParam(c *gin.Context, manage bool) (*m.Course, bool) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil || id == 0 {
		fail(c, 404, "Course not found.")
		return nil, false
	}
	var course m.Course
	if database.DB.First(&course, uint(id)).Error != nil {
		fail(c, 404, "Course not found.")
		return nil, false
	}
	u := currentUser(c)
	allowed := readable(u, &course)
	if manage {
		allowed = manages(u, &course)
	}
	if !allowed {
		fail(c, 403, "You do not have access to this course.")
		return nil, false
	}
	return &course, true
}

type courseSummary struct {
	ID               uint      `json:"id"`
	TenantID         uint      `json:"tenant_id"`
	Title            string    `json:"title"`
	Description      string    `json:"description"`
	Category         string    `json:"category"`
	Level            string    `json:"level"`
	ThumbnailURL     string    `json:"thumbnail_url"`
	Price            float64   `json:"price"`
	IsPublished      bool      `json:"is_published"`
	Instructor       string    `json:"instructor"`
	LessonCount      int64     `json:"lesson_count"`
	StudentCount     int64     `json:"student_count"`
	DurationSeconds  int64     `json:"duration_seconds"`
	CompletedLessons int64     `json:"completed_lessons"`
	Progress         int       `json:"progress"`
	Enrolled         bool      `json:"enrolled"`
	UpdatedAt        time.Time `json:"updated_at"`
}

func summary(course m.Course, userID uint) courseSummary {
	s := courseSummary{ID: course.ID, TenantID: course.TenantID, Title: course.Title, Description: course.Description, Category: course.Category, Level: course.Level, ThumbnailURL: course.ThumbnailUrl, Price: course.Price, IsPublished: course.IsPublished, UpdatedAt: course.UpdatedAt}
	database.DB.Model(&m.CourseVideo{}).Where("course_id = ?", course.ID).Count(&s.LessonCount)
	database.DB.Model(&m.CourseVideo{}).Where("course_id = ?", course.ID).Select("COALESCE(SUM(duration_seconds),0)").Scan(&s.DurationSeconds)
	database.DB.Table("enrollments").Joins("JOIN batches ON batches.id = enrollments.batch_id").Where("batches.course_id = ? AND enrollments.status = ?", course.ID, "active").Distinct("enrollments.student_id").Count(&s.StudentCount)
	var staff m.TenantStaff
	if database.DB.Preload("User").First(&staff, course.CreatedBy).Error == nil {
		s.Instructor = staff.User.FirstName + " " + staff.User.LastName
	}
	if userID > 0 {
		s.Enrolled = enrolled(userID, course.ID)
		var percent float64
		database.DB.Table("video_progresses").Joins("JOIN course_videos ON course_videos.id = video_progresses.video_id").Where("course_videos.course_id = ? AND video_progresses.student_id = ?", course.ID, userID).Select("COALESCE(SUM(video_progresses.completion_percent),0)").Scan(&percent)
		database.DB.Table("video_progresses").Joins("JOIN course_videos ON course_videos.id = video_progresses.video_id").Where("course_videos.course_id = ? AND video_progresses.student_id = ? AND completion_percent >= 90", course.ID, userID).Count(&s.CompletedLessons)
		if s.LessonCount > 0 {
			s.Progress = int(percent / float64(s.LessonCount))
			if s.Progress > 100 {
				s.Progress = 100
			}
		}
	}
	return s
}
func localProfile(c *gin.Context) {
	u := currentUser(c)
	var tenantID uint
	if staffUser(u) {
		staff, err := workspace(u)
		if err != nil {
			fail(c, 500, "Could not load your academy.")
			return
		}
		tenantID = staff.TenantID
	}
	c.JSON(200, gin.H{"user": u, "tenant_id": tenantID})
}
func updateLocalProfile(c *gin.Context) {
	var req struct {
		FirstName string `json:"first_name" binding:"required,max=100"`
		LastName  string `json:"last_name" binding:"required,max=100"`
		Location  string `json:"location" binding:"max=255"`
	}
	if c.ShouldBindJSON(&req) != nil || strings.TrimSpace(req.FirstName) == "" || strings.TrimSpace(req.LastName) == "" {
		fail(c, 400, "Enter your first and last name.")
		return
	}
	if err := database.DB.Model(currentUser(c)).Updates(map[string]interface{}{"first_name": strings.TrimSpace(req.FirstName), "last_name": strings.TrimSpace(req.LastName), "location": strings.TrimSpace(req.Location)}).Error; err != nil {
		fail(c, 500, "Could not save your profile.")
		return
	}
	localProfile(c)
}
func listLocalCourses(c *gin.Context) {
	u := currentUser(c)
	var courses []m.Course
	q := database.DB.Order("updated_at DESC, id DESC")
	if c.Query("scope") == "catalog" {
		q = q.Where("is_published = ?", true)
	} else if staffUser(u) {
		staff, err := workspace(u)
		if err != nil {
			fail(c, 500, "Could not load your academy.")
			return
		}
		var tenant m.Tenant
		database.DB.First(&tenant, staff.TenantID)
		if tenant.OwnerUserID == u.ID {
			q = q.Where("tenant_id = ?", staff.TenantID)
		} else {
			q = q.Where("created_by = ?", staff.ID)
		}
	} else {
		q = q.Where("is_published = ? AND id IN (SELECT batches.course_id FROM batches JOIN enrollments ON enrollments.batch_id = batches.id WHERE enrollments.student_id = ? AND enrollments.status = ?)", true, u.ID, "active")
	}
	if err := q.Find(&courses).Error; err != nil {
		fail(c, 500, "Could not load courses.")
		return
	}
	result := make([]courseSummary, 0, len(courses))
	for _, course := range courses {
		result = append(result, summary(course, u.ID))
	}
	c.JSON(200, result)
}
func publicAcademy(c *gin.Context) {
	var tenant m.Tenant
	if database.DB.Where("academy_slug = ? AND is_active = ? AND is_suspended = ?", c.Param("slug"), true, false).First(&tenant).Error != nil {
		fail(c, 404, "Academy not found.")
		return
	}
	var courses []m.Course
	if database.DB.Where("tenant_id = ? AND is_published = ?", tenant.ID, true).Find(&courses).Error != nil {
		fail(c, 500, "Could not load academy.")
		return
	}
	result := make([]courseSummary, 0, len(courses))
	for _, course := range courses {
		result = append(result, summary(course, 0))
	}
	c.JSON(200, gin.H{"tenant": gin.H{"id": tenant.ID, "academy_name": tenant.AcademyName, "academy_slug": tenant.AcademySlug, "description": tenant.Description}, "courses": result})
}
func createLocalCourse(c *gin.Context) {
	u := currentUser(c)
	if !staffUser(u) {
		fail(c, 403, "Only teachers can create courses.")
		return
	}
	var req struct {
		Title       string `json:"title" binding:"required,max=255"`
		Description string `json:"description" binding:"max=10000"`
		Category    string `json:"category" binding:"max=100"`
		Level       string `json:"level" binding:"oneof=Beginner Intermediate Advanced"`
	}
	if c.ShouldBindJSON(&req) != nil || strings.TrimSpace(req.Title) == "" {
		fail(c, 400, "Enter a course title and choose a valid level.")
		return
	}
	staff, err := workspace(u)
	if err != nil {
		fail(c, 500, "Could not load your academy.")
		return
	}
	b := make([]byte, 8)
	if _, err = rand.Read(b); err != nil {
		fail(c, 500, "Please try again.")
		return
	}
	course := m.Course{TenantID: staff.TenantID, CreatedBy: staff.ID, Title: strings.TrimSpace(req.Title), Description: strings.TrimSpace(req.Description), Category: req.Category, Level: req.Level, Slug: hex.EncodeToString(b)}
	if database.DB.Create(&course).Error != nil {
		fail(c, 500, "Could not create the course.")
		return
	}
	c.JSON(201, summary(course, u.ID))
}
func getLocalCourse(c *gin.Context) {
	course, ok := courseFromParam(c, false)
	if !ok {
		return
	}
	var videos []m.CourseVideo
	var assignments []m.Assignment
	var progress []m.VideoProgress
	if database.DB.Where("course_id = ?", course.ID).Order("position ASC, id ASC").Find(&videos).Error != nil {
		fail(c, 500, "Could not load lessons.")
		return
	}
	database.DB.Where("course_id = ?", course.ID).Find(&assignments)
	database.DB.Where("student_id = ? AND video_id IN (SELECT id FROM course_videos WHERE course_id = ?)", currentUser(c).ID, course.ID).Find(&progress)
	pv := make([]gin.H, 0, len(progress))
	for _, p := range progress {
		pv = append(pv, gin.H{"video_id": p.VideoID, "watch_time_seconds": p.WatchTimeSeconds, "completion_percent": p.CompletionPercent})
	}
	c.JSON(200, gin.H{"course": summary(*course, currentUser(c).ID), "videos": videos, "assignments": assignments, "progress": pv, "can_manage": manages(currentUser(c), course)})
}
func updateLocalCourse(c *gin.Context) {
	course, ok := courseFromParam(c, true)
	if !ok {
		return
	}
	var req struct {
		Title       *string `json:"title" binding:"omitempty,max=255"`
		Description *string `json:"description" binding:"omitempty,max=10000"`
		IsPublished *bool   `json:"is_published"`
	}
	if c.ShouldBindJSON(&req) != nil {
		fail(c, 400, "Enter valid course details.")
		return
	}
	updates := map[string]interface{}{}
	if req.Title != nil {
		if strings.TrimSpace(*req.Title) == "" {
			fail(c, 400, "Course title cannot be empty.")
			return
		}
		updates["title"] = strings.TrimSpace(*req.Title)
	}
	if req.Description != nil {
		updates["description"] = strings.TrimSpace(*req.Description)
	}
	if req.IsPublished != nil {
		if *req.IsPublished {
			var n int64
			database.DB.Model(&m.CourseVideo{}).Where("course_id = ?", course.ID).Count(&n)
			if n == 0 {
				fail(c, 400, "Add at least one lesson before publishing.")
				return
			}
		}
		updates["is_published"] = *req.IsPublished
	}
	if database.DB.Model(course).Updates(updates).Error != nil {
		fail(c, 500, "Could not save course.")
		return
	}
	database.DB.First(course, course.ID)
	c.JSON(200, summary(*course, currentUser(c).ID))
}
func deleteLocalCourse(c *gin.Context) {
	course, ok := courseFromParam(c, true)
	if !ok {
		return
	}
	var count int64
	database.DB.Table("enrollments").Joins("JOIN batches ON batches.id = enrollments.batch_id").Where("batches.course_id = ?", course.ID).Count(&count)
	if count > 0 {
		fail(c, 409, "This course has students. Unpublish it to keep their progress safe.")
		return
	}
	var assets []m.LocalAsset
	database.DB.Where("course_id = ?", course.ID).Find(&assets)
	err := database.DB.Transaction(func(tx *gorm.DB) error {
		// Keep deletion atomic: any dependent relation failure rolls back the whole course.
		for _, entry := range []struct {
			Model interface{}
			Where string
		}{
			{&m.VideoProgress{}, "video_id IN (SELECT id FROM course_videos WHERE course_id = ?)"},
			{&m.AssignmentSubmission{}, "assignment_id IN (SELECT id FROM assignments WHERE course_id = ?)"},
			{&m.CourseVideo{}, "course_id = ?"}, {&m.Assignment{}, "course_id = ?"}, {&m.Note{}, "course_id = ?"},
			{&m.CourseTeacher{}, "course_id = ?"}, {&m.Invitation{}, "course_id = ?"}, {&m.CourseModule{}, "course_id = ?"},
			{&m.Batch{}, "course_id = ?"}, {&m.LocalAsset{}, "course_id = ?"},
		} {
			if err := tx.Where(entry.Where, course.ID).Delete(entry.Model).Error; err != nil {
				return err
			}
		}
		return tx.Delete(course).Error
	})
	if err != nil {
		fail(c, 409, "This course still has related activity and could not be deleted.")
		return
	}
	for _, asset := range assets {
		_ = storage.DeleteBlob(c.Request.Context(), "/uploads/"+asset.Name)
	}
	c.JSON(200, gin.H{"message": "Course deleted."})
}
func enrollLocalCourse(c *gin.Context) {
	u := currentUser(c)
	if u.RoleID != 3 {
		fail(c, 403, "Sign in as a student to enroll.")
		return
	}
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil || id == 0 {
		fail(c, 404, "Course not found.")
		return
	}
	err = database.DB.Transaction(func(tx *gorm.DB) error {
		var course m.Course
		if tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ? AND is_published = ?", uint(id), true).First(&course).Error != nil {
			return errors.New("Course not found.")
		}
		if course.Price > 0 {
			return errors.New("Paid enrollment is not enabled locally.")
		}
		batch := m.Batch{CourseID: course.ID, TenantID: course.TenantID, Name: "Self-paced", IsActive: true}
		if err := tx.Where("course_id = ?", course.ID).FirstOrCreate(&batch).Error; err != nil {
			return err
		}
		e := m.Enrollment{TenantID: course.TenantID, BatchID: batch.ID, StudentID: u.ID, Status: "active"}
		return tx.Where("batch_id = ? AND student_id = ?", batch.ID, u.ID).FirstOrCreate(&e).Error
	})
	if err != nil {
		fail(c, 400, err.Error())
		return
	}
	c.JSON(200, gin.H{"message": "You’re enrolled. Let’s get learning."})
}
func uploadLocalAsset(c *gin.Context) {
	u := currentUser(c)
	if !staffUser(u) {
		fail(c, 403, "Only teachers can upload course files.")
		return
	}
	if err := c.Request.ParseMultipartForm(8 << 20); err != nil {
		fail(c, 413, "Choose a file smaller than 256 MB.")
		return
	}
	defer c.Request.MultipartForm.RemoveAll()
	courseID, _ := strconv.ParseUint(c.PostForm("course_id"), 10, 32)
	var course m.Course
	if courseID == 0 || database.DB.First(&course, uint(courseID)).Error != nil || !manages(u, &course) {
		fail(c, 403, "You cannot upload to this course.")
		return
	}
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		fail(c, 400, "Choose a file to upload.")
		return
	}
	defer file.Close()
	if header.Size == 0 {
		fail(c, 400, "The file is empty.")
		return
	}
	buf := make([]byte, 512)
	n, err := file.Read(buf)
	if err != nil && err != io.EOF {
		fail(c, 400, "Could not read the file.")
		return
	}
	contentType := http.DetectContentType(buf[:n])
	ext := strings.ToLower(filepath.Ext(header.Filename))
	allowed := map[string]string{".mp4": "video/mp4", ".webm": "video/webm", ".pdf": "application/pdf", ".png": "image/png", ".jpg": "image/jpeg", ".jpeg": "image/jpeg", ".webp": "image/webp"}
	if allowed[ext] == "" || allowed[ext] != contentType {
		fail(c, 400, "Use an MP4, WebM, PDF, PNG, JPEG or WebP file with a matching file type.")
		return
	}
	if _, err = file.Seek(0, 0); err != nil {
		fail(c, 400, "Could not read the file.")
		return
	}
	b := make([]byte, 16)
	if _, err = rand.Read(b); err != nil {
		fail(c, 500, "Please try again.")
		return
	}
	name := hex.EncodeToString(b) + ext
	url, err := storage.UploadStream(c.Request.Context(), name, file, contentType)
	if err != nil {
		fail(c, 500, "Could not save the file locally.")
		return
	}
	asset := m.LocalAsset{Name: name, UserID: u.ID, CourseID: course.ID, ContentType: contentType, OriginalName: filepath.Base(header.Filename)}
	if database.DB.Create(&asset).Error != nil {
		_ = storage.DeleteBlob(c.Request.Context(), url)
		fail(c, 500, "Could not record the upload.")
		return
	}
	c.JSON(201, gin.H{"url": url, "file_name": asset.OriginalName, "content_type": contentType})
}
func serveLocalAsset(c *gin.Context) {
	name := c.Param("name")
	if filepath.Base(name) != name {
		c.Status(404)
		return
	}
	var asset m.LocalAsset
	if database.DB.First(&asset, "name = ?", name).Error != nil {
		c.Status(404)
		return
	}
	u, err := sessionUser(c)
	if err != nil {
		c.Status(401)
		return
	}
	var course m.Course
	if database.DB.First(&course, asset.CourseID).Error != nil || !readable(u, &course) {
		c.Status(403)
		return
	}
	c.Header("Content-Type", asset.ContentType)
	if asset.ContentType == "application/pdf" {
		c.Header("Content-Disposition", "attachment")
	}
	// ServeFile supports byte ranges, so seeking works with local video playback.
	c.File(filepath.Join(storage.Root(), name))
}
func addLocalLesson(c *gin.Context) {
	course, ok := courseFromParam(c, true)
	if !ok {
		return
	}
	var req struct {
		Title       string `json:"title" binding:"required,max=255"`
		Description string `json:"description" binding:"max=10000"`
		VideoURL    string `json:"video_url" binding:"required"`
		NoteURL     string `json:"note_url"`
		Duration    int    `json:"duration_seconds" binding:"min=0,max=86400"`
	}
	if c.ShouldBindJSON(&req) != nil || strings.TrimSpace(req.Title) == "" {
		fail(c, 400, "Enter a lesson title and upload a video.")
		return
	}
	var asset m.LocalAsset
	if !strings.HasPrefix(req.VideoURL, "/uploads/") || database.DB.Where("name = ? AND course_id = ? AND content_type LIKE ?", strings.TrimPrefix(req.VideoURL, "/uploads/"), course.ID, "video/%").First(&asset).Error != nil {
		fail(c, 400, "Upload a video to this course first.")
		return
	}
	if req.NoteURL != "" {
		var n int64
		database.DB.Model(&m.LocalAsset{}).Where("name = ? AND course_id = ? AND content_type = ?", strings.TrimPrefix(req.NoteURL, "/uploads/"), course.ID, "application/pdf").Count(&n)
		if n == 0 {
			fail(c, 400, "Upload a PDF to this course first.")
			return
		}
	}
	staff, err := workspace(currentUser(c))
	if err != nil {
		fail(c, 500, "Could not load your academy.")
		return
	}
	var position int
	database.DB.Model(&m.CourseVideo{}).Where("course_id = ?", course.ID).Select("COALESCE(MAX(position),0)+1").Scan(&position)
	video := m.CourseVideo{TenantID: course.TenantID, CourseID: course.ID, TeacherID: staff.ID, Title: strings.TrimSpace(req.Title), Description: req.Description, VideoUrl: req.VideoURL, NoteUrl: req.NoteURL, DurationSeconds: req.Duration, Position: position, IsPublished: true}
	if database.DB.Create(&video).Error != nil {
		fail(c, 500, "Could not save the lesson.")
		return
	}
	c.JSON(201, video)
}
func saveLocalProgress(c *gin.Context) {
	u := currentUser(c)
	if u.RoleID != 3 {
		fail(c, 403, "Student access required.")
		return
	}
	var req struct {
		VideoID    uint `json:"video_id" binding:"required"`
		WatchTime  int  `json:"watch_time_seconds" binding:"min=0,max=86400"`
		Completion int  `json:"completion_percent" binding:"min=0,max=100"`
	}
	if c.ShouldBindJSON(&req) != nil {
		fail(c, 400, "Enter valid video progress.")
		return
	}
	var video m.CourseVideo
	var course m.Course
	if database.DB.First(&video, req.VideoID).Error != nil || database.DB.First(&course, video.CourseID).Error != nil || !readable(u, &course) {
		fail(c, 403, "You are not enrolled in this course.")
		return
	}
	now := time.Now()
	p := m.VideoProgress{StudentID: u.ID, VideoID: video.ID, WatchTimeSeconds: req.WatchTime, CompletionPercent: req.Completion, LastWatchedAt: &now}
	err := database.DB.Clauses(clause.OnConflict{Columns: []clause.Column{{Name: "student_id"}, {Name: "video_id"}}, DoUpdates: clause.Assignments(map[string]interface{}{"watch_time_seconds": gorm.Expr("GREATEST(video_progresses.watch_time_seconds, ?)", req.WatchTime), "completion_percent": gorm.Expr("GREATEST(video_progresses.completion_percent, ?)", req.Completion), "last_watched_at": now})}).Create(&p).Error
	if err != nil {
		fail(c, 500, "Could not save progress.")
		return
	}
	c.JSON(200, gin.H{"message": "Progress saved."})
}
func addLocalAssignment(c *gin.Context) {
	course, ok := courseFromParam(c, true)
	if !ok {
		return
	}
	var req struct {
		Title       string     `json:"title" binding:"required,max=255"`
		Description string     `json:"description" binding:"required,max=10000"`
		MaxMarks    int        `json:"max_marks" binding:"required,min=1,max=1000"`
		DueDate     *time.Time `json:"due_date"`
	}
	if c.ShouldBindJSON(&req) != nil || strings.TrimSpace(req.Title) == "" || strings.TrimSpace(req.Description) == "" {
		fail(c, 400, "Enter a title, instructions and maximum marks.")
		return
	}
	staff, err := workspace(currentUser(c))
	if err != nil {
		fail(c, 500, "Could not load your academy.")
		return
	}
	a := m.Assignment{TenantID: course.TenantID, CourseID: course.ID, TeacherID: staff.ID, Title: req.Title, Description: req.Description, MaxMarks: req.MaxMarks, DueDate: req.DueDate}
	if database.DB.Create(&a).Error != nil {
		fail(c, 500, "Could not create the assignment.")
		return
	}
	c.JSON(201, a)
}
func listLocalAssignments(c *gin.Context) {
	u := currentUser(c)
	var assignments []m.Assignment
	var courses []m.Course
	if staffUser(u) {
		staff, err := workspace(u)
		if err != nil {
			fail(c, 500, "Could not load your academy.")
			return
		}
		database.DB.Where("tenant_id = ?", staff.TenantID).Find(&courses)
	} else {
		database.DB.Where("is_published = ? AND id IN (SELECT batches.course_id FROM batches JOIN enrollments ON enrollments.batch_id=batches.id WHERE enrollments.student_id=? AND enrollments.status=?)", true, u.ID, "active").Find(&courses)
	}
	ids := []uint{}
	titles := map[uint]string{}
	for _, course := range courses {
		if !staffUser(u) || manages(u, &course) {
			ids = append(ids, course.ID)
			titles[course.ID] = course.Title
		}
	}
	if len(ids) > 0 {
		if database.DB.Where("course_id IN ?", ids).Order("created_at DESC").Find(&assignments).Error != nil {
			fail(c, 500, "Could not load assignments.")
			return
		}
	}
	result := []gin.H{}
	for _, a := range assignments {
		submissions := []gin.H{}
		var rows []m.AssignmentSubmission
		q := database.DB.Preload("Student").Where("assignment_id = ?", a.ID)
		if !staffUser(u) {
			q = q.Where("student_id = ?", u.ID)
		}
		q.Order("id DESC").Find(&rows)
		for _, s := range rows {
			submissions = append(submissions, gin.H{"id": s.ID, "student_name": s.Student.FirstName + " " + s.Student.LastName, "text": s.SubmissionTxt, "submitted_at": s.SubmittedAt, "marks": s.MarksObtained, "feedback": s.Feedback, "status": s.Status})
		}
		result = append(result, gin.H{"id": a.ID, "course_id": a.CourseID, "course_title": titles[a.CourseID], "title": a.Title, "description": a.Description, "max_marks": a.MaxMarks, "due_date": a.DueDate, "submissions": submissions})
	}
	c.JSON(200, result)
}
func submitLocalAssignment(c *gin.Context) {
	u := currentUser(c)
	if u.RoleID != 3 {
		fail(c, 403, "Student access required.")
		return
	}
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil || id == 0 {
		fail(c, 404, "Assignment not found.")
		return
	}
	var a m.Assignment
	var course m.Course
	if database.DB.First(&a, uint(id)).Error != nil || database.DB.First(&course, a.CourseID).Error != nil || !readable(u, &course) {
		fail(c, 403, "You do not have access to this assignment.")
		return
	}
	if a.DueDate != nil && time.Now().After(*a.DueDate) {
		fail(c, 400, "The submission deadline has passed.")
		return
	}
	var req struct {
		Text string `json:"text" binding:"required,max=20000"`
	}
	if c.ShouldBindJSON(&req) != nil || strings.TrimSpace(req.Text) == "" {
		fail(c, 400, "Write your response before submitting.")
		return
	}
	now := time.Now()
	s := m.AssignmentSubmission{AssignmentID: a.ID, StudentID: u.ID, SubmissionTxt: strings.TrimSpace(req.Text), SubmittedAt: &now, Status: "submitted"}
	// Only the first submission is accepted; graded work cannot be overwritten.
	if err := database.DB.Create(&s).Error; err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			fail(c, 409, "You have already submitted this assignment.")
		} else {
			fail(c, 500, "Could not submit your assignment.")
		}
		return
	}
	c.JSON(201, gin.H{"message": "Assignment submitted."})
}
func gradeLocalSubmission(c *gin.Context) {
	u := currentUser(c)
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil || id == 0 {
		fail(c, 404, "Submission not found.")
		return
	}
	var s m.AssignmentSubmission
	var a m.Assignment
	var course m.Course
	if database.DB.First(&s, uint(id)).Error != nil || database.DB.First(&a, s.AssignmentID).Error != nil || database.DB.First(&course, a.CourseID).Error != nil || !manages(u, &course) {
		fail(c, 403, "You cannot grade this submission.")
		return
	}
	var req struct {
		Marks    *int   `json:"marks" binding:"required,min=0"`
		Feedback string `json:"feedback" binding:"max=10000"`
	}
	if c.ShouldBindJSON(&req) != nil || *req.Marks > a.MaxMarks {
		fail(c, 400, "Marks must be between 0 and the assignment maximum.")
		return
	}
	staff, err := workspace(u)
	if err != nil {
		fail(c, 500, "Could not load your academy.")
		return
	}
	now := time.Now()
	if database.DB.Model(&s).Updates(map[string]interface{}{"marks_obtained": *req.Marks, "feedback": req.Feedback, "graded_by": staff.ID, "graded_at": now, "status": "graded"}).Error != nil {
		fail(c, 500, "Could not save the grade.")
		return
	}
	c.JSON(200, gin.H{"message": "Grade saved."})
}
func listLocalStudents(c *gin.Context) {
	u := currentUser(c)
	if !staffUser(u) {
		fail(c, 403, "Teacher access required.")
		return
	}
	staff, err := workspace(u)
	if err != nil {
		fail(c, 500, "Could not load your academy.")
		return
	}
	var courses []m.Course
	database.DB.Where("tenant_id = ?", staff.TenantID).Find(&courses)
	ids := []uint{}
	for _, course := range courses {
		if manages(u, &course) {
			ids = append(ids, course.ID)
		}
	}
	type row struct {
		ID          uint      `json:"id"`
		FirstName   string    `json:"first_name"`
		LastName    string    `json:"last_name"`
		Email       string    `json:"email"`
		CourseCount int64     `json:"course_count"`
		JoinedAt    time.Time `json:"joined_at"`
	}
	students := []row{}
	if len(ids) > 0 {
		err = database.DB.Table("users").Select("users.id, users.first_name, users.last_name, users.email, COUNT(DISTINCT batches.course_id) AS course_count, MIN(enrollments.enrolled_at) AS joined_at").Joins("JOIN enrollments ON enrollments.student_id = users.id").Joins("JOIN batches ON batches.id = enrollments.batch_id").Where("batches.course_id IN ? AND enrollments.status = ?", ids, "active").Group("users.id").Order("users.first_name").Scan(&students).Error
		if err != nil {
			fail(c, 500, "Could not load students.")
			return
		}
	}
	c.JSON(200, students)
}

func sendLocalInvitation(c *gin.Context) {
	course, ok := courseFromParam(c, true)
	if !ok {
		return
	}
	if !course.IsPublished || course.Price > 0 {
		fail(c, 400, "Publish a free course before inviting a student.")
		return
	}
	var req struct {
		Email string `json:"email" binding:"required,email,max=255"`
	}
	if c.ShouldBindJSON(&req) != nil {
		fail(c, 400, "Enter a valid student email.")
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
	invitation := m.Invitation{TenantID: course.TenantID, Email: strings.ToLower(req.Email), RoleID: 3, CourseID: &course.ID, InviteToken: hex.EncodeToString(sum[:]), ExpiresAt: &expiry}
	if database.DB.Create(&invitation).Error != nil {
		fail(c, 500, "Could not save invitation.")
		return
	}
	if err := mail.SendInvitationEmail(req.Email, course.Title, "", token); err != nil {
		database.DB.Delete(&invitation)
		fail(c, 503, "The local email inbox is unavailable. Please try again.")
		return
	}
	c.JSON(201, gin.H{"message": "Invitation sent. You can view it in the local inbox."})
}
func acceptLocalInvitation(c *gin.Context) {
	u := currentUser(c)
	if u.RoleID == 2 {
		acceptAcademyTeacher(c)
		return
	}
	if u.RoleID != 3 {
		fail(c, 403, "Sign in with the student account the invitation was sent to.")
		return
	}
	var req struct {
		Token string `json:"token" binding:"required,len=64,hexadecimal"`
	}
	if c.ShouldBindJSON(&req) != nil {
		fail(c, 400, "This invitation link is incomplete.")
		return
	}
	sum := sha256.Sum256([]byte(req.Token))
	var courseID uint
	err := database.DB.Transaction(func(tx *gorm.DB) error {
		var invitation m.Invitation
		if tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("invite_token = ? AND accepted = ? AND expires_at > ?", hex.EncodeToString(sum[:]), false, time.Now()).First(&invitation).Error != nil {
			return errors.New("This invitation has expired or already been used.")
		}
		if !strings.EqualFold(invitation.Email, u.Email) || invitation.RoleID != 3 || invitation.CourseID == nil {
			return errors.New("Sign in with the student email this invitation was sent to.")
		}
		var course m.Course
		if tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ? AND is_published = ? AND price = ?", *invitation.CourseID, true, 0).First(&course).Error != nil {
			return errors.New("This course is not available for enrollment.")
		}
		batch := m.Batch{CourseID: course.ID, TenantID: course.TenantID, Name: "Self-paced", IsActive: true}
		if err := tx.Where("course_id = ?", course.ID).FirstOrCreate(&batch).Error; err != nil {
			return err
		}
		e := m.Enrollment{TenantID: course.TenantID, BatchID: batch.ID, StudentID: u.ID, Status: "active"}
		if err := tx.Where("batch_id = ? AND student_id = ?", batch.ID, u.ID).FirstOrCreate(&e).Error; err != nil {
			return err
		}
		courseID = course.ID
		return tx.Model(&invitation).Update("accepted", true).Error
	})
	if err != nil {
		fail(c, 400, err.Error())
		return
	}
	c.JSON(200, gin.H{"message": "Welcome to the course.", "course_id": courseID})
}
