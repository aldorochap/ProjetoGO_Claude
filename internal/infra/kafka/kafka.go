// Package kafka contém a implementação do producer de eventos.
// É um detalhe de infraestrutura — o usecase nunca importa este pacote diretamente.
package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/segmentio/kafka-go"
)

// ─────────────────────────────────────────────────────────────────────────────
// Constantes
// ─────────────────────────────────────────────────────────────────────────────

const (
	TopicDeliveryEvents = "delivery-events"
	brokerAddress       = "localhost:9092"
)

// ─────────────────────────────────────────────────────────────────────────────
// Evento
// ─────────────────────────────────────────────────────────────────────────────

// DeliveryEvent representa a estrutura do evento publicado no Kafka.
type DeliveryEvent struct {
	EventType  string    `json:"event_type"`
	ShipmentID string    `json:"shipment_id"`
	VehicleID  string    `json:"vehicle_id"`
	RouteID    string    `json:"route_id"`
	ScannedBy  string    `json:"scanned_by"`
	OccurredAt time.Time `json:"occurred_at"`
}

// ─────────────────────────────────────────────────────────────────────────────
// Interface
// ─────────────────────────────────────────────────────────────────────────────

// EventProducer é a interface que o usecase conhece.
// Em testes, basta passar um mock que implemente essa interface.
type EventProducer interface {
	PublishOutForDelivery(ctx context.Context, event DeliveryEvent) error
	Close() error
}

// ─────────────────────────────────────────────────────────────────────────────
// KafkaProducer
// ─────────────────────────────────────────────────────────────────────────────

// KafkaProducer implementa EventProducer usando kafka-go.
type KafkaProducer struct {
	writer *kafka.Writer
	log    zerolog.Logger
}

// NewKafkaProducer cria um producer e valida a conectividade com o broker.
func NewKafkaProducer() *KafkaProducer {
	log := zerolog.New(os.Stdout).
		With().
		Timestamp().
		Str("component", "kafka_producer").
		Str("broker", brokerAddress).
		Str("topic", TopicDeliveryEvents).
		Logger()

	// Testa conectividade antes de aceitar tráfego.
	// kafka-go é lazy por padrão (só conecta no primeiro write),
	// então usamos um Dialer explícito para falhar rápido na inicialização.
	dialer := &kafka.Dialer{Timeout: 5 * time.Second}
	conn, err := dialer.DialContext(context.Background(), "tcp", brokerAddress)
	if err != nil {
		log.Error().
			Err(err).
			Msg("falha ao conectar no broker Kafka — verifique se o cluster está rodando")
	} else {
		conn.Close()
		log.Info().Msg("conexão com broker verificada com sucesso")
	}

	writer := &kafka.Writer{
		Addr:                   kafka.TCP(brokerAddress),
		Topic:                  TopicDeliveryEvents,
		Balancer:               &kafka.LeastBytes{},
		RequiredAcks:           kafka.RequireOne,
		WriteTimeout:           10 * time.Second,
		ReadTimeout:            10 * time.Second,
		AllowAutoTopicCreation: true,
	}

	return &KafkaProducer{writer: writer, log: log}
}

// PublishOutForDelivery serializa o evento e publica no tópico delivery-events.
// A chave da mensagem é o ShipmentID — garante ordenação por pacote na partição.
func (p *KafkaProducer) PublishOutForDelivery(ctx context.Context, event DeliveryEvent) error {
	payload, err := json.Marshal(event)
	if err != nil {
		p.log.Error().
			Err(err).
			Str("shipment_id", event.ShipmentID).
			Msg("falha ao serializar evento")
		return fmt.Errorf("kafka: falha ao serializar evento: %w", err)
	}

	msg := kafka.Message{
		Key:   []byte(event.ShipmentID),
		Value: payload,
		Headers: []kafka.Header{
			{Key: "event_type", Value: []byte(event.EventType)},
			{Key: "source", Value: []byte("scan-service")},
		},
	}

	if err := p.writer.WriteMessages(ctx, msg); err != nil {
		p.log.Error().
			Err(err).
			Str("shipment_id", event.ShipmentID).
			Str("vehicle_id", event.VehicleID).
			Str("event_type", event.EventType).
			Msg("falha ao publicar evento no broker")
		return fmt.Errorf("kafka: falha ao publicar evento para shipment %s: %w", event.ShipmentID, err)
	}

	p.log.Info().
		Str("shipment_id", event.ShipmentID).
		Str("vehicle_id", event.VehicleID).
		Str("route_id", event.RouteID).
		Str("event_type", event.EventType).
		Msg("evento publicado com sucesso")

	return nil
}

// Close encerra a conexão com o broker de forma limpa.
func (p *KafkaProducer) Close() error {
	p.log.Info().Msg("encerrando producer")
	return p.writer.Close()
}