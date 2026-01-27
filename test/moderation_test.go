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
	scores := service.Moderate([]string{testFile})

	// Проверяем, что получен хотя бы один результат
	if len(scores) == 0 {
		t.Errorf("expected at least one score, got none")
	} else {
		score := scores[0]
		// Simple red image should not be NSFW, so score should be low
		if score > 0.5 {
			t.Errorf("expected score <= 0.5 for simple red image, got %f", score)
		}
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
			scores := service.Moderate([]string{filePath})

			if len(scores) == 0 {
				t.Fatalf("Processing failed for %s: no scores returned", f.Name())
			}

			score := scores[0]
			isNSFW := score > 0.5 // Предполагаем порог 0.5 для определения NSFW
			t.Logf("Image: %s, IsNSFW: %v, Score: %.4f", f.Name(), isNSFW, score)
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
	scores := service.Moderate([]string{nonExistentFile})

	// Ожидаем, что для несуществующего файла будет возвращена пустая оценка или нулевая оценка
	if len(scores) == 0 {
		t.Error("expected at least one score (even if zero) for non-existent file")
	} else if scores[0] != 0.0 {
		t.Errorf("expected score of 0.0 for non-existent file, got %f", scores[0])
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
	scores := service.Moderate([]string{testFile})

	// Проверяем, что получен хотя бы один результат
	if len(scores) == 0 {
		t.Errorf("expected at least one score, got none")
	} else {
		score := scores[0]
		// Simple red image should not be NSFW, so score should be low
		if score > 0.5 {
			t.Errorf("expected score <= 0.5 for simple red image, got %f", score)
		}
	}

	// Log the results for verification
	// Получаем оценку из результата
	score := scores[0]
	t.Logf("ViT Result - IsNSFW: %v, Score: %.4f", score > 0.5, score)
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
			scores := service.Moderate([]string{fullPath})

			if len(scores) == 0 {
				t.Fatalf("Processing failed for %s: no scores returned", f.Name())
			}

			score := scores[0]
			isNSFW := score > 0.5 // Предполагаем порог 0.5 для определения NSFW
			t.Logf("Image: %s, IsNSFW: %v, Score: %.4f", f.Name(), isNSFW, score)
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
	for i := range imagePaths {
		imagePaths[i] = filepath.Join("../assets", fmt.Sprintf("%d.png", i+1))
	}

	// Verify all files exist
	for _, path := range imagePaths {
		if _, err := os.Stat(path); os.IsNotExist(err) {
			t.Fatalf("Required test image does not exist: %s", path)
		}
	}

	// Test batch processing with all 7 images
	scores := service.Moderate(imagePaths)

	if len(scores) == 0 {
		t.Fatalf("Batch processing failed: no scores returned")
	}

	// Проверяем, что количество возвращенных оценок совпадает с количеством файлов
	if len(scores) != len(imagePaths) {
		t.Errorf("Expected %d scores, got %d", len(imagePaths), len(scores))
	}

	// Базовая проверка разумности результатов
	for i, score := range scores {
		if score < 0.0 || score > 1.0 {
			t.Errorf("Expected score[%d] between 0.0 and 1.0, got: %.4f", i, score)
		}
	}

	t.Logf("Batch Processing Result - Scores: %v", scores)
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
	for i := range imagePaths {
		imagePaths[i] = filepath.Join("../assets", fmt.Sprintf("%d.png", i+1))
	}

	// Verify all files exist
	for _, path := range imagePaths {
		if _, err := os.Stat(path); os.IsNotExist(err) {
			t.Fatalf("Required test image does not exist: %s", path)
		}
	}

	// Test batch processing with all 7 images
	scores := service.Moderate(imagePaths)

	if len(scores) == 0 {
		t.Fatalf("Batch processing failed: no scores returned")
	}

	// Проверяем, что количество возвращенных оценок совпадает с количеством файлов
	if len(scores) != len(imagePaths) {
		t.Errorf("Expected %d scores, got %d", len(imagePaths), len(scores))
	}

	// Базовая проверка разумности результатов
	for i, score := range scores {
		if score < 0.0 || score > 1.0 {
			t.Errorf("Expected score[%d] between 0.0 and 1.0, got: %.4f", i, score)
		}
	}

	t.Logf("ViT Batch Processing Result - Scores: %v", scores)
}
