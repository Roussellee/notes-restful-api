FROM golang:1.24.2

WORKDIR /app

# Копируем go.mod и go.sum
COPY go.mod go.sum ./
RUN go mod download

# Копируем все необходимые файлы
COPY cmd/ ./cmd/
COPY internal/ ./internal/
COPY docs/ ./docs/

# Устанавливаем команду для запуска приложения
CMD ["go", "run", "cmd/main.go"]
