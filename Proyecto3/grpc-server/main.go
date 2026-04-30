package main

import (
	"context"
	"encoding/json"
	"log"
	"net"
	"os"
	"time"

	pb "grpc-server/proto"

	amqp "github.com/rabbitmq/amqp091-go"
	"google.golang.org/grpc"
)

// Estructura para publicar en RabbitMQ
type WarReportMessage struct {
	Country         string `json:"country"`
	WarplanesInAir  int32  `json:"warplanes_in_air"`
	WarshipsInWater int32  `json:"warships_in_water"`
	Timestamp       string `json:"timestamp"`
}

// Servidor gRPC
type warReportServer struct {
	pb.UnimplementedWarReportServiceServer
	rabbitChannel *amqp.Channel
	queueName     string
}

// Implementación del método gRPC SendReport
func (s *warReportServer) SendReport(ctx context.Context, req *pb.WarReportRequest) (*pb.WarReportResponse, error) {
	log.Printf("Reporte gRPC recibido: país=%s aviones=%d barcos=%d",
		req.Country.String(), req.WarplanesInAir, req.WarshipsInWater)

	// Convertir enum de país a string
	countryStr := req.Country.String()

	// Construir mensaje para RabbitMQ
	message := WarReportMessage{
		Country:         countryStr,
		WarplanesInAir:  req.WarplanesInAir,
		WarshipsInWater: req.WarshipsInWater,
		Timestamp:       req.Timestamp,
	}

	// Serializar a JSON
	body, err := json.Marshal(message)
	if err != nil {
		log.Printf("Error serializando mensaje: %v", err)
		return &pb.WarReportResponse{Status: "error"}, err
	}

	// Publicar en RabbitMQ
	err = s.rabbitChannel.Publish(
		"",          // exchange
		s.queueName, // routing key (nombre de la queue)
		false,       // mandatory
		false,       // immediate
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
		},
	)
	if err != nil {
		log.Printf("Error publicando en RabbitMQ: %v", err)
		return &pb.WarReportResponse{Status: "error"}, err
	}

	log.Printf("Mensaje publicado en RabbitMQ: %s", string(body))
	return &pb.WarReportResponse{Status: "ok"}, nil
}

func main() {
	// Variables de entorno
	rabbitURL := os.Getenv("RABBITMQ_URL")
	if rabbitURL == "" {
		rabbitURL = "amqp://guest:guest@rabbitmq:5672/"
	}

	queueName := os.Getenv("QUEUE_NAME")
	if queueName == "" {
		queueName = "war_reports"
	}

	grpcPort := os.Getenv("GRPC_PORT")
	if grpcPort == "" {
		grpcPort = "50051"
	}

	// Conectar a RabbitMQ con reintentos
	log.Printf("Conectando a RabbitMQ en: %s", rabbitURL)
	var conn *amqp.Connection
	var err error

	for i := 0; i < 10; i++ {
		log.Printf("Intento %d de conexión a RabbitMQ...", i+1)
		conn, err = amqp.Dial(rabbitURL)
		if err == nil {
			break
		}
		log.Printf("RabbitMQ no disponible, reintentando en 3 segundos: %v", err)
		time.Sleep(3 * time.Second)
	}

	if err != nil {
		log.Fatalf("No se pudo conectar a RabbitMQ después de 10 intentos: %v", err)
	}
	defer conn.Close()

	channel, err := conn.Channel()
	if err != nil {
		log.Fatalf("Error abriendo canal RabbitMQ: %v", err)
	}
	defer channel.Close()

	// Declarar la queue (la crea si no existe)
	_, err = channel.QueueDeclare(
		queueName, // nombre
		true,      // durable
		false,     // auto-delete
		false,     // exclusive
		false,     // no-wait
		nil,       // args
	)
	if err != nil {
		log.Fatalf("Error declarando queue: %v", err)
	}

	log.Printf("Queue '%s' lista", queueName)

	// Iniciar servidor gRPC
	listener, err := net.Listen("tcp", ":"+grpcPort)
	if err != nil {
		log.Fatalf("Error iniciando listener: %v", err)
	}

	server := grpc.NewServer()
	pb.RegisterWarReportServiceServer(server, &warReportServer{
		rabbitChannel: channel,
		queueName:     queueName,
	})

	log.Printf("gRPC server escuchando en puerto %s", grpcPort)
	if err := server.Serve(listener); err != nil {
		log.Fatalf("Error en gRPC server: %v", err)
	}
}