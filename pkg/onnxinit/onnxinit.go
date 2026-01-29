package onnxinit

import (
	"sync"

	"github.com/yalue/onnxruntime_go"
)

var (
	once sync.Once
)

// Initialize инициализирует ONNX Runtime для всего приложения
func Initialize() {
	once.Do(func() {
		// Устанавливаем путь к shared library
		onnxruntime_go.SetSharedLibraryPath("/usr/local/lib/libonnxruntime.dylib")

		// Инициализируем ONNX Runtime
		err := onnxruntime_go.InitializeEnvironment()
		if err != nil {
			panic(err.Error())
		}
	})
}
