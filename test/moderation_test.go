package test

import (
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"strings"
	"testing"

	filereader "github.com/comerc/nsfw-mod/internal/repo/file_reader"
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

	fileReader := &filereader.Repo{}
	service := moderation.New(tempDir, fileReader)
	result := service.Moderate(testFile)

	if result.Error != "" {
		t.Errorf("expected no error, got %s", result.Error)
	}

	if result.IsNSFW {
		t.Errorf("expected IsNSFW to be false, got true")
	}

	if result.Confidence != 0.95 {
		t.Errorf("expected confidence 0.95, got %f", result.Confidence)
	}
}

func TestModerationService_Moderate_FileNotFound(t *testing.T) {
	tempDir := t.TempDir()
	fileReader := &filereader.Repo{}
	service := moderation.New(tempDir, fileReader)

	nonExistentFile := filepath.Join(tempDir, "nonexistent.png")
	result := service.Moderate(nonExistentFile)

	if result.Error == "" {
		t.Error("expected error for non-existent file")
	}

	if !strings.Contains(result.Error, "file not found") || !strings.Contains(result.Error, "nonexistent.png") {
		t.Errorf("expected error containing 'file not found' and 'nonexistent.png', got %s", result.Error)
	}
}
