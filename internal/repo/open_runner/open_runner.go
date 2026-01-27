package openrunner

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/yalue/onnxruntime_go"
)

// Repo - репозиторий для запуска ML моделей
type Repo struct {
	session *onnxruntime_go.DynamicAdvancedSession
}

// New создает новый экземпляр репозитория моделей
func New() *Repo {

	// Путь к модели
	modelPath := filepath.Join("assets", "opennsfw2.onnx")

	// Проверяем существование модели
	if _, err := os.Stat(modelPath); os.IsNotExist(err) {
		// Модель не найдена, используем mock
		fmt.Printf("openrunner: Model not found at %s, using mock.\n", modelPath)
		return &Repo{
			session: nil,
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
func (r *Repo) Infer(frames [][]byte) ([]float32, error) {
	if len(frames) == 0 || r.session == nil {
		return []float32{}, nil
	}

	batchSize := len(frames)
	inputShape := onnxruntime_go.NewShape(int64(batchSize), 224, 224, 3)

	// Подготавливаем данные
	inputData := make([]float32, batchSize*224*224*3)
	for i, frame := range frames {
		if len(frame) != 224*224*3 {
			return []float32{}, errors.New("invalid frame size")
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
		return []float32{}, fmt.Errorf("failed to create input tensor: %v", err)
	}
	defer func() {
		_ = inputTensor.Destroy()
	}()

	// Выполняем инференс
	outputs := []onnxruntime_go.Value{nil}
	err = r.session.Run([]onnxruntime_go.Value{inputTensor}, outputs)
	if err != nil {
		return []float32{}, fmt.Errorf("inference failed: %v", err)
	}
	defer func() {
		_ = outputs[0].Destroy()
	}()

	// Получаем результат
	outputData := outputs[0].(*onnxruntime_go.Tensor[float32]).GetData()

	// Возвращаем все оценки для каждого кадра
	scores := make([]float32, batchSize)
	for i := 0; i < batchSize; i++ {
		score := outputData[i]
		// Ограничиваем значение в диапазоне [0, 1] для безопасности
		if score < 0.0 {
			score = 0.0
		} else if score > 1.0 {
			score = 1.0
		}
		scores[i] = score
	}

	return scores, nil
}
