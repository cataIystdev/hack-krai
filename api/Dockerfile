# Сборка Go API
FROM golang:1.25.0-alpine AS builder

# Установка необходимых пакетов
RUN apk add --no-cache git tzdata

WORKDIR /app

# Кэширование модулей
COPY go.mod go.sum ./
RUN go mod download

# Копирование исходного кода
COPY . .

# Сборка бинарного файла
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo -o deepkrai-api ./cmd/api/main.go

# Минимальный рантайм
FROM alpine:latest

WORKDIR /app

RUN apk --no-cache add ca-certificates tzdata

# Копируем бинарник из билдера
COPY --from=builder /app/deepkrai-api .

# По умолчанию API работает на порту 8080 (или берется из .env)
EXPOSE 8080

CMD ["./deepkrai-api"]
