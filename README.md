# upsync

Утилита для синхронизации медиафайлов (фото, видео) из различных источников в [Immich](https://immich.app/).

## Возможности

- **Загрузка из FTP** — подключение к FTP-серверу и загрузка файлов
- **Загрузка из локальной директории** — рекурсивный обход папок с фильтрацией по расширениям файлов
- **Загрузка в Immich** — через REST API (multipart upload) с корректной обработкой дубликатов
- **Воркер-пул** — параллельная загрузка файлов с настраиваемым количеством воркеров
- **Дедупликация** — пропуск уже загруженных файлов на основе списка из Immich
- **Graceful shutdown** — корректное завершение работы по сигналу (Ctrl+C, SIGTERM)
- **Логирование** — одновременный вывод в терминал (цветной) и в JSON-файл

## Архитектура

```
┌─────────────────────────────────────────────────────────────┐
│                        cmd/upsync                           │
│                    точка входа / main()                     │
└──────────────────────┬──────────────────────────────────────┘
                       │
                       ▼
┌─────────────────────────────────────────────────────────────┐
│                   internal/core/config                      │
│              загрузка конфига из .env                        │
└─────────────────────────────────────────────────────────────┘
                       │
                       ▼
┌─────────────────────────────────────────────────────────────┐
│                   internal/core/upsync                      │
│              основной движок синхронизации                   │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌──────────┐   │
│  │ worker 0 │  │ worker 1 │  │ worker 2 │  │ worker 3 │   │
│  └──────────┘  └──────────┘  └──────────┘  └──────────┘   │
│              worker pool (каналы + горутины)                 │
└──────┬──────────────────────────────────┬───────────────────┘
       │                                  │
       ▼                                  ▼
┌──────────────┐                ┌──────────────────┐
│  Источники   │                │   Хранилище      │
│  (Remote)    │                │   (Store)        │
│              │                │                  │
│  ┌────────┐  │                │  ┌────────────┐  │
│  │  FTP   │  │                │  │  Immich    │  │
│  └────────┘  │                │  └────────────┘  │
│  ┌────────┐  │                └──────────────────┘
│  │  Dir   │  │
│  └────────┘  │
└──────────────┘
```

### Компоненты

| Компонент | Описание |
|-----------|----------|
| [`cmd/upsync/upsync.go`](cmd/upsync/upsync.go) | Точка входа: инициализация конфига, логгера, хранилища и запуск синхронизации |
| [`internal/core/config`](internal/core/config/config.go) | Загрузка конфигурации из `.env` и переменных окружения |
| [`internal/core/upsync`](internal/core/upsync/upsync.go) | Основной движок: читает задачи, сравнивает списки файлов, распределяет работу по воркерам |
| [`internal/adapter/uploader/ftploader`](internal/adapter/uploader/ftploader/ftploader.go) | Загрузчик файлов с FTP-сервера |
| [`internal/adapter/uploader/directory`](internal/adapter/uploader/directory/directory.go) | Загрузчик файлов из локальной файловой системы (рекурсивно) |
| [`internal/adapter/uploader/immich`](internal/adapter/uploader/immich/immich.go) | Загрузчик файлов в Immich через REST API |
| [`internal/adapter/storage/jstore`](internal/adapter/storage/jstore/jstore.go) | Чтение конфигурации задач из JSON-файла |
| [`internal/adapter/logger`](internal/adapter/logger/logger.go) | Логгер на базе zap: терминал (цветной) + JSON-файл |

## Быстрый старт

### 1. Настройка

Создайте файл `.env` в корне проекта:

```env
LOG_LEVEL=info
LOG_PATH=./logs

UPSYNC_WORKER_COUNT=10
```

Создайте файл `data.remotes.json` с описанием задач синхронизации:

```json
[
    {
        "title": "from ftp server",
        "remote": {
            "protocol": "ftp",
            "host": "192.168.1.100",
            "port": 21,
            "login": "user",
            "password": "password",
            "path": "/DCIM/Camera"
        },
        "store": {
            "protocol": "immich",
            "address": "http://192.168.1.100:2283",
            "apiKey": "YOUR_IMMICH_API_KEY",
            "path": "upsync"
        }
    },
    {
        "title": "from directory",
        "remote": {
            "protocol": "dir",
            "path": "/path/to/export",
            "extensions": "jpg,mp4"
        },
        "store": {
            "protocol": "immich",
            "address": "http://192.168.1.100:2283",
            "apiKey": "YOUR_IMMICH_API_KEY",
            "path": "upsync"
        }
    }
]
```

### 2. Сборка

**Linux:**
```bash
make build-linux
```

**Windows:**
```bash
make build-windows
```

Или напрямую:

```bash
GOOS=linux GOARCH=amd64 go build -o ./upsync ./cmd/upsync/upsync.go
```

### 3. Запуск

```bash
./upsync
```

Или через `go run`:

```bash
make run
```

## Конфигурация

### Переменные окружения (`.env`)

| Переменная | Значение по умолчанию | Описание |
|-----------|----------------------|----------|
| `LOG_LEVEL` | `info` | Уровень логирования (`debug`, `info`, `warn`, `error`) |
| `LOG_PATH` | `./logs` | Директория для файла логов |
| `UPSYNC_WORKER_COUNT` | `1` | Количество воркеров для параллельной загрузки |

### Файл задач (`data.remotes.json`)

Массив объектов, каждый из которых описывает одну задачу синхронизации.

#### Поля задачи

| Поле | Тип | Обязательное | Описание |
|------|-----|-------------|----------|
| `title` | `string` | нет | Название задачи (для логирования) |
| `remote.protocol` | `"dir"` \| `"ftp"` | да | Протокол источника |
| `remote.host` | `string` | для FTP | Адрес FTP-сервера |
| `remote.port` | `int` | для FTP | Порт FTP-сервера |
| `remote.login` | `string` | для FTP | Логин FTP |
| `remote.password` | `string` | для FTP | Пароль FTP |
| `remote.path` | `string` | да | Путь к файлам на источнике |
| `remote.extensions` | `string` | нет | Фильтр расширений через запятую (только для `dir`), например `"jpg,mp4"` |
| `store.protocol` | `"immich"` | да | Протокол целевого хранилища |
| `store.address` | `string` | да | Адрес Immich-сервера, например `http://localhost:2283` |
| `store.apiKey` | `string` | да | API-ключ Immich |
| `store.path` | `string` | да | Идентификатор устройства (`deviceId`) в Immich. Используется как имя устройства-отправителя и для формирования `deviceAssetId` в формате `{deviceId}-{filename}-{unix_timestamp}`. Это позволяет Immich корректно определять дубликаты — если файл с таким же `deviceAssetId` уже существует, он не будет загружен повторно. Рекомендуется указывать имя пользователя или понятный идентификатор, например `alice` или `phone-android`. |

## Запуск по расписанию (cron)

```cron
25 * * * * cd /your/directory/upsync && ./upsync >> /your/directory/upsync/logs/cron.log
```

## Разработка

```bash
make run
```

## Структура проекта

```
upsync/
├── cmd/upsync/              # точка входа
├── internal/
│   ├── adapter/
│   │   ├── logger/          # логгер (zap)
│   │   ├── models/          # модели данных
│   │   ├── storage/         # хранилище конфигурации
│   │   │   └── jstore/      #   чтение из JSON
│   │   └── uploader/        # загрузчики
│   │       ├── ftploader/   #   FTP
│   │       ├── directory/   #   локальная директория
│   │       └── immich/      #   Immich API
│   └── core/
│       ├── config/          # конфиг приложения
│       └── upsync/          # движок синхронизации
├── deploy/                  # Docker Compose для Immich
├── scripts/                 # вспомогательные скрипты
├── template/                # шаблоны
├── Makefile                 # цели сборки
└── go.mod / go.sum          # зависимости