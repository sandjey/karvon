package storage

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/google/uuid"
)

type Storage interface {
	Save(dir, filename string, r io.Reader) (string, error)
}

type LocalStorage struct {
	BasePath string
}

func NewLocal(basePath string) *LocalStorage {
	return &LocalStorage{BasePath: basePath}
}

func (s *LocalStorage) Save(dir, originalName string, r io.Reader) (string, error) {
	fullDir := filepath.Join(s.BasePath, dir)
	if err := os.MkdirAll(fullDir, 0755); err != nil {
		return "", fmt.Errorf("create dir: %w", err)
	}

	ext := filepath.Ext(originalName)
	filename := fmt.Sprintf("%d_%s%s", time.Now().Unix(), uuid.NewString()[:8], ext)
	fullPath := filepath.Join(fullDir, filename)

	f, err := os.Create(fullPath)
	if err != nil {
		return "", fmt.Errorf("create file: %w", err)
	}
	defer f.Close()

	if _, err := io.Copy(f, r); err != nil {
		return "", fmt.Errorf("write file: %w", err)
	}

	return "/uploads/" + filepath.ToSlash(filepath.Join(dir, filename)), nil
}
