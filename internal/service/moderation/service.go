package moderation

import (
	"fmt"
)

type MediaReader interface {
	Read(filePath string) ([]byte, error)
}

type ModelRunner interface {
	Infer(data []byte) (float64, error)
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
	imgData, err := s.mediaReader.Read(filePath)

	if err != nil {
		return Result{
			Error: err.Error(),
		}
	}

	// Run inference via ModelRunner
	nsfwProb, err := s.modelRunner.Infer(imgData)
	if err != nil {
		return Result{
			Error: fmt.Sprintf("inference failed: %v", err),
		}
	}

	// Decision threshold (configurable)
	isNSFW := nsfwProb > 0.5
	confidence := nsfwProb
	if !isNSFW {
		confidence = 1.0 - nsfwProb // confidence for "safe" class
	}

	return Result{
		IsNSFW:     isNSFW,
		Confidence: confidence,
		Categories: []string{},
	}
}
