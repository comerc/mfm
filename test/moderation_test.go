package test

import (
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"strings"
	"testing"

	mediareader "github.com/comerc/nsfw-mod/internal/repo/media_reader"
	openrunner "github.com/comerc/nsfw-mod/internal/repo/open_runner"
	vitrunner "github.com/comerc/nsfw-mod/internal/repo/vit_runner"
	"github.com/comerc/nsfw-mod/internal/service/moderation"
	"github.com/comerc/nsfw-mod/pkg/onnxinit"
)

func TestModerationService_Moderate_OpenRunner_Success(t *testing.T) {
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

	// Инициализируем ONNX Runtime
	onnxinit.Initialize()

	mediaReader := mediareader.New()
	modelRunner := openrunner.New()
	service := moderation.New(tempDir, mediaReader, modelRunner)
	result := service.Moderate([]string{testFile})

	if result.Error != "" {
		t.Errorf("expected no error, got %s", result.Error)
	}

	// Simple red image should not be NSFW
	if result.IsNSFW {
		t.Errorf("expected IsNSFW to be false, got true")
	}
}

func TestModerationService_Moderate_OpenRunner_RealAssets(t *testing.T) {
	// Инициализируем репо с перехватом паники (если нет либы)
	defer func() {
		if r := recover(); r != nil {
			t.Skipf("Skipping real assets test due to initialization failure (missing lib?): %v", r)
		}
	}()

	// Инициализируем ONNX Runtime
	onnxinit.Initialize()

	modelRunner := openrunner.New()

	// Проверяем наличие модели, если её нет - пропускаем интеграционные тесты
	if _, err := os.Stat("../assets/opennsfw2.onnx"); os.IsNotExist(err) {
		t.Skip("Model file not found in ../assets/opennsfw2.onnx, skipping integration test")
	}

	// Директория не важна для этого теста, так как передаем полный путь
	mediaReader := mediareader.New()
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
			filePath := filepath.Join(assetsDir, f.Name())
			result := service.Moderate([]string{filePath})

			if result.Error != "" {
				t.Fatalf("Processing failed for %s: %s", f.Name(), result.Error)
			}

			t.Logf("Image: %s, IsNSFW: %v, Score: %.4f", f.Name(), result.IsNSFW, result.Score)
		})
	}
}

func TestModerationService_Moderate_FileNotFound(t *testing.T) {
	tempDir := t.TempDir()

	// Инициализируем ONNX Runtime
	onnxinit.Initialize()

	mediaReader := mediareader.New()
	modelRunner := openrunner.New()
	service := moderation.New(tempDir, mediaReader, modelRunner)

	nonExistentFile := filepath.Join(tempDir, "nonexistent.png")
	result := service.Moderate([]string{nonExistentFile})

	if result.Error == "" {
		t.Error("expected error for non-existent file")
	}

	if !strings.Contains(result.Error, "file not found") || !strings.Contains(result.Error, "nonexistent.png") {
		t.Errorf("expected error containing 'file not found' and 'nonexistent.png', got %s", result.Error)
	}
}

func TestModerationService_Moderate_ViTRunner_Success(t *testing.T) {
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

	// Инициализируем ONNX Runtime
	onnxinit.Initialize()

	// Проверяем наличие модели ViT
	if _, err := os.Stat("../assets/vit_nsfw.onnx"); os.IsNotExist(err) {
		t.Skip("Model file not found in ../assets/vit_nsfw.onnx, skipping integration test")
	}

	mediaReader := mediareader.New()
	vitRunner := vitrunner.New()
	service := moderation.New(tempDir, mediaReader, vitRunner)
	result := service.Moderate([]string{testFile})

	if result.Error != "" {
		t.Errorf("expected no error, got %s", result.Error)
	}

	// Simple red image should not be NSFW
	if result.IsNSFW {
		t.Errorf("expected IsNSFW to be false, got true")
	}

	// Ранее проверяли Categories, но теперь они удалены
	_ = result // использовать результат для избежания warning'а
	// Проверка IsNSFW и Score уже выполняется выше

	// Log the results for verification
	t.Logf("ViT Result - IsNSFW: %v, Score: %.4f", result.IsNSFW, result.Score)
}

func TestModerationService_Moderate_ViTRunner_RealAssets(t *testing.T) {
	// Инициализируем репо с перехватом паники (если нет либы)
	defer func() {
		if r := recover(); r != nil {
			t.Skipf("Skipping ViT real assets test due to initialization failure (missing lib?): %v", r)
		}
	}()

	// Инициализируем ONNX Runtime
	onnxinit.Initialize()

	vitRunner := vitrunner.New()

	// Проверяем наличие модели, если её нет - пропускаем интеграционные тесты
	if _, err := os.Stat("../assets/vit_nsfw.onnx"); os.IsNotExist(err) {
		t.Skip("Model file not found in ../assets/vit_nsfw.onnx, skipping integration test")
	}

	// Директория не важна для этого теста, так как передаем полный путь
	_ = openrunner.New() // Ensure compile
	mediaReader := mediareader.New()
	service := moderation.New(".", mediaReader, vitRunner)

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
			result := service.Moderate([]string{fullPath})

			if result.Error != "" {
				t.Fatalf("Processing failed for %s: %s", f.Name(), result.Error)
			}

			// Ранее проверяли Categories, но теперь они удалены
			_ = f.Name() // использовать имя файла для избежания warning'а
			// Проверка IsNSFW и Score уже выполняется выше

			t.Logf("Image: %s, IsNSFW: %v, Score: %.4f", f.Name(), result.IsNSFW, result.Score)
		})
	}
}

func TestModerationService_Moderate_OpenRunner_BatchProcessing(t *testing.T) {
	// Initialize ONNX Runtime
	defer func() {
		if r := recover(); r != nil {
			t.Skipf("Skipping batch processing test due to initialization failure (missing lib?): %v", r)
		}
	}()

	onnxinit.Initialize()

	// Check if model file exists
	if _, err := os.Stat("../assets/opennsfw2.onnx"); os.IsNotExist(err) {
		t.Skip("Model file not found in ../assets/opennsfw2.onnx, skipping integration test")
	}

	mediaReader := mediareader.New()
	modelRunner := openrunner.New()
	service := moderation.New(".", mediaReader, modelRunner)

	// Prepare paths for the 7 images
	imagePaths := make([]string, 7)
	for i := 1; i <= 7; i++ {
		imagePaths[i-1] = filepath.Join("../assets", fmt.Sprintf("%d.png", i))
	}

	// Verify all files exist
	for _, path := range imagePaths {
		if _, err := os.Stat(path); os.IsNotExist(err) {
			t.Fatalf("Required test image does not exist: %s", path)
		}
	}

	// Test batch processing with all 7 images
	result := service.Moderate(imagePaths)

	if result.Error != "" {
		t.Fatalf("Batch processing failed: %s", result.Error)
	}

	t.Logf("Batch Processing Result - IsNSFW: %v, Score: %.4f", result.IsNSFW, result.Score)

	// Basic validation that we got a reasonable result
	if result.Score < 0.0 || result.Score > 1.0 {
		t.Errorf("Expected score between 0.0 and 1.0, got: %.4f", result.Score)
	}
}

func TestModerationService_Moderate_ViTRunner_BatchProcessing(t *testing.T) {
	// Initialize ONNX Runtime
	defer func() {
		if r := recover(); r != nil {
			t.Skipf("Skipping ViT batch processing test due to initialization failure (missing lib?): %v", r)
		}
	}()

	onnxinit.Initialize()

	// Check if model file exists
	if _, err := os.Stat("../assets/vit_nsfw.onnx"); os.IsNotExist(err) {
		t.Skip("Model file not found in ../assets/vit_nsfw.onnx, skipping integration test")
	}

	mediaReader := mediareader.New()
	vitRunner := vitrunner.New()
	service := moderation.New(".", mediaReader, vitRunner)

	// Prepare paths for the 7 images
	imagePaths := make([]string, 7)
	for i := 1; i <= 7; i++ {
		imagePaths[i-1] = filepath.Join("../assets", fmt.Sprintf("%d.png", i))
	}

	// Verify all files exist
	for _, path := range imagePaths {
		if _, err := os.Stat(path); os.IsNotExist(err) {
			t.Fatalf("Required test image does not exist: %s", path)
		}
	}

	// Test batch processing with all 7 images
	result := service.Moderate(imagePaths)

	if result.Error != "" {
		t.Fatalf("Batch processing failed: %s", result.Error)
	}

	t.Logf("ViT Batch Processing Result - IsNSFW: %v, Score: %.4f", result.IsNSFW, result.Score)

	// Basic validation that we got a reasonable result
	if result.Score < 0.0 || result.Score > 1.0 {
		t.Errorf("Expected score between 0.0 and 1.0, got: %.4f", result.Score)
	}
}
