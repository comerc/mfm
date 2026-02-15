package moderation

import (
	"log/slog"
)

//go:generate mockery --name=mediaReader --exported
type mediaReader interface {
	Read(filePath string) ([][]byte, error)
}

//go:generate mockery --name=modelRunner --exported
type modelRunner interface {
	Infer(data [][]byte) ([]float32, error)
}

// Service implements Service
type Service struct {
	log          *slog.Logger
	mediaReader  mediaReader
	modelRunners []modelRunner
}

// New creates a new instance of ModerationService
func New(mediaReader mediaReader, modelRunners ...modelRunner) *Service {
	return &Service{
		log:          slog.With("module", "moderation"),
		mediaReader:  mediaReader,
		modelRunners: modelRunners,
	}
}

func (s *Service) Moderate(filePaths []string) ([]float32, error) {
	s.log.Info("Starting moderation", "file_count", len(filePaths), "file_paths", filePaths)

	// Будем хранить информацию о том, сколько фреймов приходится на каждый файл
	var data [][]byte     // Все фреймы для инференса
	var frameCounts []int // Количество фреймов для каждого файла

	for _, filePath := range filePaths {
		s.log.Debug("Reading media file", "file_path", filePath)

		fileFrames, err := s.mediaReader.Read(filePath)
		if err != nil {
			s.log.Error("Failed to read media file", "file_path", filePath, "error", err.Error())
			return nil, err
		}

		s.log.Debug("Media file read successfully", "file_path", filePath, "frame_count", len(fileFrames))

		frameCounts = append(frameCounts, len(fileFrames))

		// Добавляем все фреймы этого файла в общий массив
		data = append(data, fileFrames...)
	}

	// Выполняем инференс для всех фреймов
	s.log.Info("Starting inference", "total_frame_count", len(data))

	results := make([]float32, len(filePaths))

	for i, runner := range s.modelRunners {
		allScores, err := runner.Infer(data)
		if err != nil {
			s.log.Error("Inference failed", "error", err.Error())
			return nil, err
		}

		s.log.Info("Inference completed", "model_index", i, "score_count", len(allScores))

		// Объединяем результаты для фреймов одного видео
		var modelResults []float32
		frameIndex := 0

		for i, count := range frameCounts {
			if count == 1 {
				// Это изображение - просто добавляем один результат
				// Check bounds before accessing allScores
				if frameIndex >= len(allScores) {
					s.log.Error("Index out of bounds when processing image", "file_path", filePaths[i], "frame_index", frameIndex, "all_scores_len", len(allScores))
					continue
				}
				modelResults = append(modelResults, allScores[frameIndex])
				s.log.Debug("Image processed", "file_path", filePaths[i], "score", float64(allScores[frameIndex]))
				frameIndex++
			} else {
				// Это видео - берем максимальный результат среди всех фреймов
				// Check bounds before slicing allScores
				if frameIndex+count > len(allScores) {
					s.log.Error("Index out of bounds when processing video", "file_path", filePaths[i], "frame_index", frameIndex, "count", count, "all_scores_len", len(allScores))
					continue
				}
				videoScores := allScores[frameIndex : frameIndex+count]
				maxScore := getMaxScore(videoScores)
				modelResults = append(modelResults, maxScore)
				s.log.Debug("Video processed", "file_path", filePaths[i], "max_score", float64(maxScore), "frame_count", count)
				frameIndex += count
			}
		}

		// Объединяем результаты всех моделей
		for i, score := range modelResults {
			if score > results[i] {
				results[i] = score
			}
		}

		// TODO: если отработавшая модель уже определила score выше порога, то можно не продолжать
	}

	s.log.Info("Moderation completed", "result_count", len(results))
	return results, nil
}

func getMaxScore(scores []float32) float32 {
	var maxScore float32
	for _, score := range scores {
		if score > maxScore {
			maxScore = score
		}
	}
	return maxScore
}
