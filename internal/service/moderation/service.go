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

// ModerateOne обрабатывает один файл (картинку или видосик)
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

// ModerateImages обрабатывает только картинки в пакете
func (s *Service) ModerateImages(filePaths []string) ([]float32, error) {
	var data [][]byte

	for _, filePath := range filePaths {
		fileFrames, err := s.mediaReader.Read(filePath)
		if err != nil {
			return nil, err
		}

		data = append(data, fileFrames[0])
	}

	return s.modelRunner.Infer(data)
}
