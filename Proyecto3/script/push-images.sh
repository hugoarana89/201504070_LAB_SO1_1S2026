#!/bin/bash

# =============================================================
# Script para construir y subir imágenes a Zot Registry (Bash)
# =============================================================

# Configuración del dominio
ZOT="curtly-duress-exciting.ngrok-free.dev" # Reemplaza con tu dominio ngrok

# Colores para la terminal
CYAN='\033[0;36m'
YELLOW='\033[1;33m'
GREEN='\033[0;32m'
RED='\033[0;31m'
NC='\033[0m' # Sin color

echo -e "${CYAN}====================================================${NC}"
echo -e "${CYAN} Subiendo imágenes al registry Zot: https://$ZOT${NC}"
echo -e "${CYAN}====================================================${NC}"

# Función para verificar errores
check_error() {
    if [ $? -ne 0 ]; then
        echo -e "${RED}ERROR: $1${NC}"
        exit 1
    fi
}

# ------------------------------------------------------------------
# 1. rust-api
# ------------------------------------------------------------------
echo -e "\n${YELLOW}[1/6] Construyendo rust-api...${NC}"
docker build -t "$ZOT/rust-api:v1" ./rust-api
check_error "build rust-api"

docker push "$ZOT/rust-api:v1"
check_error "push rust-api"
echo -e "${GREEN}      rust-api OK${NC}"

# ------------------------------------------------------------------
# 2. go-service
# ------------------------------------------------------------------
echo -e "\n${YELLOW}[2/6] Construyendo go-service...${NC}"
docker build -t "$ZOT/go-service:v1" ./go-service
check_error "build go-service"

docker push "$ZOT/go-service:v1"
check_error "push go-service"
echo -e "${GREEN}      go-service OK${NC}"

# ------------------------------------------------------------------
# 3. grpc-server
# ------------------------------------------------------------------
echo -e "\n${YELLOW}[3/6] Construyendo grpc-server...${NC}"
docker build -t "$ZOT/grpc-server:v1" ./grpc-server
check_error "build grpc-server"

docker push "$ZOT/grpc-server:v1"
check_error "push grpc-server"
echo -e "${GREEN}      grpc-server OK${NC}"

# ------------------------------------------------------------------
# 4. consumer
# ------------------------------------------------------------------
echo -e "\n${YELLOW}[4/6] Construyendo consumer...${NC}"
docker build -t "$ZOT/consumer:v1" ./consumer
check_error "build consumer"

docker push "$ZOT/consumer:v1"
check_error "push consumer"
echo -e "${GREEN}      consumer OK${NC}"

# ------------------------------------------------------------------
# 5. rabbitmq
# ------------------------------------------------------------------
echo -e "\n${YELLOW}[5/6] Construyendo rabbitmq...${NC}"
docker build -t "$ZOT/rabbitmq:3-management" ./rabbitmq
check_error "build rabbitmq"

docker push "$ZOT/rabbitmq:3-management"
check_error "push rabbitmq"
echo -e "${GREEN}      rabbitmq OK${NC}"

# ------------------------------------------------------------------
# 6. valkey
# ------------------------------------------------------------------
echo -e "\n${YELLOW}[6/6] Construyendo valkey...${NC}"
docker build -t "$ZOT/valkey:7.2" ./valkey-image
check_error "build valkey"

docker push "$ZOT/valkey:7.2"
check_error "push valkey"
echo -e "${GREEN}      valkey OK${NC}"

# ------------------------------------------------------------------
# Verificar catálogo
# ------------------------------------------------------------------
echo -e "\n${YELLOW}[Verificando] Catálogo en Zot...${NC}"
curl -k "https://$ZOT/v2/_catalog"

echo -e "\n${CYAN}====================================================${NC}"
echo -e "${GREEN} TODO SUBIDO CORRECTAMENTE${NC}"
echo -e "${CYAN}====================================================${NC}"