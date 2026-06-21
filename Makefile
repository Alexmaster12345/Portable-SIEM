.PHONY: all backend frontend agent up down logs clean

# Start full stack
up:
	docker compose up -d

# Stop full stack
down:
	docker compose down

# View logs
logs:
	docker compose logs -f

# Build and run backend only (dev mode)
backend:
	cd backend && go run ./cmd/server

# Run frontend dev server
frontend:
	cd frontend && npm run dev

# Build Linux agent
agent-linux:
	cd agents/linux-agent && go build -o ../../dist/siem-agent-linux .

# Build all agents
agents:
	$(MAKE) agent-linux

# Run all tests
test:
	cd backend && go test ./...

# Download Go deps
deps:
	cd backend && go mod download
	cd agents/linux-agent && go mod download
	cd frontend && npm install

# Build production binaries
build:
	cd backend && CGO_ENABLED=0 go build -ldflags="-s -w" -o ../dist/siem-server ./cmd/server
	$(MAKE) agent-linux
	cd frontend && npm run build

# Initialize database (requires running postgres)
db-init:
	PGPASSWORD=siem psql -h localhost -U siem -d siem -f backend/internal/storage/migrations/001_init.sql

# Load sample IOCs
load-iocs:
	curl -s -X POST http://localhost:8080/api/v1/events \
		-H "Content-Type: application/json" \
		-d '{"host":"test","source":"auth","event_type":"login_failed","severity":"medium","message":"Failed password for admin from 10.10.1.20 port 22 ssh2"}'

# Send a test alert
test-event:
	@for i in $$(seq 1 55); do \
		curl -s -X POST http://localhost:8080/api/v1/events \
			-H "Content-Type: application/json" \
			-d "{\"host\":\"server01\",\"source\":\"auth\",\"event_type\":\"login_failed\",\"severity\":\"medium\",\"message\":\"Failed password for admin from 1.2.3.4 port 22\",\"fields\":{\"src_ip\":\"1.2.3.4\",\"username\":\"admin\"}}" > /dev/null; \
	done
	@echo "Sent 55 login_failed events — brute force alert should fire"

# Package for USB deployment
usb-package:
	$(MAKE) build
	mkdir -p usb-package
	cp dist/siem-server usb-package/
	cp dist/siem-agent-linux usb-package/
	cp config.yaml usb-package/
	cp -r rules usb-package/
	cp -r intelligence usb-package/
	cp deploy/usb/setup.sh usb-package/
	chmod +x usb-package/setup.sh
	tar -czf portable-siem-$(shell date +%Y%m%d).tar.gz usb-package/
	@echo "USB package created: portable-siem-$(shell date +%Y%m%d).tar.gz"

clean:
	rm -rf dist/ usb-package/ *.tar.gz
	cd frontend && rm -rf dist/
