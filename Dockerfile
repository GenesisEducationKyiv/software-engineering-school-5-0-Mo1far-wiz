# S1
FROM golang:1.24-alpine AS builder

RUN apk add --no-cache git

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 \
    GOOS=linux \
    go build \
      -ldflags="-s -w" \
      -o /bin/weather-service \
      ./cmd/main.go

# S2
FROM alpine:3.19

RUN apk add --no-cache ca-certificates tzdata

RUN adduser -D -h /home/appuser appuser

WORKDIR /home/appuser

COPY --from=builder /bin/weather-service /usr/local/bin/weather-service

RUN chown -R appuser:appuser /home/appuser

USER appuser

EXPOSE 8080

ENV APP_ENV=production

ENTRYPOINT ["weather-service"]