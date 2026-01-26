package vitrunner

import (
	"testing"
)

func TestRepo_Infer(t *testing.T) {
	repo := New()

	// Test with nil data
	result, err := repo.Infer(nil)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if result.IsNSFW != false {
		t.Errorf("expected false, got %v", result.IsNSFW)
	}

	// Test with empty batch
	result, err = repo.Infer([][]byte{})
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if result.IsNSFW != false {
		t.Errorf("expected false, got %v", result.IsNSFW)
	}

	// Test with some data (mock implementation always returns appropriate values when session is nil)
	frame := make([]byte, 224*224*3)
	frames := [][]byte{frame}
	result, err = repo.Infer(frames)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	// With mock implementation (session=nil), result should be default
	if result.IsNSFW != false {
		t.Errorf("expected false, got %v", result.IsNSFW)
	}
}
