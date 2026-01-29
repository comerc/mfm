package main

import (
	"context"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	slogzap "github.com/samber/slog-zap/v2"
	"go.uber.org/zap"

	mediareader "github.com/comerc/nsfw-mod/internal/repo/media_reader"
	openrunner "github.com/comerc/nsfw-mod/internal/repo/open_runner"
	"github.com/comerc/nsfw-mod/internal/service/moderation"
	"github.com/comerc/nsfw-mod/pkg/onnxinit"
	"github.com/comerc/nsfw-mod/pkg/utils"
)

func main() {
	// Загружаем переменные окружения из .env файла
	godotenv.Load()

	// Инициализация ONNX Runtime
	onnxinit.Initialize()

	// Инициализация глобального логгера
	zapLogger, _ := zap.NewProduction()
	handler := slogzap.Option{Logger: zapLogger}.NewZapHandler()
	slog.SetDefault(slog.New(handler))

	// Создаем логгер с контекстом модуля
	log := slog.With("module", "main")

	// Инициализация репозиториев
	mediaReader := mediareader.New()
	openRunner := openrunner.New()
	defer openRunner.Close()
	// TODO: подключить vitRunner := vitrunner.New()
	// defer vitRunner.Close()

	// Создаем сервис модерации (используем OpenRunner по умолчанию)
	moderationService := moderation.New(mediaReader, openRunner)

	log.Info("All dependencies are initialized")

	http.HandleFunc("/live", liveHandler)

	// Добавляем эндпоинт для модерации
	http.HandleFunc("/moderate", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// TODO: moderationService.Moderate()
		_ = moderationService

		// Простой ответ, чтобы показать, что сервис готов к работе
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status": "moderation service ready"}`))
	})

	// Получаем порт из переменной окружения или используем 8081 по умолчанию
	port := os.Getenv("HTTP_PORT")
	if port == "" {
		log.Error("Undefined port")
		return
	}

	host := "localhost" // или можно использовать "", чтобы принимать со всех интерфейсов

	// Проверяем, свободен ли порт
	addr := net.JoinHostPort(host, port)
	if !utils.IsPortFree(addr) {
		log.Error("Port is already in use", "addr", addr)
		return
	}

	log.Info("Server starting", "addr", addr)
	server := &http.Server{
		Addr: addr,
		// Add timeouts to prevent resource exhaustion
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Запускаем сервер в отдельной горутине
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Error("Server failed to start", "error", err.Error())
		}
	}()

	log.Info("Server started", "addr", addr)

	// Ожидаем сигнал завершения
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("Shutting down server...")

	// Устанавливаем таймаут для graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Завершаем работу сервера
	if err := server.Shutdown(ctx); err != nil {
		log.Error("Server forced to shutdown", "error", err.Error())
	} else {
		log.Info("Server exited properly")
	}
}

// liveHandler возвращает 200 OK
func liveHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}
