package onnxinit

import (
	"github.com/yalue/onnxruntime_go"
)

// Initialize инициализирует ONNX Runtime для всего приложения
func Initialize() {
	// Устанавливаем путь к shared library
	onnxruntime_go.SetSharedLibraryPath("/usr/local/lib/libonnxruntime.dylib")

	// Инициализируем ONNX Runtime
	// Игнорируем ошибку повторной инициализации, которая может возникнуть в тестах
	_ = onnxruntime_go.InitializeEnvironment()
}
