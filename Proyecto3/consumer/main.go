package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/redis/go-redis/v9"
)

// Estructura del mensaje que llega de RabbitMQ
type WarReportMessage struct {
	Country         string `json:"country"`
	WarplanesInAir  int32  `json:"warplanes_in_air"`
	WarshipsInWater int32  `json:"warships_in_water"`
	Timestamp       string `json:"timestamp"`
}

func connectRabbitMQ(url string) (*amqp.Connection, error) {
	var conn *amqp.Connection
	var err error
	for i := 0; i < 10; i++ {
		log.Printf("Intento %d de conexion a RabbitMQ...", i+1)
		conn, err = amqp.Dial(url)
		if err == nil {
			return conn, nil
		}
		log.Printf("RabbitMQ no disponible, reintentando en 3s: %v", err)
		time.Sleep(3 * time.Second)
	}
	return nil, err
}

func connectValkey(addr, password string) (*redis.Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       0,
	})
	for i := 0; i < 10; i++ {
		log.Printf("Intento %d de conexion a Valkey...", i+1)
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		_, err := client.Ping(ctx).Result()
		cancel()
		if err == nil {
			return client, nil
		}
		log.Printf("Valkey no disponible, reintentando en 3s: %v", err)
		time.Sleep(3 * time.Second)
	}
	return nil, fmt.Errorf("no se pudo conectar a Valkey")
}

func processMessage(ctx context.Context, msg WarReportMessage, valkey *redis.Client) error {
	// CRITICO: usar el pais TAL COMO LLEGA en el mensaje (minusculas: usa, russia, china...)
	// NO aplicar ToUpper ni ToLower, para que coincida con las keys de Grafana
	country := msg.Country

	log.Printf("Procesando: pais=%s aviones=%d barcos=%d", country, msg.WarplanesInAir, msg.WarshipsInWater)

	pipe := valkey.Pipeline()

	// ── 1. Contador de reportes por pais ─────────────────────────────────────
	// Grafana: GET reports:count:usa
	pipe.Incr(ctx, fmt.Sprintf("reports:count:%s", country))

	// ── 2. Ultimo reporte completo del pais ──────────────────────────────────
	reportJSON, _ := json.Marshal(msg)
	pipe.Set(ctx, fmt.Sprintf("reports:latest:%s", country), reportJSON, 24*time.Hour)

	// ── 3. TOP PAISES - Sorted Sets ───────────────────────────────────────────
	// Grafana ZRANGE: Key=top:warplanes  Score range  Min=0  Max=9999999999
	// Grafana ZRANGE: Key=top:warships   Score range  Min=0  Max=9999999999
	pipe.ZIncrBy(ctx, "top:warplanes", float64(msg.WarplanesInAir), country)
	pipe.ZIncrBy(ctx, "top:warships", float64(msg.WarshipsInWater), country)

	// ── 4. Maximos globales ───────────────────────────────────────────────────
	// Grafana: GET global:max:warplanes  y  GET global:max:warships
	currentMax, err := valkey.Get(ctx, "global:max:warplanes").Int()
	if err != nil || int(msg.WarplanesInAir) > currentMax {
		pipe.Set(ctx, "global:max:warplanes", msg.WarplanesInAir, 0)
	}
	currentMaxShips, err := valkey.Get(ctx, "global:max:warships").Int()
	if err != nil || int(msg.WarshipsInWater) > currentMaxShips {
		pipe.Set(ctx, "global:max:warships", msg.WarshipsInWater, 0)
	}

	// ── 5. Minimos globales ───────────────────────────────────────────────────
	// Grafana: GET global:min:warplanes  y  GET global:min:warships
	currentMin, err := valkey.Get(ctx, "global:min:warplanes").Int()
	if err != nil || int(msg.WarplanesInAir) < currentMin {
		pipe.Set(ctx, "global:min:warplanes", msg.WarplanesInAir, 0)
	}
	currentMinShips, err := valkey.Get(ctx, "global:min:warships").Int()
	if err != nil || int(msg.WarshipsInWater) < currentMinShips {
		pipe.Set(ctx, "global:min:warships", msg.WarshipsInWater, 0)
	}

	// ── 6. SERIE TEMPORAL - Stream por pais ───────────────────────────────────
	// Key: stream:usa  (minusculas, igual que country)
	// Grafana XREVRANGE: Key=stream:usa  Start=+  End=-  (SIN Count o Count=500)
	// Grafana mostrara dos series: warplanes_in_air y warships_in_water
	streamKey := fmt.Sprintf("stream:%s", country)
	pipe.XAdd(ctx, &redis.XAddArgs{
		Stream: streamKey,
		MaxLen: 500,
		Approx: true,
		Values: map[string]interface{}{
			"warplanes_in_air":  msg.WarplanesInAir,
			"warships_in_water": msg.WarshipsInWater,
			"timestamp":         msg.Timestamp,
		},
	})

	// ── 7. MODA - Sorted sets de frecuencias ──────────────────────────────────
	// Grafana: GET moda:warplanes:current  y  GET moda:warships:current
	pipe.ZIncrBy(ctx, "moda:warplanes", 1, fmt.Sprintf("%d", msg.WarplanesInAir))
	pipe.ZIncrBy(ctx, "moda:warships", 1, fmt.Sprintf("%d", msg.WarshipsInWater))

	// Ejecutar pipeline
	if _, err = pipe.Exec(ctx); err != nil {
		return fmt.Errorf("error ejecutando pipeline en Valkey: %v", err)
	}

	// ── 8. Calcular moda actual (fuera del pipeline) ──────────────────────────
	modaAviones, err := valkey.ZRevRangeWithScores(ctx, "moda:warplanes", 0, 0).Result()
	if err == nil && len(modaAviones) > 0 {
		valkey.Set(ctx, "moda:warplanes:current", modaAviones[0].Member, 0)
	}
	modaBarcos, err := valkey.ZRevRangeWithScores(ctx, "moda:warships", 0, 0).Result()
	if err == nil && len(modaBarcos) > 0 {
		valkey.Set(ctx, "moda:warships:current", modaBarcos[0].Member, 0)
	}

	log.Printf("OK pais=%s stream=%s", country, streamKey)
	return nil
}

func main() {
	rabbitURL := os.Getenv("RABBITMQ_URL")
	if rabbitURL == "" {
		rabbitURL = "amqp://guest:guest@rabbitmq:5672/"
	}
	queueName := os.Getenv("QUEUE_NAME")
	if queueName == "" {
		queueName = "war_reports"
	}
	valkeyAddr := os.Getenv("VALKEY_ADDR")
	if valkeyAddr == "" {
		valkeyAddr = "valkey:6379"
	}
	valkeyPassword := os.Getenv("VALKEY_PASSWORD")

	rabbitConn, err := connectRabbitMQ(rabbitURL)
	if err != nil {
		log.Fatalf("No se pudo conectar a RabbitMQ: %v", err)
	}
	defer rabbitConn.Close()
	log.Println("Conectado a RabbitMQ")

	valkeyClient, err := connectValkey(valkeyAddr, valkeyPassword)
	if err != nil {
		log.Fatalf("No se pudo conectar a Valkey: %v", err)
	}
	defer valkeyClient.Close()
	log.Println("Conectado a Valkey")

	channel, err := rabbitConn.Channel()
	if err != nil {
		log.Fatalf("Error abriendo canal: %v", err)
	}
	defer channel.Close()

	_, err = channel.QueueDeclare(queueName, true, false, false, false, nil)
	if err != nil {
		log.Fatalf("Error declarando queue: %v", err)
	}

	channel.Qos(1, 0, false)

	msgs, err := channel.Consume(queueName, "", false, false, false, false, nil)
	if err != nil {
		log.Fatalf("Error iniciando consumer: %v", err)
	}

	log.Printf("Consumer escuchando queue '%s'...", queueName)

	ctx := context.Background()
	for msg := range msgs {
		var report WarReportMessage
		if err := json.Unmarshal(msg.Body, &report); err != nil {
			log.Printf("Error parseando mensaje: %v", err)
			msg.Nack(false, false)
			continue
		}
		if err := processMessage(ctx, report, valkeyClient); err != nil {
			log.Printf("Error procesando mensaje: %v", err)
			msg.Nack(false, true)
			continue
		}
		msg.Ack(false)
	}
}
