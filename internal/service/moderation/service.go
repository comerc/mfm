package moderation

import (
	"log/slog"
)

//go:generate mockery --name=MediaReader
type MediaReader interface {
	Read(filePath string) ([][]byte, error)
}

//go:generate mockery --name=ModelRunner
type ModelRunner interface {
	Infer(data [][]byte) ([]float32, error)
}

// Service implements Service
type Service struct {
	log         *slog.Logger
	mediaReader MediaReader
	modelRunner ModelRunner
}

// New creates a new instance of ModerationService
func New(mediaReader MediaReader, modelRunner ModelRunner) *Service {
	return &Service{
		log:         slog.With("module", "moderation"),
		mediaReader: mediaReader,
		modelRunner: modelRunner,
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

	allScores, err := s.modelRunner.Infer(data)
	if err != nil {
		s.log.Error("Inference failed", "error", err.Error())
		return nil, err
	}

	s.log.Info("Inference completed", "score_count", len(allScores))

	// Объединяем результаты для фреймов одного видео
	var results []float32
	frameIndex := 0

	for i, count := range frameCounts {
		if count == 1 {
			// Это изображение - просто добавляем один результат
			// Check bounds before accessing allScores
			if frameIndex >= len(allScores) {
				s.log.Error("Index out of bounds when processing image", "file_path", filePaths[i], "frame_index", frameIndex, "all_scores_len", len(allScores))
				continue
			}
			results = append(results, allScores[frameIndex])
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
			maxScore := s.getMaxScore(videoScores)
			results = append(results, maxScore)
			s.log.Debug("Video processed", "file_path", filePaths[i], "max_score", float64(maxScore), "frame_count", count)
			frameIndex += count
		}
	}

	s.log.Info("Moderation completed", "result_count", len(results))
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
