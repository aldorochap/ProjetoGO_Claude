package repository

import "ProjetoGO_Claude/internal/domain/entity"

type ShipmentRepository interface {
    FindByID(id string) (*entity.Shipment, error)
    Save(shipment *entity.Shipment) error
    UpdateStatus(id, status string) error
}