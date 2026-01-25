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

see: doc/models/<model>.md

## TODO

- [ ] Добавить Grafana MCP

```json
    "grafana": {
      "command": "docker",
      "args": [
        "run", "--rm", "-i",
        "-e", "GRAFANA_URL=http://host.docker.internal:3000",
        "-e", "GRAFANA_API_KEY=ваш_токен_здесь",
        "grafana/mcp-grafana:latest"
      ]
    }
```

- исследование моделей NSFW (хочу попробовать запустить каждую модель из перечисленных и замерить их производительность)
  - [ ] FaceONNX/NsfwONNX (на базе MobileNet или SqueezeNet) - быстрый фильтр "Yes/No", через OpenVINO
  - [ ] MobileNetV3
  - [ ] CLIP (LAION‑AI/CLIP‑based‑NSFW‑Detector) - Understands semantics & nuance.
  - [ ] OpenNSFW2 (Yahoo Open‑NSFW, Keras/TensorFlow 2) - оптимизирована для быстрой проверки кадров в потоке "Yes/No"
  - [ ] NudeNet - находит конкретные зоны (грудь, гениталии и т.д.)
  - [ ] YOLOv8 - аналог NudeNet
  - [ ] DeepDanbooru (KichangKim/DeepDanbooru) - для аниме / хентай
  - [ ] WD14-ViT (Vision Transformer) - как основное ядро
  - [ ] GantMan/NSFW_model (версия InceptionV3 или ResNet)
