package engine
import (
	"archive/zip"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"lms-builder-service/internal/storage"
)
func BuildCustomReact(tenantSlug string, zipFilePath string) error {
	log.Printf("Starting Custom React Build for %s", tenantSlug)
	workDir := filepath.Join(os.TempDir(), "builder_react_"+tenantSlug)
	os.MkdirAll(workDir, os.ModePerm)
	defer os.RemoveAll(workDir)
	err := unzip(zipFilePath, workDir)
	if err != nil {
		return fmt.Errorf("failed to unzip source: %w", err)
	}
	log.Println("Running npm install...")
	cmdInstall := exec.Command("npm", "install")
	cmdInstall.Dir = workDir
	outInstall, errInstall := cmdInstall.CombinedOutput()
	if errInstall != nil {
		log.Printf("NPM Install Failed: %s", string(outInstall))
		return fmt.Errorf("npm install failed: %w", errInstall)
	}
	log.Println("Running npm run build...")
	cmdBuild := exec.Command("npm", "run", "build")
	cmdBuild.Dir = workDir
	outBuild, errBuild := cmdBuild.CombinedOutput()
	if errBuild != nil {
		log.Printf("NPM Build Failed: %s", string(outBuild))
		return fmt.Errorf("npm build failed: %w", errBuild)
	}
	outputDir := filepath.Join(workDir, "dist")
	if _, err := os.Stat(outputDir); os.IsNotExist(err) {
		outputDir = filepath.Join(workDir, "build") 
		if _, err := os.Stat(outputDir); os.IsNotExist(err) {
			outputDir = filepath.Join(workDir, "out") 
		}
	}
	err = storage.UploadDirectory(tenantSlug, outputDir)
	if err != nil {
		return fmt.Errorf("failed to upload dist to S3: %w", err)
	}
	log.Printf("Custom React Build Successful for %s!", tenantSlug)
	return nil
}
func unzip(src string, dest string) error {
	r, err := zip.OpenReader(src)
	if err != nil {
		return err
	}
	defer r.Close()
	for _, f := range r.File {
		fpath := filepath.Join(dest, f.Name)
		if f.FileInfo().IsDir() {
			os.MkdirAll(fpath, os.ModePerm)
			continue
		}
		if err = os.MkdirAll(filepath.Dir(fpath), os.ModePerm); err != nil {
			return err
		}
		outFile, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return err
		}
		rc, err := f.Open()
		if err != nil {
			return err
		}
		_, err = io.Copy(outFile, rc)
		outFile.Close()
		rc.Close()
		if err != nil {
			return err
		}
	}
	return nil
}
