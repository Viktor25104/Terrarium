FROM golang:bookworm AS builder

# Установка заголовков libgpiod для компиляции CGO
RUN apt-get update && apt-get install -y libgpiod-dev build-essential

WORKDIR /app

# Кэширование зависимостей
COPY go.mod go.sum* ./
RUN go mod download || true # Игнорировать, если файла gomod еще нет

COPY . .

# Сборка бинарного файла с включенным CGo (Обязательно для нативного libgpiod)
ENV CGO_ENABLED=1
ENV GOOS=linux
ENV GOARCH=arm64

# Предполагаем, что main находится в cmd/server/main.go
RUN go build -ldflags="-w -s" -o terrarium-server ./cmd/server

# Финальная стадия для production
FROM debian:bookworm-slim

# Требуется динамическая библиотека libgpiod во время выполнения, плюс сертификаты для API Telegram
RUN apt-get update && apt-get install -y \
    libgpiod2 \
    ca-certificates \
    tzdata \
    python3 \
    python3-pip \
    python3-venv \
    && rm -rf /var/lib/apt/lists/*

# Создаем виртуальное окружение для Python и ставим нужные пакеты
RUN python3 -m venv /opt/venv
ENV PATH="/opt/venv/bin:$PATH"
RUN pip install --no-cache-dir adafruit-circuitpython-dht RPi.GPIO
# Установка правильного часового пояса для алгоритмов расписания освещения
ENV TZ=Europe/London

WORKDIR /app

COPY --from=builder /app/terrarium-server .
RUN mkdir -p /app/state

EXPOSE 8080

CMD ["./terrarium-server"]
