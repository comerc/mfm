## Инструкция по экспорту OpenNSFW2 в ONNX

### Шаг 1: Подготовка изолированной среды

Для корректной работы TensorFlow и конвертера используем `uv` для создания окружения с версией Python 3.11.

```bash
# Создаем окружение
uv venv --python 3.11
source .venv/bin/activate

# Устанавливаем необходимые пакеты
uv pip install opennsfw2 tensorflow-cpu tf2onnx onnx

```

---

### Шаг 2: Скрипт конвертации (`convert.py`)

Создай файл `convert.py` и вставь в него следующий код. Он учитывает актуальный API библиотеки и твои требования к выходным параметрам.

```python
import os
import tensorflow as tf
import tf2onnx
import opennsfw2 as nsfw

# Подготовка папки
os.makedirs("assets", exist_ok=True)

print("📦 Загрузка модели...")
# Используем актуальную функцию из библиотеки
base_model = nsfw.make_open_nsfw_model()

# Настройка входа и выхода:
# Вход: [batch, 224, 224, 3]
# Выход: берем только индекс 1 (NSFW probability) -> форма [batch, 1]
inputs = base_model.input
outputs = base_model.output[:, 1:] 
model = tf.keras.Model(inputs=inputs, outputs=outputs)

print("🔄 Конвертация в ONNX...")
spec = (tf.TensorSpec((None, 224, 224, 3), tf.float32, name="input"),)

output_path = "assets/opennsfw2.onnx"
model_proto, _ = tf2onnx.convert.from_keras(
    model, 
    input_signature=spec, 
    opset=13
)

with open(output_path, "wb") as f:
    f.write(model_proto.SerializeToString())

print(f"✨ Успешно сохранено в {output_path}")
print(f"Вход: {model.input_shape}")
print(f"Выход: {model.output_shape}")

```

---

### Шаг 3: Запуск

Просто запусти скрипт. При первом запуске он автоматически скачает веса модели с GitHub.

```bash
python convert.py

```

---

### Шаг 4: Технические параметры модели

После выполнения у тебя в папке `assets/` появится файл `opennsfw2.onnx.onnx` со следующими характеристиками:

* **Input (`input`)**: Тензор формы `[batch_size, 224, 224, 3]`.
* **Тип данных**: `float32`.
* **Нормализация**: Вычитание среднего [104, 117, 123].
* **Цветовая модель**: **BGR**.
* **Output**: Тензор формы `[batch_size, 1]`, где значение — это вероятность того, что контент является NSFW (от 0.0 до 1.0).

