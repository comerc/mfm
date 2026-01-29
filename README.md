# nsfw-mod
Moderation NSFW

## Installation

### ONNX Runtime

The project uses ONNX Runtime for ML model inference. Install the ONNX Runtime shared library:

- **macOS**:

```bash
brew install onnxruntime
```

- **Other platforms**: See [ONNX Runtime installation guide](https://onnxruntime.ai/docs/install/).

### Model

Download an ONNX NSFW detection model and place it as `assets/<model>.onnx`. The model should accept input shape `[batch_size, 224, 224, 3]` (float32, normalized 0-1) and output shape `[batch_size, 1]` (float32, probability of NSFW).

См. подробности по установке моделей: doc/models/<model>.md

### ffmpeg

- **macOS**:

```bash
brew install ffmpeg
```

### Прочие Зависимости

```bash
go install github.com/go-task/task/v3/cmd/task@latest
go install github.com/vektra/mockery/v2@v2.53.3
go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@latest
```

### Environment Variables

Copy the `.env.example` file to `.env` and configure the environment variables:

```bash
cp .env.example .env
```

