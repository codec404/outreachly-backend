# ---- dev stage ----
FROM golang:1.25-alpine3.21 AS dev

WORKDIR /app

RUN go install github.com/air-verse/air@v1.64.5

COPY go.mod go.sum ./
RUN go mod download

COPY . .

EXPOSE 8080
CMD ["air", "-c", ".air.toml"]


# ---- builder stage ----
FROM golang:1.25-alpine3.21 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /app/bin/chat-service .


# ---- prod stage ----
FROM alpine:3.21 AS prod

RUN apk --no-cache add ca-certificates && \
    addgroup -S appgroup && adduser -S appuser -G appgroup

WORKDIR /app

COPY --from=builder /app/bin/chat-service .
COPY --from=builder /app/configs ./configs

RUN chown -R appuser:appgroup /app

USER appuser

EXPOSE 8080
CMD ["./chat-service"]
