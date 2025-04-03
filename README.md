- postgres

### Запуск
Чтобы запустить проект необходимо указать переменные окружения в .env файле

```
export DB_HOST=localhost
export DB_PORT=5432
export DB_USERNAME=postgres
export DB_NAME=postgres
export DB_SSLMODE=disable
export DB_PASSWORD=qwertyuiop
```

Сборка и запуск
```
source .env && go build -o app cmd/main.go && ./app
```