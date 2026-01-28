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
	t.Parallel()
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

	repo := New()
	frames, err := repo.Read(testFile)

	if err != nil {
		t.Errorf("expected no error, got %s", err)
	}

	if len(frames) != 1 {
		t.Errorf("expected 1 frame, got %d", len(frames))
	}

	frame := frames[0]
	expectedSize := 224 * 224 * 3
	if len(frame) != expectedSize {
		t.Errorf("expected frame size %d, got %d", expectedSize, len(frame))
	}
}

func TestService_Read_FileNotFound(t *testing.T) {
	t.Parallel()
	tempDir := t.TempDir()
	repo := New()

	nonExistentFile := filepath.Join(tempDir, "nonexistent.png")
	_, err := repo.Read(nonExistentFile)

	if err == nil {
		t.Error("expected error for non-existent file")
	}

	if !strings.Contains(err.Error(), "file not found") {
		t.Errorf("expected error containing 'file not found', got %s", err.Error())
	}
}

func TestService_Read_UnsupportedType(t *testing.T) {
	t.Parallel()
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test.txt")

	// Create a file with invalid content
	err := os.WriteFile(testFile, []byte("this is not an image"), 0600)
	if err != nil {
		t.Fatal(err)
	}

	repo := New()
	_, err = repo.Read(testFile)

	if err == nil {
		t.Error("expected error for unsupported file type")
	}

	if !strings.Contains(err.Error(), "unsupported file type") {
		t.Errorf("expected 'unsupported file type' error, got %s", err.Error())
	}
}

func TestService_Read_Video_Success(t *testing.T) {
	// TODO: реализовать юнит-тест, а пока есть только интеграционный тест с реальным вызовом ffmpeg
	// TODO: обернуть exec.Command("ffmpeg", ...) в Runner, чтобы инкапсулировать cmd.Start() & cmd.Wait() и замокать в тестах
}
