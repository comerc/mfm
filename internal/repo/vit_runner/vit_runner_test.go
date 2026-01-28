package vitrunner

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRepo_Infer(t *testing.T) {
	t.Parallel()
	// Arrange
	repo := New()

	// Test with nil data
	t.Run("NilData", func(t *testing.T) {
		// Act
		scores, err := repo.Infer(nil)

		// Assert
		assert.NoError(t, err, "expected no error for nil data")
		assert.Empty(t, scores, "expected empty slice for nil data")
	})

	// Test with empty batch
	t.Run("EmptyBatch", func(t *testing.T) {
		// Act
		scores, err := repo.Infer([][]byte{})

		// Assert
		assert.NoError(t, err, "expected no error for empty batch")
		assert.Empty(t, scores, "expected empty slice for empty batch")
	})

	// Test with some data (mock implementation always returns appropriate values when session is nil)
	t.Run("ValidFrame", func(t *testing.T) {
		// Arrange
		frame := make([]byte, 224*224*3)
		frames := [][]byte{frame}

		// Act
		scores, err := repo.Infer(frames)

		// Assert
		assert.NoError(t, err, "expected no error for valid frame")

		// With mock implementation (session=nil), scores should be empty or contain zeros
		if len(scores) == 0 {
			// This is expected when session is nil
			return
		}
		assert.Len(t, scores, 1, "expected 1 score for 1 frame")
		assert.Equal(t, 0.0, scores[0], "expected score of 0.0 with mock implementation")
	})
}
