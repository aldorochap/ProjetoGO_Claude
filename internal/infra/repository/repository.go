// Package repository define a interface de acesso a dados para Shipments.
// A interface fica no domínio; a implementação concreta fica na infra.
package repository

import (
	"errors"
	"sync"

	"ProjetoGO_Claude/internal/domain/entity"
)

// ─────────────────────────────────────────────────────────────────────────────
// Implementação em memória — simula o banco de dados
// Arquivo: internal/infra/repository/shipment_memory_repository.go
// ─────────────────────────────────────────────────────────────────────────────

// ErrShipmentNotFound é retornado quando o pacote não existe no repositório.
var ErrShipmentNotFound = errors.New("shipment not found")

// MemoryShipmentRepository é uma implementação in-memory de ShipmentRepository.
// Útil para desenvolvimento e testes sem dependência de banco de dados.
type MemoryShipmentRepository struct {
	mu       sync.RWMutex
	shipments map[string]*entity.Shipment
}

// NewMemoryShipmentRepository cria o repositório já populado com dados de exemplo.
func NewMemoryShipmentRepository() *MemoryShipmentRepository {
	repo := &MemoryShipmentRepository{
		shipments: make(map[string]*entity.Shipment),
	}
	repo.seed()
	return repo
}

// seed popula o repositório com pacotes fictícios para facilitar testes manuais.
func (r *MemoryShipmentRepository) seed() {
	samples := []entity.Shipment{
		{ID: "SHP-001", Description: "Eletrônicos frágeis",  RouteID: "ROUTE-SP-01", Status: "pending"},
		{ID: "SHP-002", Description: "Documentos fiscais",   RouteID: "ROUTE-SP-02", Status: "pending"},
		{ID: "SHP-003", Description: "Peças automotivas",    RouteID: "ROUTE-RJ-01", Status: "pending"},
		{ID: "SHP-004", Description: "Medicamentos urgentes",RouteID: "ROUTE-SP-01", Status: "pending"},
	}
	for i := range samples {
		r.shipments[samples[i].ID] = &samples[i]
	}
}

// FindByID busca um pacote pelo seu ID. Retorna ErrShipmentNotFound se não existir.
func (r *MemoryShipmentRepository) FindByID(id string) (*entity.Shipment, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	s, ok := r.shipments[id]
	if !ok {
		return nil, ErrShipmentNotFound
	}
	return s, nil
}

// Save persiste um novo pacote. Retorna erro se o ID já existir.
func (r *MemoryShipmentRepository) Save(shipment *entity.Shipment) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.shipments[shipment.ID]; exists {
		return errors.New("shipment already exists: " + shipment.ID)
	}
	r.shipments[shipment.ID] = shipment
	return nil
}

// UpdateStatus atualiza o status de um pacote existente.
func (r *MemoryShipmentRepository) UpdateStatus(id, status string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	s, ok := r.shipments[id]
	if !ok {
		return ErrShipmentNotFound
	}
	s.Status = status
	return nil
}