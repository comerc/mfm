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

Download ONNX NSFW detection models and configure their paths in `.env` file. The service requires environment variables to be set:

- `MODEL_OPEN` - absolute path to OpenNSFW2 model
- `MODEL_VIT` - absolute path to ViT model

Models should accept input shape `[batch_size, 224, 224, 3]` (float32, normalized 0-1) and output shape `[batch_size, 1]` (float32, probability of NSFW) for OpenNSFW2 or `[batch_size, 2]` for ViT (probability pair: normal, nsfw).

If environment variables are not set, the service will panic with an error message.

See details on model conversion: doc/models/<model>.md

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
go install github.com/unused-interface-methods/unused-interface-methods@latest
```

### Environment Variables

Copy the `.env.example` file to `.env` and configure the environment variables:

```bash
cp .env.example .env
```

