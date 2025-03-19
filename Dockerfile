FROM golang:1.20-alpine AS builder

WORKDIR /app
COPY . .
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o music-api ./cmd/server/

FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/music-api .
COPY --from=builder /app/migrations ./migrations
COPY --from=builder /app/.env .

EXPOSE 8080
CMD ["./music-api"]