package openrunner

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/comerc/nsfw-mod/internal/domain"
	"github.com/yalue/onnxruntime_go"
)

// Repo - репозиторий для запуска ML моделей
type Repo struct {
	session *onnxruntime_go.DynamicAdvancedSession
}

// New создает новый экземпляр репозитория моделей
func New() *Repo {
	// TODO: надо вынести эту инициализацию наружу, чтобы не зависеть от запуска конкретно этого ранера
	// Устанавливаем путь к shared library
	onnxruntime_go.SetSharedLibraryPath("/usr/local/lib/libonnxruntime.dylib")

	// Инициализируем ONNX Runtime
	// Игнорируем ошибку повторной инициализации, которая может возникнуть в тестах
	_ = onnxruntime_go.InitializeEnvironment()

	// Путь к модели
	modelPath := filepath.Join("assets", "opennsfw2.onnx")

	// Проверяем существование модели
	if _, err := os.Stat(modelPath); os.IsNotExist(err) {
		// Try parent directory (for tests)
		parentPath := filepath.Join("..", "assets", "opennsfw2.onnx")
		if _, err := os.Stat(parentPath); err == nil {
			modelPath = parentPath
		} else {
			// Модель не найдена, используем mock
			fmt.Printf("openrunner: Model not found at %s or %s, using mock.\n", modelPath, parentPath)
			return &Repo{
				session: nil,
			}
		}
	}

	// Загружаем модель
	session, err := onnxruntime_go.NewDynamicAdvancedSession(modelPath, []string{"input"}, []string{"get_item"}, nil)
	if err != nil {
		panic(fmt.Sprintf("Failed to create session: %v", err))
	}

	return &Repo{
		session: session,
	}
}

// Infer запускает inference на данных (пакет кадров)
func (r *Repo) Infer(frames [][]byte) (domain.ModerationResult, error) {
	if len(frames) == 0 || r.session == nil {
		return domain.ModerationResult{}, nil
	}

	batchSize := len(frames)
	inputShape := onnxruntime_go.NewShape(int64(batchSize), 224, 224, 3)

	// Подготавливаем данные
	inputData := make([]float32, batchSize*224*224*3)
	for i, frame := range frames {
		if len(frame) != 224*224*3 {
			return domain.ModerationResult{}, errors.New("invalid frame size")
		}

		// Preprocessing for OpenNSFW2: BGR, subtract mean
		// Actually best works with: BGR order, subtract mean [104, 117, 123]
		// This matches standard Caffe preprocessing which the model seems to still rely on despite docs saying otherwise or conversion quirks.
		for p := 0; p < 224*224; p++ {
			r := float32(frame[p*3])
			g := float32(frame[p*3+1])
			b := float32(frame[p*3+2])

			// BGR order and subtract mean
			inputData[i*224*224*3+p*3] = b - 104.0
			inputData[i*224*224*3+p*3+1] = g - 117.0
			inputData[i*224*224*3+p*3+2] = r - 123.0
		}
	}

	// Создаем тензор
	inputTensor, err := onnxruntime_go.NewTensor(inputShape, inputData)
	if err != nil {
		return domain.ModerationResult{}, fmt.Errorf("failed to create input tensor: %v", err)
	}
	defer inputTensor.Destroy()

	// Выполняем инференс
	outputs := []onnxruntime_go.Value{nil}
	err = r.session.Run([]onnxruntime_go.Value{inputTensor}, outputs)
	if err != nil {
		return domain.ModerationResult{}, fmt.Errorf("inference failed: %v", err)
	}
	defer outputs[0].Destroy()

	// Получаем результат
	outputData := outputs[0].(*onnxruntime_go.Tensor[float32]).GetData()

	var maxScore float32
	// Предполагаем, что output - вероятность NSFW для каждого кадра
	for i := 0; i < batchSize; i++ {
		score := outputData[i]
		if score > maxScore {
			maxScore = score
		}
		if score > 0.8 { // Поднимаем порог до 0.8 для большей уверенности
			return domain.ModerationResult{
				IsNSFW: true,
				Score:  score,
			}, nil
		}
	}

	return domain.ModerationResult{
		IsNSFW: false,
		Score:  maxScore,
	}, nil
}
