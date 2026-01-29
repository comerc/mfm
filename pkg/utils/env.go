package utils

import (
	"os"
	"path/filepath"
	"sync"

	"github.com/joho/godotenv"
)

var (
	envOnce sync.Once
)

// LoadEnvFromRoot загружает .env файл из корня проекта
func LoadEnvFromRoot() {
	envOnce.Do(func() {
		err := loadEnvFromRootImpl()
		if err != nil {
			panic(err.Error())
		}
	})
}

// loadEnvFromRootImpl содержит оригинальную реализацию загрузки .env файла
func loadEnvFromRootImpl() error {
	dir, _ := os.Getwd()
	// Поднимаемся вверх, пока не найдем .env или не упремся в корень системы
	for {
		path := filepath.Join(dir, ".env")
		if _, err := os.Stat(path); err == nil {
			return godotenv.Load(path)
		}

		parent := filepath.Dir(dir)
		if parent == dir { // Достигли корня системы
			break
		}
		dir = parent
	}
	return nil
}
