FROM golang:latest AS builder
WORKDIR /app

# Instalar protoc y plugins con versiones fijas compatibles
RUN apt-get update && apt-get install -y protobuf-compiler
RUN go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.35.0
RUN go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.5.1

# Copiar todo el código
COPY . .

# Generar código proto
RUN protoc --go_out=. --go-grpc_out=. proto/war_report.proto

# Tidy y download
RUN go mod tidy
RUN go mod download

# Compilar
RUN go build -o go-service .

FROM debian:bookworm-slim
RUN apt-get update && apt-get install -y ca-certificates && rm -rf /var/lib/apt/lists/*
WORKDIR /app
COPY --from=builder /app/go-service .
CMD ["./go-service"]