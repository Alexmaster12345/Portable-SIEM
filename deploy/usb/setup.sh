#!/usr/bin/env bash
# Portable SIEM - USB Setup Script
# Run this on the target machine to start the SIEM

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
SIEM_DIR="$(cd "${SCRIPT_DIR}/../.." && pwd)"
DATA_DIR="${HOME}/.portable-siem"

echo "============================================"
echo "  Portable SIEM - Setup"
echo "============================================"

# Check dependencies
check_dep() {
    if ! command -v "$1" &> /dev/null; then
        echo "ERROR: $1 is required but not installed."
        exit 1
    fi
}

check_dep docker

# Support both docker compose v2 (plugin) and docker-compose v1
if docker compose version &>/dev/null 2>&1; then
    COMPOSE="docker compose"
elif command -v docker-compose &>/dev/null; then
    COMPOSE="docker-compose"
else
    echo "ERROR: Docker Compose is required but not installed."
    echo "  Install it with:"
    echo "    sudo dnf install docker-compose-plugin   # Rocky/RHEL/Fedora"
    echo "    sudo apt install docker-compose-plugin   # Debian/Ubuntu"
    exit 1
fi

# Create data directory
mkdir -p "${DATA_DIR}"/{postgres,redis,logs}

# Copy config if not exists
if [[ ! -f "${DATA_DIR}/config.yaml" ]]; then
    cp "${SIEM_DIR}/config.yaml" "${DATA_DIR}/config.yaml"
    echo "Config created at: ${DATA_DIR}/config.yaml"
fi

# Start services
echo ""
echo "Starting services..."
cd "${SIEM_DIR}"

$COMPOSE up -d postgres redis nats

echo "Waiting for database..."
sleep 5

# Initialize DB
echo "Initializing database..."
$COMPOSE exec -T postgres psql -U siem -d siem \
    -f /docker-entrypoint-initdb.d/001_init.sql 2>/dev/null || true

# Start SIEM server
$COMPOSE up -d siem-server siem-frontend

echo ""
echo "============================================"
echo "  Portable SIEM is running!"
echo ""
echo "  Dashboard:  http://localhost:3000"
echo "  API:        http://localhost:8888/api/v1"
echo "  Health:     http://localhost:8888/health"
echo ""
echo "  Syslog UDP: 0.0.0.0:514"
echo "  Agent port: 0.0.0.0:9000"
echo ""
echo "  To deploy Linux agents on target machines:"
echo "  ./siem-agent-linux -server http://$(hostname -I | awk '{print $1}'):8888"
echo ""
echo "  To stop: cd "/path/to/Portable SIEM" && docker compose down"
echo "============================================"
