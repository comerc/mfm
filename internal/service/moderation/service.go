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

func (s *Service) ModerateOne(filePath string) (float32, error) {
	data, err := s.mediaReader.Read(filePath)
	if err != nil {
		return 0.0, err
	}
	frameScores, err := s.modelRunner.Infer(data)
	if err != nil {
		return 0.0, err
	}
	var maxScore float32
	for _, score := range frameScores {
		if score > maxScore {
			maxScore = score
		}
	}
	return maxScore, nil
}

func (s *Service) Moderate(filePaths []string) ([]float32, error) {
	data := make([][]byte, 0, len(filePaths))

	for _, filePath := range filePaths {
		fileFrames, err := s.mediaReader.Read(filePath)
		if err != nil {
			return nil, err
		}

		data = append(data, fileFrames[0])
	}

	result, err := s.modelRunner.Infer(data)
	if err != nil {
		return nil, err
	}
	return result, nil
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
