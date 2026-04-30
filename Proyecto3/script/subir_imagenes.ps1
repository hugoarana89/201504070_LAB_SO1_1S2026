# =============================================================
# Script para construir y subir imágenes a Zot Registry
# =============================================================

#$ZOT = "<IP_PUBLICA_DE_VM>:5000" # Reemplaza <IP_PUBLICA_DE_VM> con la IP pública de tu VM ejp: 34.66.217.216
$ZOT = "curtly-duress-exciting.ngrok-free.dev" # Reemplaza con tu dominio ngrok


Write-Host "====================================================" -ForegroundColor Cyan
Write-Host " Subiendo imágenes al registry Zot: https://$ZOT" -ForegroundColor Cyan
Write-Host "====================================================" -ForegroundColor Cyan

function Check-Error {
    param($msg)
    if ($LASTEXITCODE -ne 0) {
        Write-Host "ERROR: $msg" -ForegroundColor Red
        exit 1
    }
}

# ------------------------------------------------------------------
# 1. rust-api
# ------------------------------------------------------------------
Write-Host "`n[1/6] Construyendo rust-api..." -ForegroundColor Yellow
docker build -t "$ZOT/rust-api:v1" ./rust-api
Check-Error "build rust-api"

docker push "$ZOT/rust-api:v1"
Check-Error "push rust-api"
Write-Host "      rust-api OK" -ForegroundColor Green

# ------------------------------------------------------------------
# 2. go-service
# ------------------------------------------------------------------
Write-Host "`n[2/6] Construyendo go-service..." -ForegroundColor Yellow
docker build -t "$ZOT/go-service:v1" ./go-service
Check-Error "build go-service"

docker push "$ZOT/go-service:v1"
Check-Error "push go-service"
Write-Host "      go-service OK" -ForegroundColor Green

# ------------------------------------------------------------------
# 3. grpc-server
# ------------------------------------------------------------------
Write-Host "`n[3/6] Construyendo grpc-server..." -ForegroundColor Yellow
docker build -t "$ZOT/grpc-server:v1" ./grpc-server
Check-Error "build grpc-server"

docker push "$ZOT/grpc-server:v1"
Check-Error "push grpc-server"
Write-Host "      grpc-server OK" -ForegroundColor Green

# ------------------------------------------------------------------
# 4. consumer
# ------------------------------------------------------------------
Write-Host "`n[4/6] Construyendo consumer..." -ForegroundColor Yellow
docker build -t "$ZOT/consumer:v1" ./consumer
Check-Error "build consumer"

docker push "$ZOT/consumer:v1"
Check-Error "push consumer"
Write-Host "      consumer OK" -ForegroundColor Green

# ------------------------------------------------------------------
# 5. rabbitmq (BUILD LOCAL - igual que valkey)
# ------------------------------------------------------------------
Write-Host "`n[5/6] Construyendo rabbitmq desde Dockerfile..." -ForegroundColor Yellow
docker build -t "$ZOT/rabbitmq:3-management" ./rabbitmq
Check-Error "build rabbitmq"

docker push "$ZOT/rabbitmq:3-management"
Check-Error "push rabbitmq"
Write-Host "      rabbitmq OK" -ForegroundColor Green

# ------------------------------------------------------------------
# 6. valkey
# ------------------------------------------------------------------
Write-Host "`n[6/6] Construyendo valkey desde Dockerfile..." -ForegroundColor Yellow
docker build -t "$ZOT/valkey:7.2" ./valkey-image
Check-Error "build valkey"

docker push "$ZOT/valkey:7.2"
Check-Error "push valkey"
Write-Host "      valkey OK" -ForegroundColor Green

# ------------------------------------------------------------------
# Verificar catálogo
# ------------------------------------------------------------------
Write-Host "`n[Verificando] Catálogo en Zot..." -ForegroundColor Yellow
curl.exe -k "https://$ZOT/v2/_catalog"

Write-Host "`n====================================================" -ForegroundColor Cyan
Write-Host " TODO SUBIDO CORRECTAMENTE" -ForegroundColor Green
Write-Host "====================================================" -ForegroundColor Cyan