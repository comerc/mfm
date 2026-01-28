package mediareader

import (
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRead_Success(t *testing.T) {
	t.Parallel()
	// Arrange
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test.png")

	img := image.NewRGBA(image.Rect(0, 0, 1, 1))
	img.Set(0, 0, color.RGBA{255, 0, 0, 255}) // Red pixel

	file, err := os.Create(testFile)
	require.NoError(t, err, "should create test file successfully")
	defer file.Close()

	err = png.Encode(file, img)
	require.NoError(t, err, "should encode PNG successfully")

	repo := New()

	// Act
	frames, err := repo.Read(testFile)

	// Assert
	assert.NoError(t, err, "expected no error when reading valid image")
	assert.Len(t, frames, 1, "expected exactly 1 frame from a single image")

	frame := frames[0]
	expectedSize := 224 * 224 * 3
	assert.Len(t, frame, expectedSize, "frame size should match expected dimensions")
}

func TestRead_FileNotFound(t *testing.T) {
	t.Parallel()
	// Arrange
	tempDir := t.TempDir()
	repo := New()

	nonExistentFile := filepath.Join(tempDir, "nonexistent.png")

	// Act
	_, err := repo.Read(nonExistentFile)

	// Assert
	assert.Error(t, err, "expected error for non-existent file")
	assert.Contains(t, err.Error(), "file not found", "error message should contain 'file not found'")
}

func TestRead_UnsupportedType(t *testing.T) {
	t.Parallel()
	// Arrange
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test.txt")

	err := os.WriteFile(testFile, []byte("this is not an image"), 0600)
	require.NoError(t, err, "should create test file successfully")

	repo := New()

	// Act
	_, err = repo.Read(testFile)

	// Assert
	assert.Error(t, err, "expected error for unsupported file type")
	assert.Contains(t, err.Error(), "unsupported file type", "error message should contain 'unsupported file type'")
}

func TestService_Read_Video_Success(t *testing.T) {
	// TODO: реализовать юнит-тест, а пока есть только интеграционный тест с реальным вызовом ffmpeg
	// TODO: обернуть exec.Command("ffmpeg", ...) в Runner, чтобы инкапсулировать cmd.Start() & cmd.Wait() и замокать в тестах
}
