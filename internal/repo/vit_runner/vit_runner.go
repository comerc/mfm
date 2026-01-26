package vitrunner

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/comerc/nsfw-mod/internal/domain"
	"github.com/yalue/onnxruntime_go"
)

// Repo - репозиторий для запуска ViT (Vision Transformer) модели
type Repo struct {
	session *onnxruntime_go.DynamicAdvancedSession
}

// New создает новый экземпляр репозитория ViT
func New() *Repo {
	// Инициализация среды (идемпотентно)
	_ = onnxruntime_go.InitializeEnvironment()

	// Путь к модели
	modelPath := filepath.Join("assets", "vit_nsfw.onnx")

	// Проверяем существование модели
	if _, err := os.Stat(modelPath); os.IsNotExist(err) {
		// Try parent directory (for tests)
		parentPath := filepath.Join("..", "assets", "vit_nsfw.onnx")
		if _, err := os.Stat(parentPath); err == nil {
			modelPath = parentPath
		} else {
			fmt.Printf("ViTRepo: Model not found at %s or %s, using mock.\n", modelPath, parentPath)
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
func (r *Repo) Infer(frames [][]byte) (domain.ModerationResult, error) {
	if len(frames) == 0 || r.session == nil {
		return domain.ModerationResult{}, nil
	}

	batchSize := len(frames)
	// ViT ожидает формат NCHW: [batch, channels, height, width]
	inputShape := onnxruntime_go.NewShape(int64(batchSize), 3, 224, 224)

	// Подготавливаем данные
	// Важно: ViT требует нормализации ImageNet и перестановки осей
	inputData := make([]float32, batchSize*3*224*224)

	for i, frame := range frames {
		if len(frame) != 224*224*3 {
			return domain.ModerationResult{}, errors.New("invalid frame size")
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

	// Результат содержит два значения: [normal_prob, nsfw_prob]
	// Индекс 0 - вероятность нормального контента, индекс 1 - вероятность NSFW
	if len(outputData) < 2 {
		return domain.ModerationResult{}, errors.New("unexpected output length from model")
	}

	_ = outputData[0] // normal probability (not used directly)
	nsfwProb := outputData[1]

	// Определяем, является ли контент NSFW на основе вероятности NSFW (>0.5)
	isNSFW := nsfwProb > 0.5

	return domain.ModerationResult{
		IsNSFW: isNSFW,
		Score:  nsfwProb, // Используем вероятность NSFW как основной Score, как в эталонной реализации
	}, nil
}
