#!/bin/bash
# Скрипт для ручной загрузки файла в Immich через API
# Используется для диагностики проблем с загрузкой
#
# Использование:
#   ./scripts/curl-immich-upload.sh -u <url> -k <api-key> <файл> [device_id]
#
# Пример:
#   ./scripts/curl-immich-upload.sh -u http://192.168.1.100:2283 -k secret123 photo.jpg
#   ./scripts/curl-immich-upload.sh -u http://192.168.1.100:2283 -k secret123 photo.jpg "my-camera"
#
# Переменные окружения (альтернатива флагам):
#   IMMICH_URL      - адрес сервера Immich
#   IMMICH_API_KEY  - API-ключ для аутентификации

set -euo pipefail

# Парсинг флагов
while getopts "u:k:h" opt; do
    case $opt in
        u) IMMICH_URL="$OPTARG" ;;
        k) IMMICH_API_KEY="$OPTARG" ;;
        h)
            echo "Использование: $0 -u <url> -k <api-key> <файл> [device_id]"
            echo ""
            echo "Флаги:"
            echo "  -u <url>      Адрес сервера Immich (например http://192.168.1.100:2283)"
            echo "  -k <api-key>  API-ключ Immich"
            echo "  -h            Показать эту справку"
            echo ""
            echo "Переменные окружения:"
            echo "  IMMICH_URL      альтернатива флагу -u"
            echo "  IMMICH_API_KEY  альтернатива флагу -k"
            echo ""
            echo "Пример:"
            echo "  $0 -u http://192.168.1.100:2283 -k secret123 photo.jpg"
            exit 0
            ;;
        *) 
            echo "Использование: $0 -u <url> -k <api-key> <файл> [device_id]"
            exit 1
            ;;
    esac
done
shift $((OPTIND - 1))

IMMICH_URL="${IMMICH_URL:-}"
IMMICH_API_KEY="${IMMICH_API_KEY:-}"

# Проверяем обязательные параметры
if [ -z "$IMMICH_URL" ]; then
    echo "Ошибка: не указан адрес сервера. Используйте флаг -u или переменную IMMICH_URL"
    exit 1
fi

if [ -z "$IMMICH_API_KEY" ]; then
    echo "Ошибка: не указан API-ключ. Используйте флаг -k или переменную IMMICH_API_KEY"
    exit 1
fi

FILE="${1:-}"
DEVICE_ID="${2:-upsync-cli}"

if [ -z "$FILE" ]; then
    echo "Ошибка: укажите путь к файлу для загрузки"
    echo "Использование: $0 -u <url> -k <api-key> <файл> [device_id]"
    exit 1
fi

if [ ! -f "$FILE" ]; then
    echo "Ошибка: файл '$FILE' не найден"
    exit 1
fi

FILENAME=$(basename "$FILE")
TIMESTAMP=$(date +%s)
DEVICE_ASSET_ID="${DEVICE_ID}-${FILENAME}-${TIMESTAMP}"
FILE_MOD_TIME=$(date -r "$FILE" '+%Y-%m-%d %H:%M:%S' 2>/dev/null || stat -f "%Sm" -t "%Y-%m-%d %H:%M:%S" "$FILE" 2>/dev/null || echo "$(date '+%Y-%m-%d %H:%M:%S')")

echo "=== Параметры запроса ==="
echo "URL:              ${IMMICH_URL}/api/assets"
echo "Device ID:        ${DEVICE_ID}"
echo "Device Asset ID:  ${DEVICE_ASSET_ID}"
echo "File:             ${FILE}"
echo "File size:        $(wc -c < "$FILE") bytes"
echo "File created at:  ${FILE_MOD_TIME}"
echo "========================"
echo ""

# Выполняем запрос в точности как в Go-коде upsync
curl -v --request POST \
    "${IMMICH_URL}/api/assets" \
    --header "x-api-key: ${IMMICH_API_KEY}" \
    --form "assetData=@${FILE}" \
    --form "deviceAssetId=${DEVICE_ASSET_ID}" \
    --form "deviceId=${DEVICE_ID}" \
    --form "fileCreatedAt=${FILE_MOD_TIME}" \
    --form "fileModifiedAt=${FILE_MOD_TIME}" \
    --connect-timeout 10 \
    --max-time 60

echo ""
echo "=== Запрос завершён ==="
echo ""
echo "Ожидаемые коды ответа:"
echo "  201 Created - файл успешно загружен"
echo "  200 OK      - дубликат (файл уже существует)"
echo "  Любой другой код - ошибка"