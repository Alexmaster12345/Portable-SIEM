#!/usr/bin/env bash
# Portable SIEM - USB Setup Script
# Supports: Ubuntu, Debian, Kali, Mint, Rocky, RHEL, CentOS, AlmaLinux,
#           Fedora, Arch, Manjaro, openSUSE, Alpine, Raspberry Pi OS

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
SIEM_DIR="$(cd "${SCRIPT_DIR}/../.." && pwd)"
DATA_DIR="${HOME}/.portable-siem"

echo "============================================"
echo "  Portable SIEM - Setup"
echo "  $(uname -srm)"
echo "============================================"

# ---- Detect Linux distro ----
detect_distro() {
    if [ -f /etc/os-release ]; then
        . /etc/os-release
        echo "${ID:-unknown}"
    elif [ -f /etc/redhat-release ]; then
        echo "rhel"
    elif [ -f /etc/debian_version ]; then
        echo "debian"
    elif [ -f /etc/alpine-release ]; then
        echo "alpine"
    else
        echo "unknown"
    fi
}

DISTRO="$(detect_distro)"
echo "Detected OS: ${DISTRO}"
echo ""

# ---- Auto-install Docker if missing ----
install_docker() {
    echo "Docker not found — installing for ${DISTRO}..."

    case "${DISTRO}" in
        ubuntu|debian|linuxmint|pop|elementary|kali|raspbian)
            sudo apt-get update -qq
            sudo apt-get install -y -qq docker.io docker-compose-plugin curl
            sudo systemctl enable --now docker
            sudo usermod -aG docker "${USER}"
            ;;
        centos|rhel|rocky|almalinux)
            sudo dnf install -y -q docker docker-compose-plugin curl
            sudo systemctl enable --now docker
            sudo usermod -aG docker "${USER}"
            ;;
        fedora)
            sudo dnf install -y -q docker docker-compose-plugin curl
            sudo systemctl enable --now docker
            sudo usermod -aG docker "${USER}"
            ;;
        arch|manjaro|endeavouros|garuda)
            sudo pacman -Sy --noconfirm docker docker-compose curl
            sudo systemctl enable --now docker
            sudo usermod -aG docker "${USER}"
            ;;
        opensuse*|sles|suse)
            sudo zypper install -y docker docker-compose curl
            sudo systemctl enable --now docker
            sudo usermod -aG docker "${USER}"
            ;;
        alpine)
            sudo apk add --no-cache docker docker-compose curl
            sudo rc-update add docker boot
            sudo service docker start
            sudo addgroup "${USER}" docker
            ;;
        *)
            echo ""
            echo "ERROR: Unknown distro '${DISTRO}'."
            echo "Please install Docker manually:"
            echo "  https://docs.docker.com/engine/install/"
            exit 1
            ;;
    esac

    echo "Docker installed successfully."
    echo ""
    echo "IMPORTANT: You have been added to the 'docker' group."
    echo "           Run: newgrp docker   (to apply without logout)"
    echo ""
}

# ---- Check Docker ----
if ! command -v docker &>/dev/null; then
    install_docker
fi

# ---- Check Docker daemon is running ----
if ! docker info &>/dev/null 2>&1; then
    echo "Docker daemon is not running. Starting it..."
    if command -v systemctl &>/dev/null; then
        sudo systemctl start docker
    elif command -v service &>/dev/null; then
        sudo service docker start
    elif [ "${DISTRO}" = "alpine" ]; then
        sudo service docker start
    fi
    sleep 3
fi

# ---- Detect Compose command ----
if docker compose version &>/dev/null 2>&1; then
    COMPOSE="docker compose"
elif command -v docker-compose &>/dev/null; then
    COMPOSE="docker-compose"
else
    echo "ERROR: Docker Compose not found."
    echo "  Install: sudo apt install docker-compose-plugin  (Debian/Ubuntu)"
    echo "  Install: sudo dnf install docker-compose-plugin  (RHEL/Rocky/Fedora)"
    exit 1
fi

echo "Using: $COMPOSE"

# ---- Setup data directory & config ----
mkdir -p "${DATA_DIR}"/{postgres,redis,logs}

if [[ ! -f "${DATA_DIR}/config.yaml" ]]; then
    cp "${SIEM_DIR}/config.yaml" "${DATA_DIR}/config.yaml"
    echo "Config created at: ${DATA_DIR}/config.yaml"
fi

# ---- Start services ----
echo ""
echo "Starting services..."
cd "${SIEM_DIR}"

$COMPOSE up -d postgres redis nats

echo "Waiting for database to be ready..."
for i in $(seq 1 15); do
    if $COMPOSE exec -T postgres pg_isready -U siem &>/dev/null 2>&1; then
        echo "  Database ready."
        break
    fi
    echo "  attempt $i/15..."
    sleep 3
done

# ---- Initialize DB ----
echo "Initializing database schema..."
$COMPOSE exec -T postgres psql -U siem -d siem \
    -f /docker-entrypoint-initdb.d/001_init.sql 2>/dev/null || true

# ---- Start SIEM ----
$COMPOSE up -d siem-server siem-frontend

echo "Waiting for SIEM server to be ready..."
for i in $(seq 1 20); do
    if curl -sf http://localhost:8888/health > /dev/null 2>&1; then
        echo "  Server is up!"
        break
    fi
    echo "  attempt $i/20..."
    sleep 3
done

# ---- Final check ----
if ! curl -sf http://localhost:8888/health > /dev/null 2>&1; then
    echo ""
    echo "WARNING: SIEM server did not respond. Showing logs:"
    $COMPOSE logs --tail=30 siem-server
    echo ""
    echo "Run '$COMPOSE logs siem-server' for full logs."
    exit 1
fi

HOST_IP=$(hostname -I 2>/dev/null | awk '{print $1}' || echo "localhost")

echo ""
echo "============================================"
echo "  Portable SIEM is running!"
echo ""
echo "  Dashboard:  http://localhost:3000"
echo "  API:        http://localhost:8888/api/v1"
echo "  Health:     http://localhost:8888/health"
echo ""
echo "  From network: http://${HOST_IP}:3000"
echo ""
echo "  Syslog UDP:   0.0.0.0:514"
echo "  Agent port:   0.0.0.0:9000"
echo ""
echo "  Seed test data:"
echo "    ./deploy/usb/seed-test-data.sh"
echo ""
echo "  Deploy agent on remote Linux hosts:"
echo "    ./siem-agent-linux -server http://${HOST_IP}:8888"
echo ""
echo "  Stop: cd \"${SIEM_DIR}\" && $COMPOSE down"
echo "============================================"
