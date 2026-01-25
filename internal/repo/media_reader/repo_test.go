package mediareader

import (
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestService_Read_Success(t *testing.T) {
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

	service := &Repo{}
	data, err := service.Read(testFile)

	if err != nil {
		t.Errorf("expected no error, got %s", err)
	}

	if len(data) == 0 {
		t.Error("expected image data, got empty")
	}
}

func TestService_Read_FileNotFound(t *testing.T) {
	tempDir := t.TempDir()
	service := &Repo{}

	nonExistentFile := filepath.Join(tempDir, "nonexistent.png")
	_, err := service.Read(nonExistentFile)

	if err == nil {
		t.Error("expected error for non-existent file")
	}

	if !strings.Contains(err.Error(), "file not found") {
		t.Errorf("expected error containing 'file not found', got %s", err.Error())
	}
}

func TestService_Read_UnsupportedType(t *testing.T) {
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test.txt")

	// Create a file with invalid content
	err := os.WriteFile(testFile, []byte("this is not an image"), 0644)
	if err != nil {
		t.Fatal(err)
	}

	service := &Repo{}
	_, err = service.Read(testFile)

	if err == nil {
		t.Error("expected error for unsupported file type")
	}

	if !strings.Contains(err.Error(), "unsupported file type") {
		t.Errorf("expected 'unsupported file type' error, got %s", err.Error())
	}
}
