# ProjetoGO_Claude

API de scan de carregamento em Go com Kafka.

## Stack
- Go 1.22+
- Apache Kafka (Confluent)
- Docker / Docker Compose
- zerolog

## Como rodar
```bash
# Subir o cluster Kafka
docker compose up -d

# Rodar a API
go run ./cmd/app

# Testar
.\test-api.ps1
```

## Arquitetura
Clean Architecture com separação de domínio, casos de uso e infraestrutura.