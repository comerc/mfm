package moderation

import (
	"fmt"
)

type MediaReader interface {
	Read(filePath string) ([][]byte, error)
}

type ModelRunner interface {
	Infer(data [][]byte) (bool, error)
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
func (s *Service) Moderate(filePath string) Result {
	frames, err := s.mediaReader.Read(filePath)

	if err != nil {
		return Result{
			Error: err.Error(),
		}
	}

	// Run inference via ModelRunner
	isNSFW, err := s.modelRunner.Infer(frames)
	if err != nil {
		return Result{
			Error: fmt.Sprintf("inference failed: %v", err),
		}
	}

	return Result{
		IsNSFW:     isNSFW,
		Categories: []string{},
	}
}
