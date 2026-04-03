package main

import (
	"net/http"
	"os"

	"github.com/rs/zerolog"

	infra_http "ProjetoGO_Claude/internal/infra/http"
	"ProjetoGO_Claude/internal/infra/kafka"
	"ProjetoGO_Claude/internal/infra/repository"
	"ProjetoGO_Claude/internal/usecase"
)

func main() {
	// Logger raiz da aplicação — todos os outros herdam este formato
	log := zerolog.New(os.Stdout).
		With().
		Timestamp().
		Str("service", "scan-api").
		Logger()

	log.Info().Msg("iniciando serviço de scan de carregamento")

	// 1. Repositório em memória
	repo := repository.NewMemoryShipmentRepository()

	// 2. Producer Kafka
	producer := kafka.NewKafkaProducer()
	defer func() {
		if err := producer.Close(); err != nil {
			log.Error().Err(err).Msg("erro ao encerrar producer Kafka")
		}
	}()

	// 3. Caso de uso
	scanService := usecase.NewScanService(repo, producer)

	// 4. Handler HTTP
	scanHandler := infra_http.NewScanHandler(scanService)

	// 5. Roteador
	mux := http.NewServeMux()
	scanHandler.RegisterRoutes(mux)

	// 6. Servidor
	addr := ":8081"
	log.Info().
		Str("addr", addr).
		Str("endpoint", "POST /v1/scan").
		Msg("servidor HTTP pronto")

	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatal().Err(err).Msg("servidor encerrado inesperadamente")
	}
}