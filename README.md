# Portable SIEM

A portable Security Information and Event Management system that runs entirely from a USB drive.
Plug it into any environment for instant security monitoring and incident response.

---

## Mounting the USB Drive

### Linux

**Find your USB device:**
```bash
lsblk -o NAME,SIZE,FSTYPE,TRAN,VENDOR,MOUNTPOINT
# Look for tran=usb — typically /dev/sda or /dev/sdb
```

**Mount it:**
```bash
sudo mkdir -p /mnt/usb
sudo mount /dev/sda1 /mnt/usb
```

**Or without sudo (auto-mount via udisks):**
```bash
udisksctl mount -b /dev/sda1
# Mounts to /run/media/$USER/<label>
```

**Start the SIEM:**
```bash
cd /mnt/usb/"Portable SIEM"
./deploy/usb/setup.sh
```

**Unmount when done:**
```bash
sudo umount /mnt/usb
# or
udisksctl unmount -b /dev/sda1
```

---

### Windows

**Find the drive letter:**

The USB drive will appear automatically in File Explorer (e.g. `D:\`, `E:\`).

To confirm via PowerShell:
```powershell
Get-Disk | Where-Object BusType -eq USB
Get-Partition | Where-Object DiskNumber -eq 1 | Select DriveLetter, Size
```

**Start the SIEM (PowerShell as Administrator):**
```powershell
# Navigate to the drive (replace D: with your drive letter)
cd "D:\Portable SIEM"

# Run setup (requires Docker Desktop to be running)
docker compose up -d
```

**Or via WSL2 (Windows Subsystem for Linux):**
```bash
# In WSL2, Windows drives are at /mnt/<letter>
cd "/mnt/d/Portable SIEM"
./deploy/usb/setup.sh
```

**Access the drive in WSL2 if not auto-mounted:**
```bash
sudo mkdir -p /mnt/d
sudo mount -t drvfs D: /mnt/d
```

**Unmount when done (PowerShell):**
```powershell
# Right-click the USB in File Explorer → Eject
# Or via PowerShell:
$drive = Get-WmiObject Win32_Volume | Where-Object { $_.DriveType -eq 2 }
$drive.Dismount($false, $false)
```

---

### Requirements on Target Machine

| Requirement | Linux | Windows |
|-------------|-------|---------|
| Docker Engine | `sudo apt install docker.io` | [Docker Desktop](https://www.docker.com/products/docker-desktop/) |
| Docker Compose | included with Docker | included with Docker Desktop |
| WSL2 (optional) | — | Recommended for script support |

> **Note:** The USB drive is formatted as **exFAT**, which is readable on Linux, Windows, and macOS without extra drivers.

---

## Architecture

```
Portable SIEM
│
├── backend/                    Go server
│   ├── cmd/server/             Main entry point
│   └── internal/
│       ├── api/                REST API (Gin)
│       ├── collector/          Log collection engines
│       │   ├── linux/          journald, auth.log, syslog, secure
│       │   ├── windows/        Windows Event Log (stub)
│       │   └── network/        UDP syslog receiver (routers, firewalls)
│       ├── parser/             Enrichment pipeline (MITRE, IOC, fields)
│       ├── rules/              JSON-based detection rule engine
│       ├── correlation/        Threshold & impossible-travel detection
│       ├── alert/              Alert persistence & Slack/email notifications
│       ├── search/             Full-text + field search over PostgreSQL
│       ├── intelligence/       Threat feed downloader & IOC lookup
│       ├── incident/           Incident management with timeline
│       ├── storage/            PostgreSQL + Redis adapters
│       └── models/             Shared data types
│
├── frontend/                   React + TypeScript + Tailwind dashboard
│   └── src/
│       ├── pages/              Overview, Events, Alerts, Search, Incidents, Rules
│       ├── components/         Layout, Sidebar, SeverityBadge, StatCard
│       └── api/                Typed API client (axios)
│
├── agents/
│   └── linux-agent/            Lightweight Go agent for remote hosts
│
├── rules/                      JSON detection rules (loaded at startup)
│   ├── ssh_brute_force.json
│   ├── impossible_travel.json
│   ├── privilege_escalation.json
│   ├── port_scan.json
│   └── user_account_created.json
│
├── intelligence/feeds/         Threat intelligence IOC files
├── deploy/usb/setup.sh         One-command USB deployment script
└── docker-compose.yml          Full stack (Postgres, Redis, NATS, Server, Frontend)
```

## Quick Start

### USB / Direct Deployment

```bash
# 1. Copy this entire folder to your USB drive
# 2. On the target machine:
chmod +x deploy/usb/setup.sh
./deploy/usb/setup.sh

# Dashboard is now at http://localhost:3000
```

### Development

```bash
# Install dependencies
make deps

# Start infrastructure
docker compose up -d postgres redis nats

# Init database
make db-init

# Run backend
make backend

# Run frontend (separate terminal)
make frontend

# Dashboard: http://localhost:3000
# API:       http://localhost:8080/api/v1
```

### Build for USB

```bash
make usb-package
# Creates: portable-siem-YYYYMMDD.tar.gz
```

## Detection Rules

Rules live in `rules/` as JSON files and are loaded at startup.

```json
{
  "id": "ssh_brute_force",
  "name": "SSH Brute Force",
  "type": "threshold",
  "severity": "high",
  "source": "auth",
  "event_type": "login_failed",
  "threshold": 50,
  "window_secs": 300,
  "group_by": ["src_ip"],
  "actions": ["alert"],
  "mitre_ids": ["T1110"]
}
```

Rules can also be created via the API or the Rules page in the dashboard.

## Log Collection

| Source | Method | Notes |
|--------|--------|-------|
| Linux journald | `journalctl -f -o json` | Structured, includes unit/pid/priority |
| auth.log / secure | File tail | SSH logins, sudo, user changes |
| syslog | File tail | General system logs |
| Network devices | UDP syslog :514 | Routers, firewalls, switches, VPNs |
| Remote hosts | Linux agent | Deploy `siem-agent-linux` on target machines |

## Deploying the Linux Agent

```bash
# Copy agent binary to target host
scp dist/siem-agent-linux user@target:/tmp/

# Run agent
ssh user@target '/tmp/siem-agent-linux -server http://SIEM_IP:8080 -interval 10s'

# Or as a systemd service
ssh user@target 'sudo systemctl enable --now siem-agent'
```

## API Reference

```
GET  /health                    Health check
GET  /api/v1/events             List events (filters: host, source, event_type, severity, q, from, to)
POST /api/v1/events             Ingest a single event
GET  /api/v1/events/stats       Event statistics (last 24h)
GET  /api/v1/search?q=TEXT      Full-text search
GET  /api/v1/alerts             List alerts (filter by status)
PATCH /api/v1/alerts/:id/status Update alert status (open/acknowledged/resolved/false_positive)
GET  /api/v1/rules              List detection rules
POST /api/v1/rules              Create a rule
PUT  /api/v1/rules/:id          Update a rule
DELETE /api/v1/rules/:id        Delete a rule
GET  /api/v1/incidents          List incidents
POST /api/v1/incidents          Create incident
GET  /api/v1/incidents/:id      Get incident with timeline
PATCH /api/v1/incidents/:id/status  Update status
POST /api/v1/incidents/:id/timeline Add timeline entry
```

## Technology Stack

| Layer | Technology |
|-------|-----------|
| Backend | Go 1.22, Gin |
| Database | PostgreSQL 16 (events, alerts, incidents, IOCs, rules) |
| Cache / Correlation | Redis 7 (sliding window counters, pub/sub, state) |
| Messaging | NATS (agent communication) |
| Frontend | React 18, TypeScript, Vite, Tailwind CSS, Recharts |
| Agents | Go (Linux), stub ready for Rust (Windows) |
| Container | Docker, Docker Compose |

## Roadmap

- [x] Phase 1 — Log collection (Linux, network syslog)
- [x] Phase 2 — Search engine (full-text, field filters)
- [x] Phase 3 — Correlation engine (threshold, impossible travel)
- [x] Phase 4 — Detection rules (JSON, API-managed)
- [x] Phase 5 — Dashboard (Overview, Events, Alerts, Search, Incidents, Rules)
- [x] Phase 6 — Threat intelligence (feed downloader, IOC matching)
- [x] Phase 7 — Linux agent
- [x] Phase 8 — Incident management with timeline
- [ ] Windows agent (Event Log collection)
- [ ] MITRE ATT&CK navigator integration
- [ ] Asset inventory page
- [ ] User behavior analytics (UBA)
- [ ] AI investigation assistant
- [ ] GeoIP enrichment (MaxMind GeoLite2)
- [ ] Report generator (PDF export)
- [ ] Multi-tenancy support
