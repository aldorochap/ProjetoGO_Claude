package entity

import "time"

// Shipment representa um pacote/carga no sistema de logística.
type Shipment struct {
	ID          string    `json:"id"`
	Description string    `json:"description"`
	RouteID     string    `json:"route_id"`   // ID da rota à qual o pacote pertence
	Status      string    `json:"status"`      // pending | in_transit | delivered
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// ScanRequest representa o payload enviado pelo operador ao escanear
// um pacote dentro de um veículo.
type ScanRequest struct {
	ShipmentID string `json:"shipment_id"` // ID do pacote escaneado
	VehicleID  string `json:"vehicle_id"`  // ID do veículo onde o pacote foi colocado
	RouteID    string `json:"route_id"`    // ID da rota programada para o veículo
	ScannedBy  string `json:"scanned_by"`  // operador responsável pelo scan
}

// ScanResult representa o resultado da operação de scan.
type ScanResult struct {
	ShipmentID string `json:"shipment_id"`
	VehicleID  string `json:"vehicle_id"`
	Valid       bool   `json:"valid"`             // true se rota do pacote == rota do veículo
	Message    string `json:"message"`
	ScannedAt  time.Time `json:"scanned_at"`
}