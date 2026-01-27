package moderation

import (
	"github.com/comerc/nsfw-mod/internal/domain"
)

type MediaReader interface {
	Read(filePath string) ([][]byte, error)
}

type ModelRunner interface {
	Infer(data [][]byte) (domain.ModerationResult, error)
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

// Moderate analyzes the file with the given filePath for NSFW content
func (s *Service) Moderate(filePaths []string) []float32 {
	scores := make([]float32, 0, len(filePaths))

	for _, filePath := range filePaths {
		oneFileFrames, err := s.mediaReader.Read(filePath)
		if err != nil {
			// В случае ошибки чтения файла, добавляем нулевую оценку
			scores = append(scores, 0.0)
			continue
		}

		// Run inference via ModelRunner для каждого файла отдельно
		result, err := s.modelRunner.Infer(oneFileFrames)
		if err != nil {
			// В случае ошибки инференса, добавляем нулевую оценку
			scores = append(scores, 0.0)
			continue
		}

		scores = append(scores, result.Score)
	}

	return scores
}
