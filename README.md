# загрузка ваших медиа фалов в Immich

### enviorements
```env
LOG_LEVEL=info
LOG_PATH=./logs

UPSYNC_WOKRER_COUNT=10
```

### configure
разместите файла `data.remotes.json` в корне проекта.
```json
[
    {
        "title": "from ftp server",
        "remote": {
            "protocol": "ftp",
            "host": "192.168.0.110",
            "port": 2221,
            "login": "android",
            "password": "android",
            "path": "/DCIM/Camera"
        },
        "store": {
            "protocol": "immich",
            "address": "http://localhost:2283",
            "apiKey": "TcnCqCgplW7LXV4yc9qKfW6q9SbCL4AKkHxgabClQ",
            "path": "upsync"
        }
    },
    {
        "title": "from directory",
        "remote": {
            "protocol": "dir",
            "path": "F:/Downloads/Takeout",
            "extensions": "jpg,mp4"
        },
        "store": {
            "protocol": "immich",
            "address": "http://localhost:2283",
            "apiKey": "TcnCqCgplW7LXV4yc9qKfW6q9SbCL4AKkHxgabClQ",
            "path": "upsync"
        }
    },
    ...
]
```