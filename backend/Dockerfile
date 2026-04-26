# Stage 1: Build
FROM golang:1.24-alpine AS builder

RUN apk add --no-cache git ca-certificates tzdata

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Use real config if present (local build), otherwise fall back to example (CI)
RUN if [ ! -f config/config.yaml ]; then cp config/config.yaml.example config/config.yaml; fi

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o main ./cmd/main.go

# Stage 2: Run
FROM alpine:3.19

RUN apk --no-cache add ca-certificates tzdata curl

WORKDIR /app

# Binary
COPY --from=builder /app/main .

# Config
RUN mkdir -p config
COPY --from=builder /app/config/config.yaml ./config/

# Migrations (goose reads ./migrations at runtime)
COPY --from=builder /app/migrations ./migrations/

EXPOSE 8080 8081

CMD ["./main", "-config", "./config/config.yaml"]
