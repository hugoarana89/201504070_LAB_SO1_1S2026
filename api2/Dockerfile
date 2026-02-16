FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY go.mod ./
RUN go mod download
COPY . .
RUN go build -o api2

FROM alpine:latest
WORKDIR /root/
COPY --from=builder /app/api2 .
EXPOSE 8080
CMD ["./api2"]
