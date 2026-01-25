package moderation

import (
	"fmt"

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
func (s *Service) Moderate(filePath string) domain.ModerationResult {
	frames, err := s.mediaReader.Read(filePath)

	if err != nil {
		return domain.ModerationResult{
			Error: err.Error(),
		}
	}

	// Run inference via ModelRunner
	result, err := s.modelRunner.Infer(frames)
	if err != nil {
		return domain.ModerationResult{
			Error: fmt.Sprintf("inference failed: %v", err),
		}
	}

	return result
}
