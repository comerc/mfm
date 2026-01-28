package moderation

import (
	"log/slog"
	"testing"

	"github.com/comerc/nsfw-mod/internal/service/moderation/mocks"
	"github.com/stretchr/testify/assert"
)

func TestModerate_Success(t *testing.T) {
	t.Parallel()

	// Таблица тестов
	tests := []struct {
		name        string
		inputFiles  []string
		frameData   [][]byte
		inferResult []float32
		expected    []float32
		setupMocks  func() (*Service, *mocks.MediaReader, *mocks.ModelRunner)
	}{
		{
			name:        "single_image_with_low_nsfw_score",
			inputFiles:  []string{"test_image.jpg"},
			frameData:   [][]byte{{1, 2, 3}}, // Только 1 фрейм для изображения
			inferResult: []float32{0.2},
			expected:    []float32{0.2},
			setupMocks: func() (*Service, *mocks.MediaReader, *mocks.ModelRunner) {
				mockMediaReader := mocks.NewMediaReader(t)
				mockModelRunner := mocks.NewModelRunner(t)

				service := New(mockMediaReader, mockModelRunner)

				mockMediaReader.EXPECT().Read("test_image.jpg").Return([][]byte{{1, 2, 3}}, nil)
				mockModelRunner.EXPECT().Infer([][]byte{{1, 2, 3}}).Return([]float32{0.2}, nil)

				return service, mockMediaReader, mockModelRunner
			},
		},
		{
			name:        "single_image_with_high_nsfw_score",
			inputFiles:  []string{"nsfw_image.jpg"},
			frameData:   [][]byte{{10, 20, 30}},
			inferResult: []float32{0.8},
			expected:    []float32{0.8},
			setupMocks: func() (*Service, *mocks.MediaReader, *mocks.ModelRunner) {
				mockMediaReader := mocks.NewMediaReader(t)
				mockModelRunner := mocks.NewModelRunner(t)

				service := New(mockMediaReader, mockModelRunner)

				mockMediaReader.EXPECT().Read("nsfw_image.jpg").Return([][]byte{{10, 20, 30}}, nil)
				mockModelRunner.EXPECT().Infer([][]byte{{10, 20, 30}}).Return([]float32{0.8}, nil)

				return service, mockMediaReader, mockModelRunner
			},
		},
		{
			name:        "multiple_images",
			inputFiles:  []string{"image1.jpg", "image2.jpg"},
			frameData:   [][]byte{}, // будет заполнено в setup
			inferResult: []float32{0.2, 0.7},
			expected:    []float32{0.2, 0.7},
			setupMocks: func() (*Service, *mocks.MediaReader, *mocks.ModelRunner) {
				mockMediaReader := mocks.NewMediaReader(t)
				mockModelRunner := mocks.NewModelRunner(t)

				service := New(mockMediaReader, mockModelRunner)

				frameData1 := [][]byte{{1, 2, 3}}
				frameData2 := [][]byte{{4, 5, 6}}

				mockMediaReader.EXPECT().Read("image1.jpg").Return(frameData1, nil)
				mockMediaReader.EXPECT().Read("image2.jpg").Return(frameData2, nil)
				mockModelRunner.EXPECT().Infer(append(frameData1, frameData2...)).Return([]float32{0.2, 0.7}, nil)

				return service, mockMediaReader, mockModelRunner
			},
		},
		{
			name:        "video_with_multiple_frames",
			inputFiles:  []string{"test_video.mp4"},
			frameData:   [][]byte{{1, 2, 3}, {4, 5, 6}, {7, 8, 9}}, // 3 фрейма для видео
			inferResult: []float32{0.2, 0.8, 0.5},                  // Максимальный результат должен быть 0.8
			expected:    []float32{0.8},                            // Берется максимальное значение из всех фреймов
			setupMocks: func() (*Service, *mocks.MediaReader, *mocks.ModelRunner) {
				mockMediaReader := mocks.NewMediaReader(t)
				mockModelRunner := mocks.NewModelRunner(t)

				service := New(mockMediaReader, mockModelRunner)

				videoFrameData := [][]byte{{1, 2, 3}, {4, 5, 6}, {7, 8, 9}}

				mockMediaReader.EXPECT().Read("test_video.mp4").Return(videoFrameData, nil)
				mockModelRunner.EXPECT().Infer(videoFrameData).Return([]float32{0.2, 0.8, 0.5}, nil)

				return service, mockMediaReader, mockModelRunner
			},
		},
		{
			name:        "multiple_videos",
			inputFiles:  []string{"video1.mp4", "video2.mp4"},
			frameData:   [][]byte{},                    // будет заполнено в setup
			inferResult: []float32{0.2, 0.8, 0.3, 0.6}, // [0.2, 0.8] -> max=0.8 и [0.3, 0.6] -> max=0.6
			expected:    []float32{0.8, 0.6},           // Максимальные значения для каждого видео
			setupMocks: func() (*Service, *mocks.MediaReader, *mocks.ModelRunner) {
				mockMediaReader := mocks.NewMediaReader(t)
				mockModelRunner := mocks.NewModelRunner(t)

				service := New(mockMediaReader, mockModelRunner)

				video1FrameData := [][]byte{{1, 2, 3}, {4, 5, 6}}    // 2 фрейма
				video2FrameData := [][]byte{{7, 8, 9}, {10, 11, 12}} // 2 фрейма
				allFrameData := append([][]byte{}, video1FrameData...)
				allFrameData = append(allFrameData, video2FrameData...)

				mockMediaReader.EXPECT().Read("video1.mp4").Return(video1FrameData, nil)
				mockMediaReader.EXPECT().Read("video2.mp4").Return(video2FrameData, nil)
				mockModelRunner.EXPECT().Infer(allFrameData).Return([]float32{0.2, 0.8, 0.3, 0.6}, nil)

				return service, mockMediaReader, mockModelRunner
			},
		},
		{
			name:        "mixed_files_image_and_video",
			inputFiles:  []string{"image.jpg", "video.mp4"},
			frameData:   [][]byte{},               // будет заполнено в setup
			inferResult: []float32{0.1, 0.3, 0.9}, // изображение=0.1, видео=[0.3, 0.9]->max=0.9
			expected:    []float32{0.1, 0.9},      // результаты: изображение=0.1, видео=0.9
			setupMocks: func() (*Service, *mocks.MediaReader, *mocks.ModelRunner) {
				mockMediaReader := mocks.NewMediaReader(t)
				mockModelRunner := mocks.NewModelRunner(t)

				service := New(mockMediaReader, mockModelRunner)

				imageFrameData := [][]byte{{1, 2, 3}}            // 1 фрейм для изображения
				videoFrameData := [][]byte{{4, 5, 6}, {7, 8, 9}} // 2 фрейма для видео
				allFrameData := append([][]byte{}, imageFrameData...)
				allFrameData = append(allFrameData, videoFrameData...)

				mockMediaReader.EXPECT().Read("image.jpg").Return(imageFrameData, nil)
				mockMediaReader.EXPECT().Read("video.mp4").Return(videoFrameData, nil)
				mockModelRunner.EXPECT().Infer(allFrameData).Return([]float32{0.1, 0.3, 0.9}, nil)

				return service, mockMediaReader, mockModelRunner
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Устанавливаем моки
			service, mockMediaReader, mockModelRunner := tt.setupMocks()

			// Вызываем тестируемый метод
			result, err := service.Moderate(tt.inputFiles)

			// Проверяем результат
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, result)

			// Проверяем, что вызовы были выполнены (автоматически проверяется через Cleanup в NewMediaReader/NewModelRunner)
			_ = mockMediaReader
			_ = mockModelRunner
		})
	}
}

func TestModerate_Error_ReadingFile(t *testing.T) {
	t.Parallel()

	// Подготовка данных
	file := "nonexistent.jpg"

	// Устанавливаем моки
	mockMediaReader := mocks.NewMediaReader(t)
	mockModelRunner := mocks.NewModelRunner(t)

	// Создаем сервис через конструктор
	service := New(mockMediaReader, mockModelRunner)

	// Устанавливаем ожидания для мока с использованием .EXPECT()
	mockMediaReader.EXPECT().Read(file).Return(([][]byte)(nil), assert.AnError)

	// Вызываем тестируемый метод
	result, err := service.Moderate([]string{file})

	// Проверяем результат
	assert.Nil(t, result)
	assert.Error(t, err)
}

func TestModerate_Error_Inference(t *testing.T) {
	t.Parallel()

	// Подготовка данных
	file := "image.jpg"
	frameData := [][]byte{{1, 2, 3}}

	// Устанавливаем моки
	mockMediaReader := mocks.NewMediaReader(t)
	mockModelRunner := mocks.NewModelRunner(t)

	// Создаем сервис через конструктор
	service := New(mockMediaReader, mockModelRunner)

	// Устанавливаем ожидания для мока с использованием .EXPECT()
	mockMediaReader.EXPECT().Read(file).Return(frameData, nil)
	mockModelRunner.EXPECT().Infer(frameData).Return(([]float32)(nil), assert.AnError)

	// Вызываем тестируемый метод
	result, err := service.Moderate([]string{file})

	// Проверяем результат
	assert.Nil(t, result)
	assert.Error(t, err)
}

func Test_getMaxScore(t *testing.T) {
	t.Parallel()

	// Таблица тестов для вспомогательной функции
	tests := []struct {
		name     string
		scores   []float32
		expected float32
	}{
		{
			name:     "all_positive_scores",
			scores:   []float32{0.1, 0.5, 0.3, 0.9, 0.2},
			expected: 0.9,
		},
		{
			name:     "single_score",
			scores:   []float32{0.7},
			expected: 0.7,
		},
		{
			name:     "scores_with_zeros",
			scores:   []float32{0.0, 0.0, 0.0},
			expected: 0.0,
		},
		{
			name:     "mixed_positive_and_negative_scores",
			scores:   []float32{-0.5, 0.2, -0.1, 0.8},
			expected: 0.8,
		},
		{
			name:     "empty_slice",
			scores:   []float32{},
			expected: 0.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Создаем сервис через конструктор с пустыми моками
			service := &Service{
				log:         slog.Default(),
				mediaReader: nil,
				modelRunner: nil,
			}

			result := service.getMaxScore(tt.scores)
			assert.Equal(t, tt.expected, result)
		})
	}
}
