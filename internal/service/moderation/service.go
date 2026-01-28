package moderation

import (
	"log/slog"
)

type MediaReader interface {
	Read(filePath string) ([][]byte, error)
}

type ModelRunner interface {
	Infer(data [][]byte) ([]float32, error)
}

// Service implements Service
type Service struct {
	log         *slog.Logger
	uploadDir   string
	mediaReader MediaReader
	modelRunner ModelRunner
}

// New creates a new instance of ModerationService
func New(uploadDir string, mediaReader MediaReader, modelRunner ModelRunner) *Service {
	return &Service{
		log:         slog.With(slog.String("module", "moderation")),
		uploadDir:   uploadDir,
		mediaReader: mediaReader,
		modelRunner: modelRunner,
	}
}

func (s *Service) Moderate(filePaths []string) ([]float32, error) {
	s.log.Info("Starting moderation", slog.Int("file_count", len(filePaths)), slog.Any("file_paths", filePaths))

	// Будем хранить информацию о том, сколько фреймов приходится на каждый файл
	var data [][]byte     // Все фреймы для инференса
	var frameCounts []int // Количество фреймов для каждого файла

	for _, filePath := range filePaths {
		s.log.Debug("Reading media file", slog.String("file_path", filePath))

		fileFrames, err := s.mediaReader.Read(filePath)
		if err != nil {
			s.log.Error("Failed to read media file", slog.String("file_path", filePath), slog.String("error", err.Error()))
			return nil, err
		}

		s.log.Debug("Media file read successfully", slog.String("file_path", filePath), slog.Int("frame_count", len(fileFrames)))

		frameCounts = append(frameCounts, len(fileFrames))

		// Добавляем все фреймы этого файла в общий массив
		for _, frame := range fileFrames {
			data = append(data, frame)
		}
	}

	// Выполняем инференс для всех фреймов
	s.log.Info("Starting inference", slog.Int("total_frame_count", len(data)))

	allScores, err := s.modelRunner.Infer(data)
	if err != nil {
		s.log.Error("Inference failed", slog.String("error", err.Error()))
		return nil, err
	}

	s.log.Info("Inference completed", slog.Int("score_count", len(allScores)))

	// Объединяем результаты для фреймов одного видео
	var results []float32
	frameIndex := 0

	for i, count := range frameCounts {
		if count == 1 {
			// Это изображение - просто добавляем один результат
			results = append(results, allScores[frameIndex])
			s.log.Debug("Image processed", slog.String("file_path", filePaths[i]), slog.Float64("score", float64(allScores[frameIndex])))
			frameIndex++
		} else {
			// Это видео - берем максимальный результат среди всех фреймов
			videoScores := allScores[frameIndex : frameIndex+count]
			maxScore := s.getMaxScore(videoScores)
			results = append(results, maxScore)
			s.log.Debug("Video processed", slog.String("file_path", filePaths[i]), slog.Float64("max_score", float64(maxScore)), slog.Int("frame_count", count))
			frameIndex += count
		}
	}

	s.log.Info("Moderation completed", slog.Int("result_count", len(results)))
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
