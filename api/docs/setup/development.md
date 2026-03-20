# Настройка среды разработки

## Предварительные требования

- Go 1.22+ (рекомендуется 1.25+ для Fiber v3)
- Docker и Docker Compose (для баз данных)
- Git

## Установка Go

```bash
curl -OL https://go.dev/dl/go1.24.1.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.24.1.linux-amd64.tar.gz
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
source ~/.bashrc
go version
```

## Первый запуск

1. Запустить инфраструктуру баз данных:

   ```bash
   cd /путь/к/проекту
   docker compose up -d
   ```

2. Дождаться готовности контейнеров (healthcheck):

   ```bash
   docker compose ps
   ```

3. Установить зависимости Go:

   ```bash
   cd api
   go mod tidy
   ```

4. Запустить сервер:

   ```bash
   go run ./cmd/api
   ```

5. Проверить работу:
   ```bash
   curl http://localhost:8080/api/v1/health | jq .
   ```

## Тестирование

```bash
cd api

# Запуск всех тестов
go test ./... -v

# Запуск тестов конкретного пакета
go test ./internal/config/... -v

# С подсчётом покрытия
go test ./... -cover
```

## Структура конфигурации

Конфигурация загружается из файла `.env` в корне проекта.
Переменные окружения имеют приоритет над значениями из файла.

Путь к `.env` задаётся в `cmd/api/main.go` (по умолчанию `../.env`,
так как приложение запускается из директории `api/`).

## Сборка

```bash
cd api
go build -o ./bin/deep-krai-api ./cmd/api
```
