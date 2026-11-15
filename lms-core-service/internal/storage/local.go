package storage

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

func Root() string {
	root := os.Getenv("STORAGE_DIR")
	if root == "" {
		root = "../.local/uploads"
	}
	return root
}
func InitStorage() {
	if err := os.MkdirAll(Root(), 0700); err != nil {
		panic("Cannot create local upload directory")
	}
}
func safePath(name string) (string, error) {
	clean := filepath.Clean(name)
	if clean == "." || filepath.IsAbs(clean) || clean == ".." || strings.HasPrefix(clean, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("invalid file path")
	}
	return filepath.Join(Root(), clean), nil
}
func UploadStream(ctx context.Context, name string, r io.Reader, contentType string) (string, error) {
	path, err := safePath(name)
	if err != nil {
		return "", err
	}
	if err = os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return "", err
	}
	f, err := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if err != nil {
		return "", err
	}
	_, err = io.Copy(f, r)
	closeErr := f.Close()
	if err != nil {
		_ = os.Remove(path)
		return "", err
	}
	if closeErr != nil {
		return "", closeErr
	}
	return "/uploads/" + filepath.ToSlash(name), nil
}
func DeleteBlob(ctx context.Context, name string) error {
	if !strings.HasPrefix(name, "/uploads/") {
		return nil
	}
	path, err := safePath(strings.TrimPrefix(name, "/uploads/"))
	if err != nil {
		return err
	}
	err = os.Remove(path)
	if os.IsNotExist(err) {
		return nil
	}
	return err
}
