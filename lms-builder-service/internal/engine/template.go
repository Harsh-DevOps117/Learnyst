package engine
import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"lms-builder-service/internal/storage"
)
func BuildTemplateSite(tenantSlug string, academyName string, color string) error {
	log.Printf("Starting Template Build for %s", tenantSlug)
	workDir := filepath.Join(os.TempDir(), "builder_template_"+tenantSlug)
	os.MkdirAll(workDir, os.ModePerm)
	defer os.RemoveAll(workDir)
	htmlTemplate := `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>{{ACADEMY_NAME}}</title>
    <style>
        body {
            margin: 0;
            font-family: 'Inter', sans-serif;
            background-color: #0f172a;
            color: #f8fafc;
        }
        .hero {
            display: flex;
            flex-direction: column;
            align-items: center;
            justify-content: center;
            height: 100vh;
            background: linear-gradient(135deg, #0f172a 0%, {{PRIMARY_COLOR}} 100%);
            text-align: center;
            padding: 20px;
        }
        h1 {
            font-size: 4rem;
            margin-bottom: 10px;
            text-shadow: 0 4px 10px rgba(0,0,0,0.5);
        }
        p {
            font-size: 1.5rem;
            opacity: 0.8;
            max-width: 600px;
        }
        .btn {
            margin-top: 30px;
            padding: 15px 30px;
            background-color: {{PRIMARY_COLOR}};
            color: white;
            text-decoration: none;
            font-size: 1.2rem;
            border-radius: 8px;
            font-weight: bold;
            box-shadow: 0 4px 15px {{PRIMARY_COLOR}}80;
            transition: transform 0.2s;
        }
        .btn:hover {
            transform: scale(1.05);
        }
    </style>
</head>
<body>
    <div class="hero">
        <h1>{{ACADEMY_NAME}}</h1>
        <p>Your journey to mastery starts here. Join our community of learners today.</p>
        <a href="#" class="btn">Explore Courses</a>
    </div>
</body>
</html>`
	htmlTemplate = strings.ReplaceAll(htmlTemplate, "{{ACADEMY_NAME}}", academyName)
	htmlTemplate = strings.ReplaceAll(htmlTemplate, "{{PRIMARY_COLOR}}", color)
	indexPath := filepath.Join(workDir, "index.html")
	err := os.WriteFile(indexPath, []byte(htmlTemplate), 0644)
	if err != nil {
		return fmt.Errorf("failed to write template: %w", err)
	}
	err = storage.UploadDirectory(tenantSlug, workDir)
	if err != nil {
		return fmt.Errorf("failed to upload template to S3: %w", err)
	}
	log.Printf("Template Build Successful for %s", tenantSlug)
	return nil
}
