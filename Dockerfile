# Stage 1 (Build)

FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o bin/kvgo-server ./cmd/kvgo-server/

# Stage 2 (Run)

FROM alpine:latest AS runtime

WORKDIR /app

COPY --from=builder /app/bin/kvgo-server .

EXPOSE 6379

ENTRYPOINT ["./kvgo-server"]