FROM golang:latest AS builder
WORKDIR /app

RUN apt-get update && apt-get install -y protobuf-compiler
RUN go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.35.0
RUN go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.5.1

COPY . .

RUN protoc --go_out=. --go-grpc_out=. proto/war_report.proto

RUN go mod tidy
RUN go mod download

RUN go build -o grpc-server .

FROM debian:bookworm-slim
RUN apt-get update && apt-get install -y ca-certificates && rm -rf /var/lib/apt/lists/*
WORKDIR /app
COPY --from=builder /app/grpc-server .
CMD ["./grpc-server"]