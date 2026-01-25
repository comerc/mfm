package mediareader

import (
	"bytes"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"net/http"
	"os"
	"os/exec"
	"strings"
)

type Repo struct {
}

// New creates a new instance of MediaReader repository
func New() *Repo {
	return &Repo{}
}

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

	contentType := http.DetectContentType(data)

	if strings.HasPrefix(contentType, "image/") {
		return s.processImage(data)
	}

	if strings.HasPrefix(contentType, "video/") {
		return s.processVideo(filePath)
	}

	return nil, fmt.Errorf("unsupported content type: %s", contentType)
}

func (s *Repo) processVideo(filePath string) ([][]byte, error) {
	// ffmpeg command to extract frames
	// -i filePath: input file
	// -vf fps=2,scale=224:224:force_original_aspect_ratio=increase,crop=224:224: extract 2 frames per second, resize to cover 224x224 and crop
	// -f rawvideo: output format raw video
	// -pix_fmt rgb24: pixel format RGB 24-bit
	// pipe:1: output to stdout
	cmd := exec.Command("ffmpeg",
		"-i", filePath,
		"-vf", "fps=2,scale=224:224:force_original_aspect_ratio=increase,crop=224:224",
		"-f", "rawvideo",
		"-pix_fmt", "rgb24",
		"pipe:1",
	)

	// Capture stdout
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to get stdout pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("failed to start ffmpeg: %w", err)
	}

	var frames [][]byte
	frameSize := 224 * 224 * 3
	buffer := make([]byte, frameSize)

	for {
		n, err := io.ReadFull(stdout, buffer)
		if err == io.EOF {
			break
		}
		if err != nil {
			if err == io.ErrUnexpectedEOF {
				// Less bytes than expected for a full frame, ignore incomplete frame
				break
			}
			return nil, fmt.Errorf("error reading video frame: %w", err)
		}
		if n == frameSize {
			// Copy buffer to new slice because buffer is reused
			frame := make([]byte, frameSize)
			copy(frame, buffer)
			frames = append(frames, frame)
		}
	}

	if err := cmd.Wait(); err != nil {
		// Only return error if we didn't extract any frames, otherwise partial success is okay?
		// ffmpeg might exit with non-zero on some warnings.
		// For now, strict check.
		return nil, fmt.Errorf("ffmpeg exited with error: %w", err)
	}

	return frames, nil
}

func (s *Repo) processImage(data []byte) ([][]byte, error) {
	img, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("failed to decode image: %w", err)
	}

	// Resize to 224x224
	resized := s.resize(img, 224, 224)

	// Convert to RGB24
	rgbData := make([]byte, 224*224*3)
	bounds := resized.Bounds()
	idx := 0
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			r, g, b, _ := resized.At(x, y).RGBA()
			// RGBA returns values in [0, 65535], we need [0, 255]
			rgbData[idx] = uint8(r >> 8)
			rgbData[idx+1] = uint8(g >> 8)
			rgbData[idx+2] = uint8(b >> 8)
			idx += 3
		}
	}

	return [][]byte{rgbData}, nil
}

// Check if file exists and has supported extension/type
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

// Simple nearest-neighbor resize
func (s *Repo) resize(img image.Image, width, height int) image.Image {
	bounds := img.Bounds()
	w := bounds.Dx()
	h := bounds.Dy()
	newImg := image.NewRGBA(image.Rect(0, 0, width, height))
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			// Calculate nearest source pixel
			srcX := bounds.Min.X + (x*w)/width
			srcY := bounds.Min.Y + (y*h)/height
			newImg.Set(x, y, img.At(srcX, srcY))
		}
	}
	return newImg
}
