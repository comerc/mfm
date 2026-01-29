package onnxinit

import (
	"sync"

	"github.com/yalue/onnxruntime_go"
)

var (
	once        sync.Once
	initialized bool
	err         error
)

// Initialize инициализирует ONNX Runtime для всего приложения
func Initialize() error {
	once.Do(func() {
		// Устанавливаем путь к shared library
		onnxruntime_go.SetSharedLibraryPath("/usr/local/lib/libonnxruntime.dylib")

		// Инициализируем ONNX Runtime
		err = onnxruntime_go.InitializeEnvironment()
		if err == nil {
			initialized = true
		}
	})

	return err
}
