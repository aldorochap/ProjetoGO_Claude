// Package usecase contém as regras de negócio da aplicação.
package usecase

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/rs/zerolog"

	"ProjetoGO_Claude/internal/domain/entity"
	"ProjetoGO_Claude/internal/domain/repository"
	"ProjetoGO_Claude/internal/infra/kafka"
)

// ─────────────────────────────────────────────────────────────────────────────
// Erros de negócio
// ─────────────────────────────────────────────────────────────────────────────

var ErrRouteMismatch = errors.New("route mismatch: shipment does not belong to this vehicle's route")
var ErrInvalidRequest = errors.New("invalid scan request: shipment_id, vehicle_id and route_id are required")

// ─────────────────────────────────────────────────────────────────────────────
// ScanService
// ─────────────────────────────────────────────────────────────────────────────

// ScanService contém a lógica de negócio para o processo de scan de carregamento.
type ScanService struct {
	repo     repository.ShipmentRepository
	producer kafka.EventProducer
	log      zerolog.Logger
}

// NewScanService cria um ScanService injetando repositório e producer via interfaces.
func NewScanService(repo repository.ShipmentRepository, producer kafka.EventProducer) *ScanService {
	return &ScanService{
		repo:     repo,
		producer: producer,
		log: zerolog.New(os.Stdout).
			With().
			Timestamp().
			Str("component", "scan_service").
			Logger(),
	}
}

// ProcessScan executa a validação do scan de um pacote em um veículo.
func (s *ScanService) ProcessScan(req entity.ScanRequest) (*entity.ScanResult, error) {
	log := s.log.With().
		Str("shipment_id", req.ShipmentID).
		Str("vehicle_id", req.VehicleID).
		Str("route_id", req.RouteID).
		Str("scanned_by", req.ScannedBy).
		Logger()

	// 1. Validar campos obrigatórios
	if req.ShipmentID == "" || req.VehicleID == "" || req.RouteID == "" {
		log.Warn().Msg("requisição inválida: campos obrigatórios ausentes")
		return nil, ErrInvalidRequest
	}

	// 2. Buscar pacote
	shipment, err := s.repo.FindByID(req.ShipmentID)
	if err != nil {
		log.Error().Err(err).Msg("pacote não encontrado no repositório")
		return nil, fmt.Errorf("shipment lookup failed: %w", err)
	}

	result := &entity.ScanResult{
		ShipmentID: req.ShipmentID,
		VehicleID:  req.VehicleID,
		ScannedAt:  time.Now().UTC(),
	}

	// 3. Validar rota
	if shipment.RouteID != req.RouteID {
		result.Valid = false
		result.Message = fmt.Sprintf(
			"ALERTA: pacote %s pertence à rota %s, mas o veículo está na rota %s",
			req.ShipmentID, shipment.RouteID, req.RouteID,
		)
		log.Warn().
			Str("shipment_route", shipment.RouteID).
			Str("vehicle_route", req.RouteID).
			Msg("divergência de rota detectada — evento não será publicado")
		return result, nil
	}

	// 4. Atualizar status
	if err := s.repo.UpdateStatus(req.ShipmentID, "in_transit"); err != nil {
		log.Error().Err(err).Msg("falha ao atualizar status do pacote")
		return nil, fmt.Errorf("failed to update shipment status: %w", err)
	}

	result.Valid = true
	result.Message = fmt.Sprintf(
		"OK: pacote %s carregado com sucesso no veículo %s (rota %s)",
		req.ShipmentID, req.VehicleID, req.RouteID,
	)

	log.Info().Msg("scan validado — disparando evento OUT_FOR_DELIVERY")

	// 5. Publicar evento via goroutine — não bloqueia a resposta HTTP
	event := kafka.DeliveryEvent{
		EventType:  "OUT_FOR_DELIVERY",
		ShipmentID: req.ShipmentID,
		VehicleID:  req.VehicleID,
		RouteID:    req.RouteID,
		ScannedBy:  req.ScannedBy,
		OccurredAt: result.ScannedAt,
	}

	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()

		if err := s.producer.PublishOutForDelivery(ctx, event); err != nil {
			// Erro já logado dentro do producer com detalhes do broker
			log.Error().
				Err(err).
				Str("event_type", event.EventType).
				Msg("goroutine de publicação encerrada com erro")
		}
	}()

	return result, nil
}