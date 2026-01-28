package vitrunner

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInfer(t *testing.T) {
	t.Parallel()
	// Arrange
	repo := New()

	tests := []struct {
		name    string
		setup   func(frames *[][]byte)
		wantErr bool
		wantLen int
	}{
		{
			name: "nil_data",
			setup: func(frames *[][]byte) {
				*frames = nil
			},
			wantErr: false,
			wantLen: 0,
		},
		{
			name: "empty_batch",
			setup: func(frames *[][]byte) {
				*frames = [][]byte{}
			},
			wantErr: false,
			wantLen: 0,
		},
		{
			name: "valid_frame",
			setup: func(frames *[][]byte) {
				frame := make([]byte, 224*224*3)
				*frames = [][]byte{frame}
			},
			wantErr: false,
			wantLen: 0, // Because session is nil, Infer returns empty slice
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			// Arrange
			var inputFrames [][]byte
			tt.setup(&inputFrames)
			// Act
			got, err := repo.Infer(inputFrames)

			// Assert
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.Len(t, got, tt.wantLen)
		})
	}
}
