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

func TestMain(m *testing.M) {
	// Поднимаемся на уровень выше (из /test в корень)
	_ = os.Chdir("..")

	os.Exit(m.Run())
}

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

	// Инициализируем ONNX Runtime
	onnxinit.Initialize()

	mediaReader := mediareader.New()
	modelRunner := openrunner.New()
	service := moderation.New(tempDir, mediaReader, modelRunner)
	scores, err := service.Moderate([]string{testFile})
	score := scores[0]

	if err != nil {
		t.Error(err)
	}

	// Simple red image should not be NSFW, so score should be low
	if score > 0.5 {
		t.Errorf("expected score <= 0.5 for simple red image, got %f", score)
	}
}

func TestModerationService_Moderate_OpenRunner_RealAssets(t *testing.T) {
	// Инициализируем ONNX Runtime
	onnxinit.Initialize()

	modelRunner := openrunner.New()

	// Директория не важна для этого теста, так как передаем полный путь
	mediaReader := mediareader.New()
	service := moderation.New(".", mediaReader, modelRunner)

	assetsDir := "assets"
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
			scores, err := service.Moderate([]string{filePath})
			score := scores[0]

			if err != nil {
				t.Error(err)
			}

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
	_, err := service.Moderate([]string{nonExistentFile})

	if err == nil || !strings.Contains(err.Error(), "file not found:") {
		t.Error("invalid error")
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

	// Инициализируем ONNX Runtime
	onnxinit.Initialize()

	mediaReader := mediareader.New()
	vitRunner := vitrunner.New()
	service := moderation.New(tempDir, mediaReader, vitRunner)
	scores, err := service.Moderate([]string{testFile})
	score := scores[0]

	if err != nil {
		t.Error(err)
	}

	if score > 0.5 {
		t.Errorf("expected score <= 0.5 for simple red image, got %f", score)
	}

	t.Logf("ViT Result - IsNSFW: %v, Score: %.4f", score > 0.5, score)
}

func TestModerationService_Moderate_ViTRunner_RealAssets(t *testing.T) {
	// Инициализируем ONNX Runtime
	onnxinit.Initialize()

	vitRunner := vitrunner.New()

	// Директория не важна для этого теста, так как передаем полный путь
	_ = openrunner.New() // Ensure compile
	mediaReader := mediareader.New()
	service := moderation.New(".", mediaReader, vitRunner)

	assetsDir := "assets"
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
			scores, err := service.Moderate([]string{fullPath})
			score := scores[0]

			if err != nil {
				t.Error(err)
			}

			isNSFW := score > 0.5 // Предполагаем порог 0.5 для определения NSFW
			t.Logf("Image: %s, IsNSFW: %v, Score: %.4f", f.Name(), isNSFW, score)
		})
	}
}

func TestModerationService_Moderate_OpenRunner_BatchProcessing(t *testing.T) {
	// Initialize ONNX Runtime
	onnxinit.Initialize()

	mediaReader := mediareader.New()
	modelRunner := openrunner.New()
	service := moderation.New(".", mediaReader, modelRunner)

	// Prepare paths for the 7 images
	imagePaths := make([]string, 7)
	for i := range imagePaths {
		imagePaths[i] = filepath.Join("assets", fmt.Sprintf("%d.png", i+1))
	}

	// Verify all files exist
	for _, path := range imagePaths {
		if _, err := os.Stat(path); os.IsNotExist(err) {
			t.Fatalf("Required test image does not exist: %s", path)
		}
	}

	// Test batch processing with all 7 images
	scores, err := service.Moderate(imagePaths)
	if err != nil {
		t.Error(err)
	}

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
	onnxinit.Initialize()

	mediaReader := mediareader.New()
	vitRunner := vitrunner.New()
	service := moderation.New(".", mediaReader, vitRunner)

	// Prepare paths for the 7 images
	imagePaths := make([]string, 7)
	for i := range imagePaths {
		imagePaths[i] = filepath.Join("assets", fmt.Sprintf("%d.png", i+1))
	}

	// Verify all files exist
	for _, path := range imagePaths {
		if _, err := os.Stat(path); os.IsNotExist(err) {
			t.Fatalf("Required test image does not exist: %s", path)
		}
	}

	// Test batch processing with all 7 images
	scores, err := service.Moderate(imagePaths)
	if err != nil {
		t.Error(err)
	}

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

func BenchmarkModerationService_Moderate_OpenRunner_BatchProcessing120(b *testing.B) {
	// Initialize ONNX Runtime
	onnxinit.Initialize()

	mediaReader := mediareader.New()
	modelRunner := openrunner.New()
	service := moderation.New(".", mediaReader, modelRunner)

	// Prepare paths for 120 images by repeating the 7 available images
	imagePaths := make([]string, 120)
	for i := 0; i < 120; i++ {
		imgNum := (i % 7) + 1 // Cycle through images 1-7
		imagePaths[i] = filepath.Join("assets", fmt.Sprintf("%d.png", imgNum))
	}

	// Verify all files exist
	for _, path := range imagePaths {
		if _, err := os.Stat(path); os.IsNotExist(err) {
			b.Fatalf("Required test image does not exist: %s", path)
		}
	}

	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		scores, err := service.Moderate(imagePaths)
		if err != nil {
			b.Error(err)
		}
		if len(scores) != len(imagePaths) {
			b.Fatalf("Expected %d scores, got %d", len(imagePaths), len(scores))
		}
	}
}

func BenchmarkModerationService_Moderate_ViTRunner_BatchProcessing120(b *testing.B) {
	// Initialize ONNX Runtime
	onnxinit.Initialize()

	mediaReader := mediareader.New()
	vitRunner := vitrunner.New()
	service := moderation.New(".", mediaReader, vitRunner)

	// Prepare paths for 120 images by repeating the 7 available images
	imagePaths := make([]string, 120)
	for i := 0; i < 120; i++ {
		imgNum := (i % 7) + 1 // Cycle through images 1-7
		imagePaths[i] = filepath.Join("assets", fmt.Sprintf("%d.png", imgNum))
	}

	// Verify all files exist
	for _, path := range imagePaths {
		if _, err := os.Stat(path); os.IsNotExist(err) {
			b.Fatalf("Required test image does not exist: %s", path)
		}
	}

	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		scores, err := service.Moderate(imagePaths)
		if err != nil {
			b.Error(err)
		}
		if len(scores) != len(imagePaths) {
			b.Fatalf("Expected %d scores, got %d", len(imagePaths), len(scores))
		}
	}
}

func TestModerationService_Moderate_OpenRunner_RealVideoAssets(t *testing.T) {
	// Инициализируем ONNX Runtime
	onnxinit.Initialize()

	modelRunner := openrunner.New()

	// Директория не важна для этого теста, так как передаем полный путь
	mediaReader := mediareader.New()
	service := moderation.New(".", mediaReader, modelRunner)

	videoFiles := []string{"assets/video1.mp4", "assets/video2.mp4"}

	for _, videoPath := range videoFiles {
		// Проверяем существование видеофайла
		if _, err := os.Stat(videoPath); os.IsNotExist(err) {
			t.Logf("Video file not found: %s, skipping", videoPath)
			continue
		}

		t.Run(filepath.Base(videoPath), func(t *testing.T) {
			scores, err := service.Moderate([]string{videoPath})
			score := scores[0]
			if err != nil {
				t.Error(err)
			}

			isNSFW := score > 0.5 // Предполагаем порог 0.5 для определения NSFW
			t.Logf("Video: %s, IsNSFW: %v, Score: %.4f", filepath.Base(videoPath), isNSFW, score)
		})
	}
}

func TestModerationService_Moderate_ViTRunner_RealVideoAssets(t *testing.T) {
	// Инициализируем ONNX Runtime
	onnxinit.Initialize()

	vitRunner := vitrunner.New()

	// Директория не важна для этого теста, так как передаем полный путь
	_ = openrunner.New() // Ensure compile
	mediaReader := mediareader.New()
	service := moderation.New(".", mediaReader, vitRunner)

	videoFiles := []string{"assets/video1.mp4", "assets/video2.mp4"}

	for _, videoPath := range videoFiles {
		// Проверяем существование видеофайла
		if _, err := os.Stat(videoPath); os.IsNotExist(err) {
			t.Logf("Video file not found: %s, skipping", videoPath)
			continue
		}

		t.Run(filepath.Base(videoPath), func(t *testing.T) {
			scores, err := service.Moderate([]string{videoPath})
			score := scores[0]
			if err != nil {
				t.Error(err)
			}

			isNSFW := score > 0.5 // Предполагаем порог 0.5 для определения NSFW
			t.Logf("Video: %s, IsNSFW: %v, Score: %.4f", filepath.Base(videoPath), isNSFW, score)
		})
	}
}
