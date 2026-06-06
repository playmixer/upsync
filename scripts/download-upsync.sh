#!/usr/bin/env bash
#
# download-upsync.sh
#
# Скачивает файл upsync по переданной ссылке и сохраняет его
# с именем upsync в текущей директории.
#
# Использование:
#   ./scripts/download-upsync.sh <URL>
#
# Пример:
#   ./scripts/download-upsync.sh https://example.com/path/to/upsync
#

set -euo pipefail

if [ $# -lt 1 ]; then
    echo "Ошибка: не указан URL для скачивания."
    echo "Использование: $0 <URL>"
    exit 1
fi

URL="$1"
OUTPUT="upsync"

echo "Скачивание $URL ..."

if command -v curl &>/dev/null; then
    curl -fsSL -o "$OUTPUT" "$URL"
elif command -v wget &>/dev/null; then
    wget -q -O "$OUTPUT" "$URL"
else
    echo "Ошибка: не найден ни curl, ни wget. Установите один из них."
    exit 1
fi

chmod +x "$OUTPUT"

echo "Готово: файл сохранён как ./$OUTPUT"