### Инструкция по экспорту Falconsai/nsfw_image_detection (ViT) в ONNX

Вот полная, пошаговая и абсолютно автономная инструкция. Мы пройдем путь от чистого терминала до готового работающего классификатора на базе **ViT (Vision Transformer)**, как в проекте LocalMod.

---

### Шаг 1: Создание окружения и установка библиотек

Мы используем `uv` для скорости, но команды стандартны. Выполни это в корне своего проекта:

```bash
# 1. Создаем виртуальное окружение
uv venv --python 3.11
source .venv/bin/activate

# 2. Устанавливаем всё необходимое для экспорта и работы
uv pip install transformers torch onnx onnxruntime opencv-python numpy pillow

```

---

### Шаг 2: Скрипт экспорта модели в ONNX (`convert_vit.py`)

Этот скрипт скачивает модель **Falconsai/nsfw_image_detection**, исправляет несовместимость с форматом ONNX и встраивает Softmax (чтобы на выходе были проценты 0..1).

```python
import torch
import torch.nn as nn
from transformers import AutoModelForImageClassification
import os

# Пути
MODEL_ID = "Falconsai/nsfw_image_detection"
OUTPUT_DIR = "assets"
OUTPUT_PATH = os.path.join(OUTPUT_DIR, "vit_nsfw.onnx")

os.makedirs(OUTPUT_DIR, exist_ok=True)

print(f"📦 Загрузка модели {MODEL_ID}...")
# attn_implementation="eager" отключает несовместимые с ONNX оптимизации
model = AutoModelForImageClassification.from_pretrained(
    MODEL_ID, 
    attn_implementation="eager"
)

# Обертка для добавления Softmax прямо в граф модели
class ViTWithSoftmax(nn.Module):
    def __init__(self, model):
        super().__init__()
        self.vit = model
        self.softmax = nn.Softmax(dim=1)
    
    def forward(self, x):
        logits = self.vit(x).logits
        return self.softmax(logits)

model_wrapped = ViTWithSoftmax(model)
model_wrapped.eval()

# Подготовка входа: [Batch, Channels, Height, Width]
dummy_input = torch.randn(1, 3, 224, 224)

print("🔄 Экспорт в ONNX (Opset 14)...")
torch.onnx.export(
    model_wrapped,
    dummy_input,
    OUTPUT_PATH,
    export_params=True,
    opset_version=14,
    do_constant_folding=True,
    input_names=['input'],
    output_names=['output'],
    dynamic_axes={'input': {0: 'batch_size'}, 'output': {0: 'batch_size'}}
)

print(f"✨ Готово! Модель сохранена в: {OUTPUT_PATH}")

```

---

### Шаг 3: Скрипт для проверки (инференса) (`check_vit.py`)

Этот скрипт берет твою картинку и прогоняет её через ViT. **Важно:** ViT требует специфическую нормализацию ImageNet.

```python
import cv2
import numpy as np
import onnxruntime as ort
import sys

def predict_nsfw(img_path):
    # 1. Загрузка сессии
    session = ort.InferenceSession("assets/vit_nsfw.onnx")
    
    # 2. Чтение и ресайз
    img = cv2.imread(img_path)
    if img is None:
        print("❌ Файл не найден")
        return
        
    img = cv2.cvtColor(img, cv2.COLOR_BGR2RGB)
    img = cv2.resize(img, (224, 224)).astype(np.float32) / 255.0
    
    # 3. Нормализация ImageNet (Критически важно для ViT)
    mean = np.array([0.485, 0.456, 0.406], dtype=np.float32)
    std = np.array([0.229, 0.224, 0.225], dtype=np.float32)
    img = (img - mean) / std
    
    # 4. Перестановка осей HWC -> CHW [224, 224, 3] -> [3, 224, 224]
    img = img.transpose(2, 0, 1)
    img = np.expand_dims(img, axis=0)
    
    # 5. Прогон через модель
    outputs = session.run(None, {"input": img})[0][0]
    
    # Индексы: 0 - Normal, 1 - NSFW
    prob_nsfw = outputs[1]
    
    print(f"--- Результат ---")
    print(f"🔞 NSFW Вероятность: {prob_nsfw:.4f}")
    print(f"📢 Вердикт: {'NSFW' if prob_nsfw > 0.5 else 'SAFE'}")

if __name__ == "__main__":
    if len(sys.argv) > 1:
        predict_nsfw(sys.argv[1])
    else:
        print("Использование: python check_vit.py my_image.jpg")

```

---

### Памятка по формату данных (Чекист)

| Параметр | Значение | Почему так? |
| --- | --- | --- |
| **Размер** | 224x224 | Стандарт для базового ViT |
| **Оси** | NCHW | PyTorch-модели работают с каналами в начале |
| **Нормализация** | ImageNet Mean/Std | Модель училась видеть мир через эти коэффициенты |
| **Выход** | 2 числа (Normal, NSFW) | Это бинарный классификатор |

---

**Запускай сначала `python convert_vit.py`. После того как в папке `assets` появится файл, тестируй через `python check_vit.py <путь_к_картинке>`.**

Нужно ли мне добавить в финальный скрипт логику, которая будет автоматически скачивать и подкладывать картинку для теста?