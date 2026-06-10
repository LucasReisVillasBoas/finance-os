# ---- Development stage ----
FROM golang:1.25-alpine AS development

RUN apk add --no-cache git curl

# Try to install air for hot reload; fall back gracefully if network unavailable
RUN go install github.com/air-verse/air@latest || echo "air not available, will use direct go build"

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

EXPOSE 8000

# Use air if available, otherwise fall back to direct go build + run
CMD sh -c 'if command -v air > /dev/null 2>&1; then air -c .air.toml; else go build -o /tmp/financeos-api ./cmd/server && /tmp/financeos-api; fi'

# ---- Builder stage ----
FROM golang:1.25-alpine AS builder

RUN apk add --no-cache git

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-w -s" \
    -o bin/server \
    ./cmd/server

# ---- Production stage ----
FROM alpine:3.19 AS production

RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app

COPY --from=builder /app/bin/server .

EXPOSE 8000

USER nobody:nobody

ENTRYPOINT ["/app/server"]
