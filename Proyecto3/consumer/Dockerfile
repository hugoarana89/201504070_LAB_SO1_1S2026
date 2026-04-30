FROM golang:latest AS builder
WORKDIR /app
COPY . .
RUN go mod tidy
RUN go mod download
RUN go build -o consumer .

FROM debian:bookworm-slim
RUN apt-get update && apt-get install -y ca-certificates && rm -rf /var/lib/apt/lists/*
WORKDIR /app
COPY --from=builder /app/consumer .
CMD ["./consumer"]