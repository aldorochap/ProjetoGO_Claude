# ─────────────────────────────────────────────────────────────────────────────
# test-api.ps1 — Testes da API de Scan de Carregamento
# Uso: .\test-api.ps1
# ─────────────────────────────────────────────────────────────────────────────

$baseUrl = "http://localhost:8081/v1/scan"
$headers = @{ "Content-Type" = "application/json" }

function Write-Section($title) {
    Write-Host ""
    Write-Host "─────────────────────────────────────────" -ForegroundColor DarkGray
    Write-Host " $title" -ForegroundColor Cyan
    Write-Host "─────────────────────────────────────────" -ForegroundColor DarkGray
}

function Invoke-ScanRequest($label, $body, $expectedStatus) {
    Write-Host ""
    Write-Host "► $label" -ForegroundColor Yellow
    Write-Host "  Payload: $body" -ForegroundColor DarkGray

    try {
        $response = Invoke-WebRequest `
            -Uri $baseUrl `
            -Method POST `
            -Headers $headers `
            -Body $body `
            -ErrorAction Stop

        $status = $response.StatusCode
        $json   = $response.Content | ConvertFrom-Json | ConvertTo-Json -Depth 5

        $color = if ($status -eq $expectedStatus) { "Green" } else { "Red" }
        Write-Host "  Status : $status (esperado: $expectedStatus)" -ForegroundColor $color
        Write-Host "  Body   :" -ForegroundColor DarkGray
        Write-Host $json -ForegroundColor White

    } catch {
        $status = $_.Exception.Response.StatusCode.value__
        $raw    = $_.ErrorDetails.Message

        $color = if ($status -eq $expectedStatus) { "Green" } else { "Red" }
        Write-Host "  Status : $status (esperado: $expectedStatus)" -ForegroundColor $color

        if ($raw) {
            $json = $raw | ConvertFrom-Json | ConvertTo-Json -Depth 5
            Write-Host "  Body   :" -ForegroundColor DarkGray
            Write-Host $json -ForegroundColor White
        }
    }
}

# ─────────────────────────────────────────────────────────────────────────────
# CENÁRIO 1 — Sucesso: rota correta → 200 OK + evento Kafka publicado
# ─────────────────────────────────────────────────────────────────────────────
Write-Section "CENÁRIO 1 · Scan válido (rota correta) → esperado 200 OK"

Invoke-ScanRequest `
    -label "SHP-001 na rota ROUTE-SP-01 (correta)" `
    -body  '{"shipment_id":"SHP-001","vehicle_id":"VHC-42","route_id":"ROUTE-SP-01","scanned_by":"operador.jose"}' `
    -expectedStatus 200

# ─────────────────────────────────────────────────────────────────────────────
# CENÁRIO 2 — Erro de negócio: rota errada → 422 Unprocessable
# ─────────────────────────────────────────────────────────────────────────────
Write-Section "CENÁRIO 2 · Scan inválido (rota errada) → esperado 422"

Invoke-ScanRequest `
    -label "SHP-002 na rota ROUTE-RJ-01 (errada, deveria ser ROUTE-SP-02)" `
    -body  '{"shipment_id":"SHP-002","vehicle_id":"VHC-42","route_id":"ROUTE-RJ-01","scanned_by":"operador.jose"}' `
    -expectedStatus 422

# ─────────────────────────────────────────────────────────────────────────────
# CENÁRIO 3 — Pacote inexistente → 404 Not Found
# ─────────────────────────────────────────────────────────────────────────────
Write-Section "CENÁRIO 3 · Pacote inexistente → esperado 404"

Invoke-ScanRequest `
    -label "SHP-999 não existe no sistema" `
    -body  '{"shipment_id":"SHP-999","vehicle_id":"VHC-42","route_id":"ROUTE-SP-01","scanned_by":"operador.jose"}' `
    -expectedStatus 404

# ─────────────────────────────────────────────────────────────────────────────
# CENÁRIO 4 — Payload inválido (campo ausente) → 400 Bad Request
# ─────────────────────────────────────────────────────────────────────────────
Write-Section "CENÁRIO 4 · Campo obrigatório ausente → esperado 400"

Invoke-ScanRequest `
    -label "route_id ausente no payload" `
    -body  '{"shipment_id":"SHP-003","vehicle_id":"VHC-42"}' `
    -expectedStatus 400

# ─────────────────────────────────────────────────────────────────────────────
Write-Host ""
Write-Host "─────────────────────────────────────────" -ForegroundColor DarkGray
Write-Host " Testes concluídos." -ForegroundColor Cyan
Write-Host "─────────────────────────────────────────" -ForegroundColor DarkGray
Write-Host ""