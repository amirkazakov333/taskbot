# ==============================
# Stage 1 — сборка Go
# ==============================

FROM golang:alpine AS builder

WORKDIR /build

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 \
    GOOS=linux \
    go build \
    -o /out/taskbot \
    ./cmd/taskbot


# ==============================
# Stage 2 — runtime
# ==============================

FROM alpine:latest

RUN apk add --no-cache \
    postgresql \
    postgresql-client \
    su-exec \
    ca-certificates

WORKDIR /app

COPY --from=builder /out/taskbot /app/taskbot

COPY schema.sql /app/schema.sql

COPY docker/entrypoint.sh /app/entrypoint.sh

RUN chmod +x /app/entrypoint.sh

ENTRYPOINT ["/app/entrypoint.sh"]