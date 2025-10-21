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

| attribute | values | requer | description |
|-|-|-|-|
| title | string | optional | title of peration |
| remote.protocol | [dir, ftp] | requer | protocol of remote |
| remote.host | ip or domane example "192.168.0.100" | optional | address of remote |
| remote.port | port as int example 2221 | optional | port of remote |
| remote.login | string | optional | login of remote |
| remote.password | string | optional | password of remote |
| remote.path | string | requer | path of remote |
| remote.extensions | example "jpg,mp4" | optional | access extensions of remote files (actual only fo "dir")|
| store.protocol | [immich] | requer | protocol of store |
| store.address | example "http://localhost:2283" | requer | address of store |
| store.apiKey | string | requer | api key of store |
| store.path | string | requer | path of store |

### Build
### windows
```bash
GOOS=windows GOARCH=386 go build -o ./upsync.exe ./cmd/upsync/upsync.go 
```
### ubuntu and similar os
```bash
GOOS=linux GOARCH=amd64 go build -o ./upsync ./cmd/upsync/upsync.go
```

### Run
**go run ./...**

OR exec

 **./upsync**

### Start on cron
```
25 * * * * cd /your/directory/upsync && ./upsync >> /your/directory/upsync/logs/cron.log
```