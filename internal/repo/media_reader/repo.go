package mediareader

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

type Repo struct{}

func (s *Repo) Read(filePath string) ([][]byte, error) {
	// Check file first
	if err := s.check(filePath); err != nil {
		return nil, err
	}

	// Read entire file at once (io.ReadAll handles chunking internally)
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("unable to open file: %w", err)
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("error reading file: %w", err)
	}

	return [][]byte{data}, nil
}

// check performs initial validation without loading full file
func (s *Repo) check(filePath string) error {
	// Check if file exists
	file, err := os.Open(filePath)
	if os.IsNotExist(err) {
		return fmt.Errorf("file not found: %s", filePath)
	}
	if err != nil {
		return fmt.Errorf("unable to open file: %w", err)
	}
	defer file.Close()

	// Read first 512 bytes to detect content type
	buffer := make([]byte, 512)
	n, err := file.Read(buffer)
	if err != nil && n == 0 {
		return fmt.Errorf("unable to read file: %w", err)
	}

	// Detect content type
	contentType := http.DetectContentType(buffer[:n])

	// Check if it's an image or video type
	if strings.HasPrefix(contentType, "image/") || strings.HasPrefix(contentType, "video/") {
		return nil
	}

	return fmt.Errorf("unsupported file type: %s", contentType)
}
