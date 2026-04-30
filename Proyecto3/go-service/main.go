package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	pb "go-service/proto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// Estructura que llega desde Rust
type WarReport struct {
	Country          string `json:"country"`
	WarplanesInAir   int32  `json:"warplanes_in_air"`
	WarshipsInWater  int32  `json:"warships_in_water"`
	Timestamp        string `json:"timestamp"`
}

type ApiResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

// Mapeo de código de país a enum gRPC
func countryToEnum(country string) pb.Countries {
	switch strings.ToUpper(country) {
	case "USA":
		return pb.Countries_usa
	case "RUS":
		return pb.Countries_rus
	case "CHN":
		return pb.Countries_chn
	case "ESP":
		return pb.Countries_esp
	case "GTM":
		return pb.Countries_gtm
	default:
		return pb.Countries_countries_unknown
	}
}

func receiveReport(grpcClient pb.WarReportServiceClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		// Decodificar JSON que viene de Rust
		var report WarReport
		if err := json.NewDecoder(r.Body).Decode(&report); err != nil {
			log.Printf("Error decodificando JSON: %v", err)
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(ApiResponse{Status: "error", Message: "JSON inválido"})
			return
		}

		log.Printf("Reporte recibido de Rust: %+v", report)

		// Construir mensaje gRPC
		grpcRequest := &pb.WarReportRequest{
			Country:         countryToEnum(report.Country),
			WarplanesInAir:  report.WarplanesInAir,
			WarshipsInWater: report.WarshipsInWater,
			Timestamp:       report.Timestamp,
		}

		// Enviar via gRPC al grpc-server
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		response, err := grpcClient.SendReport(ctx, grpcRequest)
		if err != nil {
			log.Printf("Error enviando gRPC: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(ApiResponse{Status: "error", Message: "Error enviando a gRPC server"})
			return
		}

		log.Printf("Respuesta gRPC: %s", response.Status)
		json.NewEncoder(w).Encode(ApiResponse{Status: "ok", Message: response.Status})
	}
}

func health(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(ApiResponse{Status: "ok", Message: "go-service funcionando"})
}

func main() {
	// URL del grpc-server, configurable por variable de entorno
	grpcServerURL := os.Getenv("GRPC_SERVER_URL")
	if grpcServerURL == "" {
		grpcServerURL = "localhost:50051"
	}

	log.Printf("Conectando a gRPC server en: %s", grpcServerURL)

	// Conexión gRPC al servidor
	conn, err := grpc.Dial(grpcServerURL,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
		grpc.WithTimeout(10*time.Second),
	)
	if err != nil {
		log.Fatalf("No se pudo conectar al gRPC server: %v", err)
	}
	defer conn.Close()

	grpcClient := pb.NewWarReportServiceClient(conn)
	log.Println("Conexión gRPC establecida")

	// Servidor HTTP
	mux := http.NewServeMux()
	mux.HandleFunc("/report", receiveReport(grpcClient))
	mux.HandleFunc("/", health) // tiene que ser / no /health para que funcione con el gateway api

	log.Println("go-service escuchando en puerto 8081")
	if err := http.ListenAndServe(":8081", mux); err != nil {
		log.Fatalf("Error iniciando servidor: %v", err)
	}
}