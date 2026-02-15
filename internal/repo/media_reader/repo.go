package mediareader

import (
	"bytes"
	"fmt"
	"image"
	"io"
	"log/slog"
	"os"
	"os/exec"
	"strings"

	"github.com/h2non/filetype"
)

// ContentType represents the type of media content
type ContentType string

const (
	ContentTypeUnknown ContentType = "unknown"
	ContentTypeImage   ContentType = "image"
	ContentTypeVideo   ContentType = "video"
)

// Supported file extensions
const videoExts = "mp4|avi|mov|mkv"
const imageExts = "jpg|png"

type Repo struct {
	log *slog.Logger
}

// New creates a new instance of MediaReader repository
func New() *Repo {
	// Check if ffmpeg is installed
	if _, err := exec.LookPath("ffmpeg"); err != nil {
		panic("ffmpeg not found")
	}

	return &Repo{
		log: slog.With("module", "mediareader"),
	}
}

func (s *Repo) Read(filePath string) ([][]byte, error) {
	s.log.Info("Reading media file", "file_path", filePath)

	// Check file first and get content type
	contentType, err := s.Check(filePath)
	if err != nil {
		s.log.Error("Failed to check file", "file_path", filePath, "error", err.Error())
		return nil, err
	}

	// Read entire file at once (io.ReadAll handles chunking internally)
	file, err := os.Open(filePath)
	if err != nil {
		s.log.Error("Unable to open file", "file_path", filePath, "error", err.Error())
		return nil, fmt.Errorf("unable to open file: %w", err)
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		s.log.Error("Error reading file", "file_path", filePath, "error", err.Error())
		return nil, fmt.Errorf("error reading file: %w", err)
	}

	switch contentType {
	case ContentTypeImage:
		return s.processImage(data)
	case ContentTypeVideo:
		return s.processVideo(filePath)
	default:
		s.log.Error("Unsupported content type", "content_type", string(contentType), "file_path", filePath)
		return nil, fmt.Errorf("unsupported content type: %s", contentType)
	}
}

func (s *Repo) processVideo(filePath string) ([][]byte, error) {
	s.log.Info("Processing video file", "file_path", filePath)

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
		s.log.Error("Failed to get stdout pipe", "file_path", filePath, "error", err.Error())
		return nil, fmt.Errorf("failed to get stdout pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		s.log.Error("Failed to start ffmpeg", "file_path", filePath, "error", err.Error())
		return nil, fmt.Errorf("failed to start ffmpeg: %w", err)
	}

	var frames [][]byte
	frameSize := 224 * 224 * 3
	buffer := make([]byte, frameSize)

	frameCount := 0
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
			s.log.Error("Error reading video frame", "file_path", filePath, "error", err.Error())
			return nil, fmt.Errorf("error reading video frame: %w", err)
		}
		if n == frameSize {
			// Copy buffer to new slice because buffer is reused
			frame := make([]byte, frameSize)
			copy(frame, buffer)
			frames = append(frames, frame)
			frameCount++
		}
	}

	if err := cmd.Wait(); err != nil {
		// Only return error if we didn't extract any frames, otherwise partial success is okay?
		// ffmpeg might exit with non-zero on some warnings.
		// For now, strict check.
		s.log.Error("FFmpeg exited with error", "file_path", filePath, "error", err.Error())
		return nil, fmt.Errorf("ffmpeg exited with error: %w", err)
	}

	s.log.Info("Video processed successfully", "file_path", filePath, "frames_extracted", frameCount)
	return frames, nil
}

func (s *Repo) processImage(data []byte) ([][]byte, error) {
	s.log.Info("Processing image file", "data_size", len(data))

	img, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		s.log.Error("Failed to decode image", "error", err.Error())
		return nil, fmt.Errorf("failed to decode image: %w", err)
	}

	// Resize to 224x224
	resized := resize(img, 224, 224)

	// Convert to RGB24
	rgbData := make([]byte, 224*224*3)
	bounds := resized.Bounds()
	idx := 0
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			r, g, b, _ := resized.At(x, y).RGBA()
			// RGBA returns values in [0, 65535], we need [0, 255]
			// Safe conversion from uint32 to uint8
			rgbData[idx] = uint8(r / 257)   // #nosec G115 - safe conversion, values are in [0, 65535] range
			rgbData[idx+1] = uint8(g / 257) // #nosec G115 - safe conversion, values are in [0, 65535] range
			rgbData[idx+2] = uint8(b / 257) // #nosec G115 - safe conversion, values are in [0, 65535] range
			idx += 3
		}
	}

	return [][]byte{rgbData}, nil
}

// Check if file exists and has supported extension/type
func (s *Repo) Check(filePath string) (ContentType, error) {
	s.log.Debug("Checking file", "file_path", filePath)

	// Check if file exists
	file, err := os.Open(filePath)
	if os.IsNotExist(err) {
		s.log.Warn("File not found", "file_path", filePath)
		return ContentTypeUnknown, fmt.Errorf("file not found: %s", filePath)
	}
	if err != nil {
		s.log.Error("Unable to open file", "file_path", filePath, "error", err.Error())
		return ContentTypeUnknown, fmt.Errorf("unable to open file: %w", err)
	}
	defer file.Close()

	// Read first 512 bytes to detect content type
	buffer := make([]byte, 512)
	n, err := file.Read(buffer)
	if err != nil && n == 0 {
		s.log.Error("Unable to read file", "file_path", filePath, "error", err.Error())
		return ContentTypeUnknown, fmt.Errorf("unable to read file: %w", err)
	}

	// Detect content type using filetype
	kind, err := filetype.Match(buffer[:n])
	if err != nil {
		s.log.Error("Unable to detect file type", "file_path", filePath, "error", err.Error())
		return ContentTypeUnknown, fmt.Errorf("unable to detect file type: %w", err)
	}

	s.log.Debug("Detected content type", "mime_type", kind.MIME.Value, "extension", kind.Extension, "file_path", filePath)

	// Check if detected extension is in our supported list
	if !isExtensionSupported(kind.Extension) {
		s.log.Warn("Unsupported file type", "extension", kind.Extension, "file_path", filePath)
		return ContentTypeUnknown, fmt.Errorf("unsupported file type: %s", kind.Extension)
	}

	// Check if it's an image or video type
	if kind.MIME.Type == "image" {
		return ContentTypeImage, nil
	}
	if kind.MIME.Type == "video" {
		return ContentTypeVideo, nil
	}

	s.log.Warn("Unsupported file type", "mime_type", kind.MIME.Value, "file_path", filePath)
	return ContentTypeUnknown, fmt.Errorf("unsupported file type: %s", kind.MIME.Value)
}

// Simple nearest-neighbor resize
func resize(img image.Image, width, height int) image.Image {
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

// isExtensionSupported checks if the file extension is in the supported list
func isExtensionSupported(ext string) bool {
	supportedExts := videoExts + "|" + imageExts
	for _, supportedExt := range strings.Split(supportedExts, "|") {
		if ext == supportedExt {
			return true
		}
	}
	return false
}
