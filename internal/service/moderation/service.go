package moderation

type MediaReader interface {
	Read(filePath string) ([][]byte, error)
}

type ModelRunner interface {
	Infer(data [][]byte) ([]float32, error)
}

// Service implements Service
type Service struct {
	uploadDir   string
	mediaReader MediaReader
	modelRunner ModelRunner
}

// New creates a new instance of ModerationService
func New(uploadDir string, mediaReader MediaReader, modelRunner ModelRunner) *Service {
	return &Service{
		uploadDir:   uploadDir,
		mediaReader: mediaReader,
		modelRunner: modelRunner,
	}
}

func (s *Service) Moderate(filePaths []string) ([]float32, error) {
	// Будем хранить информацию о том, сколько фреймов приходится на каждый файл
	var data [][]byte     // Все фреймы для инференса
	var frameCounts []int // Количество фреймов для каждого файла

	for _, filePath := range filePaths {
		fileFrames, err := s.mediaReader.Read(filePath)
		if err != nil {
			return nil, err
		}

		frameCounts = append(frameCounts, len(fileFrames))

		// Добавляем все фреймы этого файла в общий массив
		for _, frame := range fileFrames {
			data = append(data, frame)
		}
	}

	// Выполняем инференс для всех фреймов
	allScores, err := s.modelRunner.Infer(data)
	if err != nil {
		return nil, err
	}

	// Объединяем результаты для фреймов одного видео
	var results []float32
	frameIndex := 0

	for _, count := range frameCounts {
		if count == 1 {
			// Это изображение - просто добавляем один результат
			results = append(results, allScores[frameIndex])
			frameIndex++
		} else {
			// Это видео - берем максимальный результат среди всех фреймов
			videoScores := allScores[frameIndex : frameIndex+count]
			maxScore := s.getMaxScore(videoScores)
			results = append(results, maxScore)
			frameIndex += count
		}
	}

	return results, nil
}

func (s *Service) getMaxScore(scores []float32) float32 {
	var maxScore float32
	for _, score := range scores {
		if score > maxScore {
			maxScore = score
		}
	}
	return maxScore
}
