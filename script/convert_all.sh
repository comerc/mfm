#!/usr/bin/env bash
set -euo pipefail

# ============================================================================
# convert_all.sh — Пакетная конвертация всех MP4 в папке
# ============================================================================
show_help() {
  cat << EOF
USAGE:
  $(basename "$0") <путь_к_папке> [ОПЦИИ]

ОПЦИИ:
  -r, --recursive   Рекурсивный поиск во вложенных папках
  -y, --overwrite   Перезаписывать существующие файлы без подтверждения
  -n, --dry-run     Тестовый запуск: показать что будет сделано (без конвертации)
  -h, --help        Эта справка

ПРИМЕРЫ:
  ./batch_convert_mp4.sh ./videos
  ./batch_convert_mp4.sh ./videos -r -y
  ./batch_convert_mp4.sh ./videos --dry-run
EOF
  exit 0
}

# ============================================================================
# Парсинг аргументов
# ============================================================================
RECURSIVE=false
OVERWRITE=false
DRY_RUN=false
TARGET_DIR=""

while [[ $# -gt 0 ]]; do
  case "$1" in
    -r|--recursive)  RECURSIVE=true; shift ;;
    -y|--yes|--overwrite) OVERWRITE=true; shift ;;
    -n|--dry-run)    DRY_RUN=true; shift ;;
    -h|--help)       show_help ;;
    -*)
      echo "❌ Неизвестный параметр: $1"
      echo "Используйте --help для справки."
      exit 1
      ;;
    *) TARGET_DIR="$1"; shift ;;
  esac
done

# Проверка обязательного аргумента
if [[ -z "$TARGET_DIR" ]]; then
  echo "❌ Укажите путь к папке с видео"
  echo "Используйте --help для справки."
  exit 1
fi

# Валидация папки
if [[ ! -d "$TARGET_DIR" ]]; then
  echo "❌ Папка не найдена: $TARGET_DIR"
  exit 1
fi

TARGET_DIR="$(cd "$TARGET_DIR" && pwd)"  # абсолютный путь

# Проверка внешнего скрипта при необходимости
CONVERT_CMD="__internal__"

# ============================================================================
# Встроенный конвертер
# ============================================================================
internal_convert() {
  local INPUT="$1"
  local OVERWRITE="$2"
  local BASENAME="${INPUT%.*}"
  
  for ext in avi mov mkv webm; do
    OUTPUT="${BASENAME}.${ext}"
    
    # Пропуск существующих файлов
    if [[ -f "$OUTPUT" && "$OVERWRITE" == false ]]; then
      echo "⏭️  Пропущено (уже существует): $OUTPUT"
      continue
    fi
    
    # Определение параметров кодеков
    local FFMPEG_ARGS=()
    
    case "$ext" in
      avi)
        FFMPEG_ARGS=(-c:v mpeg4 -qscale:v 4 -c:a libmp3lame -q:a 4)
        ;;
      mov)
        FFMPEG_ARGS=(-c:v libx264 -preset ultrafast -crf 28 -c:a copy)
        ;;
      mkv)
        FFMPEG_ARGS=(-c copy)
        ;;
      webm)
        FFMPEG_ARGS=(-c:v libvpx -deadline realtime -cpu-used 8 -b:v 3M -c:a libopus -b:a 96k)
        ;;
    esac
    
    # Dry-run
    if [[ "$DRY_RUN" == true ]]; then
      echo "🧪 [DRY-RUN] ffmpeg -y -i \"$INPUT\" ${FFMPEG_ARGS[*]} \"$OUTPUT\""
      continue
    fi
    
    # Конвертация
    echo -n "  🔄 $ext ... "
    if ffmpeg -y -i "$INPUT" "${FFMPEG_ARGS[@]}" "$OUTPUT" &>/dev/null; then
      echo "✅ $(du -h "$OUTPUT" | cut -f1)"
    else
      echo "❌ ОШИБКА"
      return 1
    fi
  done
}

# ============================================================================
# Поиск файлов
# ============================================================================
echo "=========================================="
echo "📁 Папка: $TARGET_DIR"
echo "=========================================="

# Формирование массива файлов
mapfile -d '' MP4_FILES < <( \
  if [[ "$RECURSIVE" == true ]]; then
    find "$TARGET_DIR" -type f -iname "*.mp4" -print0
  else
    find "$TARGET_DIR" -maxdepth 1 -type f -iname "*.mp4" -print0
  fi \
)

if [[ ${#MP4_FILES[@]} -eq 0 ]]; then
  echo "⚠️  Не найдено файлов .mp4"
  if [[ "$RECURSIVE" == false ]]; then
    echo "💡 Совет: добавьте флаг -r для поиска во вложенных папках"
  fi
  exit 0
fi

echo "🎬 Найдено файлов: ${#MP4_FILES[@]}"
[[ "$RECURSIVE" == true ]] && echo "   (рекурсивный поиск)"

# ============================================================================
# Обработка файлов
# ============================================================================
TOTAL=${#MP4_FILES[@]}
COUNT=0
FAILURES=0

for FILE in "${MP4_FILES[@]}"; do
  COUNT=$((COUNT + 1))
  FILENAME="$(basename "$FILE")"
  
  echo
  echo "📄 [$COUNT/$TOTAL] $FILENAME"
  echo "   Путь: ${FILE#$TARGET_DIR/}"
  
  if [[ "$DRY_RUN" == true ]]; then
    internal_convert "$FILE" "$OVERWRITE"
    continue
  fi
  
  # Выбор метода конвертации
  if ! internal_convert "$FILE" "$OVERWRITE" 2>/dev/null; then
    FAILURES=$((FAILURES + 1))
  fi
done

# ============================================================================
# Итоги
# ============================================================================
echo
echo "=========================================="
echo "📊 ИТОГИ"
echo "=========================================="
echo "✅ Обработано: $((TOTAL - FAILURES)) из $TOTAL файлов"
[[ $FAILURES -gt 0 ]] && echo "❌ Ошибок: $FAILURES"

if [[ "$DRY_RUN" == true ]]; then
  echo
  echo "💡 Это был тестовый запуск (--dry-run)."
  echo "   Для реальной конвертации запустите без флага -n"
fi

exit $((FAILURES > 0 ? 1 : 0))