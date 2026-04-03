// Package http contém os handlers HTTP da aplicação.
package http

import (
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"time"

	"github.com/rs/zerolog"

	"ProjetoGO_Claude/internal/domain/entity"
	"ProjetoGO_Claude/internal/domain/repository"
	"ProjetoGO_Claude/internal/usecase"
)

// ─────────────────────────────────────────────────────────────────────────────
// ScanHandler
// ─────────────────────────────────────────────────────────────────────────────

// ScanHandler gerencia as requisições HTTP do endpoint de scan.
type ScanHandler struct {
	service *usecase.ScanService
	log     zerolog.Logger
}

// NewScanHandler cria um ScanHandler injetando o ScanService.
func NewScanHandler(service *usecase.ScanService) *ScanHandler {
	return &ScanHandler{
		service: service,
		log: zerolog.New(os.Stdout).
			With().
			Timestamp().
			Str("component", "http_handler").
			Logger(),
	}
}

// RegisterRoutes registra todas as rotas do handler em um ServeMux.
func (h *ScanHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /v1/scan", h.HandleScan)
}

// HandleScan processa POST /v1/scan
func (h *ScanHandler) HandleScan(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	var req entity.ScanRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.log.Error().
			Err(err).
			Str("method", r.Method).
			Str("path", r.URL.Path).
			Int("status", http.StatusBadRequest).
			Msg("falha ao decodificar body da requisição")
		writeJSON(w, http.StatusBadRequest, errorResponse("invalid JSON body: "+err.Error()))
		return
	}
	defer r.Body.Close()

	// Logger enriquecido com dados do pacote — presente em todos os logs desta requisição
	reqLog := h.log.With().
		Str("shipment_id", req.ShipmentID).
		Str("vehicle_id", req.VehicleID).
		Str("method", r.Method).
		Str("path", r.URL.Path).
		Logger()

	reqLog.Info().Msg("requisição de scan recebida")

	result, err := h.service.ProcessScan(req)
	if err != nil {
		switch {
		case errors.Is(err, usecase.ErrInvalidRequest):
			reqLog.Warn().
				Err(err).
				Int("status", http.StatusBadRequest).
				Dur("latency_ms", time.Since(start)).
				Msg("requisição rejeitada: campos obrigatórios ausentes")
			writeJSON(w, http.StatusBadRequest, errorResponse(err.Error()))

		case errors.Is(err, repository.ErrShipmentNotFound):
			reqLog.Warn().
				Str("shipment_id", req.ShipmentID).
				Int("status", http.StatusNotFound).
				Dur("latency_ms", time.Since(start)).
				Msg("pacote não encontrado")
			writeJSON(w, http.StatusNotFound, errorResponse("shipment not found: "+req.ShipmentID))

		default:
			reqLog.Error().
				Err(err).
				Int("status", http.StatusInternalServerError).
				Dur("latency_ms", time.Since(start)).
				Msg("erro interno ao processar scan")
			writeJSON(w, http.StatusInternalServerError, errorResponse("internal server error"))
		}
		return
	}

	// Rota divergente → 422
	if !result.Valid {
		reqLog.Warn().
			Str("shipment_id", req.ShipmentID).
			Int("status", http.StatusUnprocessableEntity).
			Dur("latency_ms", time.Since(start)).
			Msg("scan inválido: divergência de rota")
		writeJSON(w, http.StatusUnprocessableEntity, result)
		return
	}

	reqLog.Info().
		Str("shipment_id", req.ShipmentID).
		Int("status", http.StatusOK).
		Dur("latency_ms", time.Since(start)).
		Msg("scan concluído com sucesso")

	writeJSON(w, http.StatusOK, result)
}

// ─────────────────────────────────────────────────────────────────────────────
// Helpers
// ─────────────────────────────────────────────────────────────────────────────

type errorBody struct {
	Error string `json:"error"`
}

func errorResponse(msg string) errorBody {
	return errorBody{Error: msg}
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(body)
}