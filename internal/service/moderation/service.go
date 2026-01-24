package moderation

type FileReader interface {
	Read(filePath string) error
}

// Service implements Service
type Service struct {
	uploadDir  string
	fileReader FileReader
}

// New creates a new instance of ModerationService
func New(uploadDir string, fileReader FileReader) *Service {
	return &Service{
		uploadDir:  uploadDir,
		fileReader: fileReader,
	}
}

// Moderate analyzes the file with the given filePath for NSFW content
// For now, this is a mock implementation that checks file extension
func (s *Service) Moderate(filePath string) Result {
	err := s.fileReader.Read(filePath)

	if err != nil {
		return Result{
			Error: err.Error(),
		}
	}

	// Mock result - in production, this would analyze the actual image
	return Result{
		IsNSFW:     false, // Assume safe for demo
		Confidence: 0.95,
		Categories: []string{},
	}
}
