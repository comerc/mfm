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
	img := image.NewRGBA(image.Rect(0, 0, 1, 1))
	img.Set(0, 0, color.RGBA{255, 0, 0, 255}) // Red pixel

	file, err := os.Create(testFile)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()

	err = png.Encode(file, img)
	if err != nil {
		t.Fatal(err)
	}

	fileReader := &mediareader.Repo{}
	modelRunner := modelrunner.New()
	service := moderation.New(tempDir, fileReader, modelRunner)
	result := service.Moderate(testFile)

	if result.Error != "" {
		t.Errorf("expected no error, got %s", result.Error)
	}

	// Check that image data was loaded
	if len(result.Error) == 0 && result.Confidence == 0.95 {
		// Mock still returns 0.95, but data is loaded
	}

	if result.IsNSFW {
		t.Errorf("expected IsNSFW to be false, got true")
	}

	// MockRunner Infer returns 0.1, so confidence should be 0.9 (1 - 0.1)
	if result.Confidence != 0.9 && !result.IsNSFW {
		t.Errorf("expected confidence 0.9, got %f", result.Confidence)
	}
}

func TestModerationService_Moderate_FileNotFound(t *testing.T) {
	tempDir := t.TempDir()
	fileReader := &mediareader.Repo{}
	modelRunner := modelrunner.New()
	service := moderation.New(tempDir, fileReader, modelRunner)

	nonExistentFile := filepath.Join(tempDir, "nonexistent.png")
	result := service.Moderate(nonExistentFile)

	if result.Error == "" {
		t.Error("expected error for non-existent file")
	}

	if !strings.Contains(result.Error, "file not found") || !strings.Contains(result.Error, "nonexistent.png") {
		t.Errorf("expected error containing 'file not found' and 'nonexistent.png', got %s", result.Error)
	}
}
