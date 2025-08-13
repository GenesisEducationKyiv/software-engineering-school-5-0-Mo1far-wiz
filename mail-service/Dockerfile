# Build stage: Compiling the Go application
FROM golang:1.24-alpine AS builder

RUN apk add --no-cache git

WORKDIR /src

COPY go.work ./
COPY pkg/go.mod                                pkg/go.sum                                ./pkg/
COPY mail-service/go.mod                                                                 ./mail-service/
COPY weather-subscription-service/go.mod       weather-subscription-service/go.sum       ./weather-subscription-service/

RUN go work sync
RUN cd mail-service && go mod download

COPY pkg                          ./pkg
COPY mail-service                 ./mail-service
COPY weather-subscription-service ./weather-subscription-service

RUN CGO_ENABLED=0 \
    GOOS=linux \
    go build \
      -ldflags="-s -w" \
      -o /bin/mail-service \
      ./mail-service/cmd/main.go

# Build stage: Copy Go build
FROM alpine:3.19

RUN apk add --no-cache ca-certificates tzdata
RUN adduser -D -h /home/appuser appuser

WORKDIR /home/appuser

COPY --from=builder /bin/mail-service /usr/local/bin/mail-service
RUN chown -R appuser:appuser /home/appuser

USER appuser

ENTRYPOINT ["mail-service"]