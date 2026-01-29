package onnxinit

import (
	"github.com/yalue/onnxruntime_go"
)

// Initialize инициализирует ONNX Runtime для всего приложения
func Initialize() error {
	// Устанавливаем путь к shared library
	onnxruntime_go.SetSharedLibraryPath("/usr/local/lib/libonnxruntime.dylib")

	// Инициализируем ONNX Runtime
	return onnxruntime_go.InitializeEnvironment()
}
