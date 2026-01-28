package openrunner

import (
	"testing"
)

func TestRepo_Infer(t *testing.T) {
	t.Parallel()
	repo := New()

	// Test with nil data
	scores, err := repo.Infer(nil)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	// For nil/empty data, should return empty slice
	if len(scores) != 0 {
		t.Errorf("expected empty slice, got length %d", len(scores))
	}

	// Test with empty batch
	scores, err = repo.Infer([][]byte{})
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	// For empty batch, should return empty slice
	if len(scores) != 0 {
		t.Errorf("expected empty slice, got length %d", len(scores))
	}

	// Test with some data (mock implementation always returns appropriate values when session is nil)
	frame := make([]byte, 224*224*3)
	frames := [][]byte{frame}
	scores, err = repo.Infer(frames)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	// With mock implementation (session=nil), scores should be empty or contain zeros
	switch {
	case len(scores) == 0:
		// This is expected when session is nil
	case len(scores) != 1:
		t.Errorf("expected 1 score, got %d", len(scores))
	case scores[0] != 0.0:
		t.Errorf("expected score of 0.0, got %f", scores[0])
	}
}
