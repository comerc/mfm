package vitrunner

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/yalue/onnxruntime_go"
)

// Repo - репозиторий для запуска ViT (Vision Transformer) модели
type Repo struct {
	session *onnxruntime_go.DynamicAdvancedSession
}

// New создает новый экземпляр репозитория ViT
func New() *Repo {
	// Путь к модели
	modelPath := filepath.Join("assets", "vit_nsfw.onnx")

	// Проверяем существование модели
	if _, err := os.Stat(modelPath); os.IsNotExist(err) {
		// Try parent directory (for tests)
		parentPath := filepath.Join("..", "assets", "vit_nsfw.onnx")
		if _, err := os.Stat(parentPath); err == nil {
			modelPath = parentPath
		} else {
			fmt.Printf("vitrunner: Model not found at %s or %s, using mock.\n", modelPath, parentPath)
			return &Repo{
				session: nil,
			}
		}
	}

	// Загружаем модель
	// Имена входа/выхода из документации: input="input", output="output"
	session, err := onnxruntime_go.NewDynamicAdvancedSession(modelPath, []string{"input"}, []string{"output"}, nil)
	if err != nil {
		panic(fmt.Sprintf("Failed to create session for ViT: %v", err))
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
	// ViT ожидает формат NCHW: [batch, channels, height, width]
	inputShape := onnxruntime_go.NewShape(int64(batchSize), 3, 224, 224)

	// Подготавливаем данные
	// Важно: ViT требует нормализации ImageNet и перестановки осей
	inputData := make([]float32, batchSize*3*224*224)

	for i, frame := range frames {
		if len(frame) != 224*224*3 {
			return []float32{}, errors.New("invalid frame size")
		}

		// Подготовка данных для ViT:
		// 1. Нормализация в диапазон [0, 1]: деление на 255.0
		// 2. Нормализация ImageNet: вычитание среднего и деление на стандартное отклонение
		// 3. Перестановка осей: HWC -> CHW (Height-Width-Channel to Channel-Height-Width)

		// Определяем mean и std для нормализации ImageNet
		mean := []float32{0.485, 0.456, 0.406}
		std := []float32{0.229, 0.224, 0.225}

		// Входной frame имеет формат HWC (Height, Width, Channel) в RGB порядке
		// Нужно преобразовать в CHW (Channel, Height, Width) формат
		for h := 0; h < 224; h++ {
			for w := 0; w < 224; w++ {
				for c := 0; c < 3; c++ {
					// Индекс в исходном кадре (HWC формат)
					srcIdx := h*224*3 + w*3 + c

					// Получаем значение и нормализуем в диапазон [0, 1]
					val := float32(frame[srcIdx]) / 255.0

					// Применяем нормализацию ImageNet
					normalizedVal := (val - mean[c]) / std[c]

					// Сохраняем в CHW формате (каналы первыми)
					// dstIdx = batch_idx * channels * height * width + channel * height * width + height_pos * width + width_pos
					dstIdx := i*3*224*224 + c*224*224 + h*224 + w
					inputData[dstIdx] = normalizedVal
				}
			}
		}
	}

	// Создаем тензор
	inputTensor, err := onnxruntime_go.NewTensor(inputShape, inputData)
	if err != nil {
		return []float32{}, fmt.Errorf("failed to create input tensor: %v", err)
	}
	defer inputTensor.Destroy()

	// Выполняем инференс
	outputs := []onnxruntime_go.Value{nil}
	err = r.session.Run([]onnxruntime_go.Value{inputTensor}, outputs)
	if err != nil {
		return []float32{}, fmt.Errorf("inference failed: %v", err)
	}
	defer outputs[0].Destroy()

	// Получаем результат
	outputData := outputs[0].(*onnxruntime_go.Tensor[float32]).GetData()

	// Результат содержит пары значений: [normal_prob, nsfw_prob, normal_prob, nsfw_prob, ...] для каждого кадра в пакете
	expectedLength := batchSize * 2 // 2 значения для каждого кадра
	if len(outputData) != expectedLength {
		return []float32{}, fmt.Errorf("unexpected output length from model: got %d, expected %d", len(outputData), expectedLength)
	}

	// Возвращаем оценки NSFW для каждого кадра
	scores := make([]float32, batchSize)
	for i := 0; i < batchSize; i++ {
		nsfwProb := outputData[i*2+1] // индекс 1, 3, 5, ... - вероятности NSFW для каждого кадра
		// Ограничиваем значение в диапазоне [0, 1] для безопасности
		if nsfwProb < 0.0 {
			nsfwProb = 0.0
		} else if nsfwProb > 1.0 {
			nsfwProb = 1.0
		}
		scores[i] = nsfwProb
	}

	return scores, nil
}
