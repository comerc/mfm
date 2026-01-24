package moderation

import (
	"fmt"

	"github.com/comerc/nsfw-mod/internal/repo/model_runner"
)

type FileReader interface {
	Read(filePath string) ([]byte, error)
}

// Service implements Service
type Service struct {
	uploadDir   string
	fileReader  FileReader
	modelRunner *modelrunner.Repo
}

// New creates a new instance of ModerationService
func New(uploadDir string, fileReader FileReader, modelRunner *modelrunner.Repo) *Service {
	return &Service{
		uploadDir:   uploadDir,
		fileReader:  fileReader,
		modelRunner: modelRunner,
	}
}

// Moderate analyzes the file with the given filePath for NSFW content
func (s *Service) Moderate(filePath string) Result {
	imgData, err := s.fileReader.Read(filePath)

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
