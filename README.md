# nsfw-mod
Moderation NSFW

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

- исследование моделей NSFW
  - [ ] FaceONNX/NsfwONNX (на базе MobileNet или SqueezeNet) - быстрый фильтр "Yes/No", через OpenVINO
  - [ ] MobileNetV3
  - [ ] CLIP (LAION‑AI/CLIP‑based‑NSFW‑Detector) - Understands semantics & nuance.
  - [ ] OpenNSFW2 (Yahoo Open‑NSFW, Keras/TensorFlow 2) - оптимизирована для быстрой проверки кадров в потоке "Yes/No"
  - [ ] NudeNet - находит конкретные зоны (грудь, гениталии и т.д.)
  - [ ] YOLOv8 - аналог NudeNet
  - [ ] DeepDanbooru (KichangKim/DeepDanbooru) - для аниме / хентай
  - [ ] WD14-ViT (Vision Transformer) - как основное ядро
  - [ ] GantMan/NSFW_model (версия InceptionV3 или ResNet)
