ARG GO_VERSION=1.22
FROM golang:${GO_VERSION}-alpine AS builder

# sqlite3 (cgo)
RUN apk add --no-cache ca-certificates git gcc musl-dev

WORKDIR /src
COPY . .

# Установка
RUN go mod download
RUN CGO_ENABLED=1 go build -o /build/app ./cmd/server/server.go

# Образ
FROM alpine:3.20 AS final

RUN apk add --no-cache ca-certificates sqlite

COPY --from=builder /build/app /app

# Делаем рабочую директорию
ENV HOME=/data
WORKDIR $HOME
VOLUME $HOME

# Только на локалхост (для реверс-прокси)
EXPOSE 127.0.0.1:8082

# Запуск
CMD ["/app", "-addr=127.0.0.1:8082", "-path=/data/servers.db"]
