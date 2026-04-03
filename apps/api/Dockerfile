# ---- Development stage ----
FROM golang:1.22-alpine AS development

RUN apk add --no-cache git curl

# Install air for hot reload
RUN go install github.com/air-verse/air@latest

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

EXPOSE 8000

CMD ["air", "-c", ".air.toml"]

# ---- Builder stage ----
FROM golang:1.22-alpine AS builder

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
