### Immich deploy

Скопируйте `docker-compose.yml` в нужную директорию
создайте `.env` рядом с ним и запустите команду `docker-compose up -d`

### env
```env
UPLOAD_LOCATION=./content

DB_DATA_LOCATION=./postgres

IMMICH_VERSION=release

DB_PASSWORD=postgres

DB_USERNAME=postgres
DB_DATABASE_NAME=immich
```

Подробная инструкция в [https://docs.immich.app/install/docker-compose/](https://docs.immich.app/install/docker-compose/)