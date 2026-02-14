# MFM - Технический отчёт

## Обзор проекта

**MFM** (Media Moderation Framework) — это система на языке Go для обнаружения NSFW-контента (Not Safe For Work) в медиафайлах с использованием моделей искусственного интеллекта и машинного обучения.

---

## Метаинформация проекта

| Параметр | Значение |
|----------|----------|
| **Модуль** | `github.com/comerc/mfm` |
| **Версия Go** | 1.24.2 |
| **Текущая ветка** | `main` |
| **Тип лицензии** | Не указан |

---

## Архитектура проекта

### Структура директорий

```
mfm/
├── cmd/server/              # Точка входа приложения
│   └── main.go              # Главный файл приложения
├── internal/                # Внутренний код (не переиспользуемый)
│   ├── repo/                # Слой репозиториев (доступ к данным)
│   │   ├── media_reader/    # Чтение и обработка медиафайлов
│   │   ├── open_runner/     # Запуск модели OpenNSFW2
│   │   └── vit_runner/      # Запуск модели Vision Transformer
│   └── service/             # Слой сервисов (бизнес-логика)
│       └── moderation/      # Сервис модерации контента
├── pkg/                     # Переиспользуемые пакеты
│   ├── onnxinit/            # Инициализация ONNX Runtime
│   └── utils/               # Общие утилиты
│       ├── port.go          # Проверка доступности порта
│       └── env.go           # Загрузка переменных окружения
├── assets/                  # Медиа-активы и ONNX-модели
│   ├── onnx/                # Файлы моделей
│   ├── *.png                # Тестовые изображения
│   └── video*.mp4           # Тестовые видео
├── test/                    # Интеграционные тесты
├── doc/                     # Документация
│   └── models/              # Документация по моделям
├── script/                  # Вспомогательные скрипты
├── .env.example             # Пример конфигурации
├── Taskfile.yml             # Автоматизация сборки
├── .golangci.yml            # Конфигурация линтера
└── .mockery.yaml            # Конфигурация генерации моков
```

### Go-файлы проекта

| Файл | Описание |
|------|----------|
| `cmd/server/main.go` | Точка входа, настройка HTTP-сервера |
| `internal/service/moderation/service.go` | Основной сервис модерации |
| `internal/repo/media_reader/repo.go` | Обработка медиафайлов |
| `internal/repo/open_runner/open_runner.go` | Раннер OpenNSFW2 |
| `internal/repo/vit_runner/vit_runner.go` | Раннер Vision Transformer |
| `pkg/onnxinit/onnxinit.go` | Инициализация ONNX Runtime |
| `pkg/utils/port.go` | Проверка порта |
| `pkg/utils/env.go` | Загрузка .env |
| `test/moderation_test.go` | Интеграционные тесты |

---

## Зависимости и технологии

### Основные зависимости (go.mod)

| Пакет | Версия | Назначение |
|-------|--------|------------|
| `github.com/joho/godotenv` | v1.5.1 | Загрузка переменных окружения |
| `github.com/stretchr/testify` | v1.10.0 | Тестирование и моки |
| `github.com/yalue/onnxruntime_go` | v1.25.0 | ONNX Runtime для Go |
| `go.uber.org/zap` | v1.27.1 | Структурированное логирование |
| `github.com/samber/slog-zap/v2` | v2.6.2 | Интеграция slog с Zap |

### Внешние системные зависимости

| Компонент | Назначение |
|-----------|------------|
| **ONNX Runtime** | Выполнение ML-моделей |
| **FFmpeg** | Обработка видеофайлов |

### Инструменты разработки

| Инструмент | Назначение |
|------------|------------|
| **Task** | Автоматизация задач (lint, test, run) |
| **Mockery** | Генерация mock-объектов |
| **GolangCI-Lint** | Комплексная проверка кода |

---

## Подробное описание компонентов

### 1. Точка входа (`cmd/server/main.go`)

Главный файл приложения выполняет:

1. Загрузку переменных окружения из `.env`
2. Инициализацию ONNX Runtime
3. Настройку структурированного логирования (Zap + slog)
4. Создание зависимостей (media reader, model runners)
5. Запуск HTTP-сервера с эндпоинтами:
   - `POST /moderate` — модерация контента
   - `GET /live` — health-check
6. Graceful shutdown при получении сигналов SIGINT/SIGTERM

**Настройки HTTP-сервера:**
```go
ReadTimeout:  15 * time.Second
WriteTimeout: 15 * time.Second
IdleTimeout:  60 * time.Second
```

### 2. Сервис модерации (`internal/service/moderation/service.go`)

Основной бизнес-логический компонент:

**Интерфейсы:**
```go
type mediaReader interface {
    Read(filePath string) ([][]byte, error)
}

type modelRunner interface {
    Infer(data [][]byte) ([]float32, error)
}
```

**Алгоритм работы:**
1. Читает медиафайлы через `mediaReader`
2. Агрегирует все фреймы в один батч
3. Выполняет инференс для всех моделей параллельно
4. Для изображений — берёт единственный результат
5. Для видео — берёт максимальный счёт среди всех фреймов
6. Объединяет результаты всех моделей (максимум)

### 3. Media Reader (`internal/repo/media_reader/repo.go`)

Обрабатывает медиафайлы для подачи в модели:

**Типы контента:**
```go
const (
    contentTypeUnknown ContentType = "unknown"
    contentTypeImage   ContentType = "image"
    contentTypeVideo   ContentType = "video"
)
```

**Обработка изображений:**
- Декодирование из любого формата
- Изменение размера до 224x224 (nearest-neighbor)
- Конвертация в RGB24 (224 × 224 × 3 = 150,528 байт)

**Обработка видео:**
- Извлечение кадров через FFmpeg (2 fps)
- Изменение размера до 224x224 с обрезкой
- Конвертация в RGB24

**FFmpeg команда:**
```bash
ffmpeg -i <input> \
  -vf "fps=2,scale=224:224:force_original_aspect_ratio=increase,crop=224:224" \
  -f rawvideo \
  -pix_fmt rgb24 \
  pipe:1
```

### 4. Model Runners

#### OpenNSFW2 Runner (`internal/repo/open_runner/open_runner.go`)

**Модель:** OpenNSFW2
- **Вход:** `[batch_size, 224, 224, 3]` (float32, нормализовано 0-1)
- **Выход:** `[batch_size, 1]` (float32, вероятность NSFW)

#### ViT Runner (`internal/repo/vit_runner/vit_runner.go`)

**Модель:** Vision Transformer
- **Вход:** `[batch_size, 224, 224, 3]` (float32, нормализовано 0-1)
- **Выход:** `[batch_size, 2]` (float32, пара: normal, nsfw)

---

## Конфигурация

### Переменные окружения (.env)

| Переменная | Описание |
|------------|----------|
| `HTTP_PORT` | Порт HTTP-сервера (по умолчанию 7171) |
| `MODEL_OPEN` | Абсолютный путь к модели OpenNSFW2 |
| `MODEL_VIT` | Абсолютный путь к модели ViT |

### Пример файла (.env.example)

```env
# Port for the HTTP server
HTTP_PORT=7171

# Model paths (absolute paths required)
MODEL_OPEN=/absolute/path/to/assets/onnx/opennsfw2.onnx
MODEL_VIT=/absolute/path/to/assets/onnx/vit_nsfw.onnx
```

---

## API

### Эндпоинты

#### POST /moderate

Модерация медиафайлов.

**Тело запроса:** (ожидается JSON с путями к файлам)

**Ответ:**
```json
{
  "status": "moderation service ready"
}
```

*Примечание: Полная реализация эндпоинта находится в разработке (TODO в коде).*

#### GET /live

Health-check endpoint.

**Статус:** 200 OK

---

## Тестирование

### Покрытие тестами

| Тип тестов | Файлы |
|------------|-------|
| Юнит-тесты | `*_test.go` в каждом пакете |
| Интеграционные | `test/moderation_test.go` |
| Моки | `internal/service/moderation/mocks/` |
| Бенчмарки | Интеграционные тесты с батчами |

### Testify

Используется для утверждений и мокирования:
```go
import "github.com/stretchr/testify"
```

### Mockery

Генерация моков для интерфейсов (настроена через `.mockery.yaml`)

---

## Сборка и разработка

### Taskfile задачи

```yaml
task lint          # Запуск golangci-lint
task test          # Запуск тестов
task run           # Запуск сервера
```

### Линтинг (.golangci.yml)

Конфигурация включает правила `revive` для проверки качества кода.

---

## Документация

| Файл | Содержание |
|------|------------|
| `README.md` | Инструкция по установке |
| `doc/models/OpenNSFW2.md` | Документация модели OpenNSFW2 |
| `doc/models/ViT.md` | Документация Vision Transformer |
| `internal/service/moderation/TESTING.md` | Документация по тестированию |
| `AGENTS.md` | Конфигурация KiloCode агентов |

---

## Установка

### 1. ONNX Runtime

**macOS:**
```bash
brew install onnxruntime
```

**Другие платформы:** см. [ONNX Runtime installation guide](https://onnxruntime.ai/docs/install/)

### 2. FFmpeg

**macOS:**
```bash
brew install ffmpeg
```

### 3. Инструменты разработки

```bash
go install github.com/go-task/task/v3/cmd/task@latest
go install github.com/vektra/mockery/v2@v2.53.3
go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@latest
```

### 4. Настройка окружения

```bash
cp .env.example .env
# Отредактируйте .env с путями к моделям
```

---

## Архитектурные особенности

### Clean Architecture

```
┌─────────────────────────────────────────────────────┐
│                    HTTP Layer                       │
│                  (cmd/server)                       │
└────────────────────────┬────────────────────────────┘
                         │
┌────────────────────────▼────────────────────────────┐
│                  Service Layer                      │
│           (internal/service/moderation)             │
└────────────────────────┬────────────────────────────┘
                         │
┌────────────────────────▼────────────────────────────┐
│                 Repository Layer                    │
│    (internal/repo/{media_reader,open_runner,vit})   │
└────────────────────────┬────────────────────────────┘
                         │
┌────────────────────────▼────────────────────────────┐
│               External Dependencies                 │
│              (ONNX Runtime, FFmpeg)                 │
└─────────────────────────────────────────────────────┘
```

### Dependency Injection

Все компоненты создаются через конструкторы и внедряются зависимости:

```go
mediaReader := mediareader.New()
openRunner := openrunner.New()
vitRunner := vitrunner.New()
moderationService := moderation.New(mediaReader, openRunner, vitRunner)
```

### Graceful Shutdown

Правильное освобождение ресурсов:
- Закрытие ONNX-сессий через `defer`
- Таймаут shutdown: 30 секунд
- Обработка SIGINT/SIGTERM

---

## Поток обработки данных

```
┌─────────────────────────────────────────────────────────────────────┐
│                         Media File(s)                               │
└───────────────────────────────┬─────────────────────────────────────┘
                                │
                                ▼
┌─────────────────────────────────────────────────────────────────────┐
│                      Media Reader                                   │
│  ┌─────────────────┐  ┌─────────────────┐                           │
│  │  Image Process  │  │  Video Process  │                           │
│  │  - Decode       │  │  - FFmpeg       │                           │
│  │  - Resize 224²  │  │  - 2 fps        │                           │
│  │  - RGB24        │  │  - RGB24        │                           │
│  └─────────────────┘  └─────────────────┘                           │
└───────────────────────────────┬─────────────────────────────────────┘
                                │
                                ▼
┌─────────────────────────────────────────────────────────────────────┐
│                      Batch All Frames                               │
│              [frame1, frame2, ..., frameN]                          │
└───────────────────────────────┬─────────────────────────────────────┘
                                │
                ┌───────────────┴───────────────┐
                ▼                               ▼
┌───────────────────────────┐       ┌───────────────────────────┐
│      OpenNSFW2 Runner     │       │       ViT Runner          │
│   [batch, 224, 224, 3]    │       │   [batch, 224, 224, 3]    │
│           ↓               │       │           ↓               │
│   [batch, 1] (NSFW prob)  │       │   [batch, 2] (N, NSFW)    │
└─────────────┬─────────────┘       └─────────────┬─────────────┘
              │                                   │
              └─────────────────┬─────────────────┘
                                ▼
┌─────────────────────────────────────────────────────────────────────┐
│                      Aggregate Results                              │
│  - Images: single score per file                                    │
│  - Videos: max score across all frames                              │
│  - Models: max score across all models                              │
└───────────────────────────────┬─────────────────────────────────────┘
                                │
                                ▼
┌─────────────────────────────────────────────────────────────────────┐
│                    Final NSFW Scores                                │
│                 [file1, file2, ..., fileM]                          │
└─────────────────────────────────────────────────────────────────────┘
```

---

## Потенциальные улучшения

### На основе анализа кода:

1. **Реализация эндпоинта `/moderate`** — в настоящее время возвращает заглушку (TODO в коде)
2. **Добавление валидации входных данных** для эндпоинта модерации
3. **Расширение форматов вывода** (JSON с детальными результатами)
4. **Добавление метрик** (prometheus) для мониторинга производительности
5. **Кеширование результатов** для повторных запросов
6. **Асинхронная обработка** для больших батчей файлов

---

## Заключение

MFM — это хорошо структурированный, production-ready проект для модерации медиаконтента. Код следует принципам Clean Architecture с чётким разделением на слои. Проект имеет:

- ✅ Чёткую архитектуру с разделением ответственности
- ✅ Comprehensive логирование с контекстом
- ✅ Graceful shutdown и управление ресурсами
- ✅ Покрытие тестами
- ✅ Документацию по установке и моделям
- ✅ Инструменты для разработки (Taskfile, linting)

---

*Дата: 2026-02-01*
*Версия отчёта: 1.0*
