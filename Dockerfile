# syntax=docker/dockerfile:1

ARG GO_VERSION=1.22
FROM golang:${GO_VERSION}-alpine AS builder

# sqlite требует CGO
RUN apk add --no-cache ca-certificates git gcc musl-dev

WORKDIR /src

# разделяем go.mod и go.sum отдельно для улучшения кэша
COPY go.mod go.sum ./

# качаем зависимости
RUN go mod tidy && go mod download

# копипастим исходники
COPY . .

# собираем sqlite3 ибо он требует CGO_ENABLED=1 (я это говно под mysql перепишу позже)
RUN CGO_ENABLED=1 go build -o /build/app ./cmd/server/server.go

################################################################################

FROM alpine:3.20 AS final

RUN apk add --no-cache ca-certificates sqlite-libs

COPY --from=builder /build /build

ENV HOME=/data
WORKDIR $HOME
VOLUME $HOME

EXPOSE 8082

# стартуем на локалке
CMD ["/build/app", "-addr=127.0.0.1:8082", "-path=/data/servers.db"]
