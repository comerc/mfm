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

	// Инициализируем ONNX Runtime
	onnxinit.Initialize()

	os.Exit(m.Run())
}

func TestModerationService_Moderate_FileNotFound(t *testing.T) {
	tempDir := t.TempDir()

	mediaReader := mediareader.New()
	modelRunner := openrunner.New()
	service := moderation.New(tempDir, mediaReader, modelRunner)

	nonExistentFile := filepath.Join(tempDir, "nonexistent.png")
	_, err := service.Moderate([]string{nonExistentFile})

	if err == nil || !strings.Contains(err.Error(), "file not found:") {
		t.Error("invalid error")
	}
}

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

	tests := []struct {
		name      string
		newRunner func() moderation.ModelRunner
	}{
		{
			name:      "OpenRunner",
			newRunner: func() moderation.ModelRunner { return openrunner.New() },
		},
		{
			name:      "ViTRunner",
			newRunner: func() moderation.ModelRunner { return vitrunner.New() },
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mediaReader := mediareader.New()
			modelRunner := tt.newRunner()
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

			if tt.name == "ViTRunner" {
				t.Logf("ViT Result - IsNSFW: %v, Score: %.4f", score > 0.5, score)
			}
		})
	}
}

func TestModerationService_Moderate_Integration(t *testing.T) {
	tests := []struct {
		name      string
		newRunner func() moderation.ModelRunner
	}{
		{
			name:      "OpenRunner",
			newRunner: func() moderation.ModelRunner { return openrunner.New() },
		},
		{
			name:      "ViTRunner",
			newRunner: func() moderation.ModelRunner { return vitrunner.New() },
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mediaReader := mediareader.New()
			modelRunner := tt.newRunner()
			service := moderation.New(".", mediaReader, modelRunner)

			// Собираем все файлы из assets (картинки и видео)
			assetsDir := "assets"
			entries, err := os.ReadDir(assetsDir)
			if err != nil {
				t.Fatalf("Failed to read assets directory: %v", err)
			}

			var filePaths []string
			for _, entry := range entries {
				if entry.IsDir() {
					continue
				}
				ext := strings.ToLower(filepath.Ext(entry.Name()))
				if ext == ".png" || ext == ".mp4" {
					filePaths = append(filePaths, filepath.Join(assetsDir, entry.Name()))
				}
			}

			if len(filePaths) == 0 {
				t.Skip("No image or video files found in assets directory")
			}

			// Test batch processing with all collected files
			scores, err := service.Moderate(filePaths)
			if err != nil {
				t.Error(err)
			}

			if len(scores) == 0 {
				t.Fatalf("Batch processing failed: no scores returned")
			}

			// Проверяем, что количество возвращенных оценок совпадает с количеством файлов
			if len(scores) != len(filePaths) {
				t.Errorf("Expected %d scores, got %d", len(filePaths), len(scores))
			}

			// Логируем результаты для каждого файла
			for i, filePath := range filePaths {
				isNSFW := scores[i] > 0.5
				ext := strings.ToLower(filepath.Ext(filePath))
				fileType := "image"
				if ext == ".mp4" {
					fileType = "video"
				}
				t.Logf("%s: %s, IsNSFW: %v, Score: %.4f", fileType, filepath.Base(filePath), isNSFW, scores[i])
			}
		})
	}
}

func BenchmarkModerate_BatchProcessing120(b *testing.B) {
	tests := []struct {
		name      string
		newRunner func() moderation.ModelRunner
	}{
		{
			name:      "OpenRunner",
			newRunner: func() moderation.ModelRunner { return openrunner.New() },
		},
		{
			name:      "ViTRunner",
			newRunner: func() moderation.ModelRunner { return vitrunner.New() },
		},
	}

	for _, tt := range tests {
		b.Run(tt.name, func(b *testing.B) {
			mediaReader := mediareader.New()
			modelRunner := tt.newRunner()
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
		})
	}
}
