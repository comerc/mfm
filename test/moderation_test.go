package test

import (
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"strings"
	"testing"

	mediareader "github.com/comerc/nsfw-mod/internal/repo/media_reader"
	modelrunner "github.com/comerc/nsfw-mod/internal/repo/model_runner"
	"github.com/comerc/nsfw-mod/internal/service/moderation"
)

func TestModerationService_Moderate_Success(t *testing.T) {
	// Create temporary directory for testing
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test.png")

	// Create a simple test image
	img := image.NewRGBA(image.Rect(0, 0, 224, 224))
	for y := 0; y < 224; y++ {
		for x := 0; x < 224; x++ {
			img.Set(x, y, color.RGBA{255, 0, 0, 255}) // Red image
		}
	}

	file, err := os.Create(testFile)
	if err != nil {
		t.Fatal(err)
	}

	err = png.Encode(file, img)
	file.Close()
	if err != nil {
		t.Fatal(err)
	}

	defer func() {
		if r := recover(); r != nil {
			t.Logf("Recovered from panic used for ONNX init failure: %v", r)
			t.Skip("ONNX Runtime initialization failed, skipping test")
		}
	}()

	mediaReader := mediareader.New()
	modelRunner := modelrunner.New()
	service := moderation.New(tempDir, mediaReader, modelRunner)
	result := service.Moderate(testFile)

	if result.Error != "" {
		t.Errorf("expected no error, got %s", result.Error)
	}

	// Simple red image should not be NSFW
	if result.IsNSFW {
		t.Errorf("expected IsNSFW to be false, got true")
	}
}

func TestModerationService_Moderate_RealAssets(t *testing.T) {
	// Инициализируем репо с перехватом паники (если нет либы)
	defer func() {
		if r := recover(); r != nil {
			t.Skipf("Skipping real assets test due to initialization failure (missing lib?): %v", r)
		}
	}()
	modelRunner := modelrunner.New()

	// Проверяем наличие модели, если её нет - пропускаем интеграционные тесты
	if _, err := os.Stat("../assets/model.onnx"); os.IsNotExist(err) {
		t.Skip("Model file not found in ../assets/model.onnx, skipping integration test")
	}

	mediaReader := mediareader.New()
	// Директория не важна для этого теста, так как передаем полный путь
	service := moderation.New(".", mediaReader, modelRunner)

	assetsDir := "../assets"
	files, err := os.ReadDir(assetsDir)
	if err != nil {
		t.Fatalf("Failed to read assets directory: %v", err)
	}

	for _, f := range files {
		if f.IsDir() || !strings.HasSuffix(f.Name(), ".png") {
			continue
		}

		t.Run(f.Name(), func(t *testing.T) {
			fullPath := filepath.Join(assetsDir, f.Name())
			result := service.Moderate(fullPath)

			if result.Error != "" {
				t.Fatalf("Processing failed for %s: %s", f.Name(), result.Error)
			}

			t.Logf("Image: %s, IsNSFW: %v, Score: %.4f", f.Name(), result.IsNSFW, result.Score)
		})
	}
}

func TestModerationService_Moderate_FileNotFound(t *testing.T) {
	tempDir := t.TempDir()
	mediaReader := mediareader.New()
	modelRunner := modelrunner.New()
	service := moderation.New(tempDir, mediaReader, modelRunner)

	nonExistentFile := filepath.Join(tempDir, "nonexistent.png")
	result := service.Moderate(nonExistentFile)

	if result.Error == "" {
		t.Error("expected error for non-existent file")
	}

	if !strings.Contains(result.Error, "file not found") || !strings.Contains(result.Error, "nonexistent.png") {
		t.Errorf("expected error containing 'file not found' and 'nonexistent.png', got %s", result.Error)
	}
}
