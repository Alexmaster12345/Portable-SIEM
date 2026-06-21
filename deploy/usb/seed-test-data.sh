#!/usr/bin/env bash
# Seeds the SIEM with realistic test events so the dashboard shows data.

set -euo pipefail

SERVER="${1:-http://localhost:8888}"

echo "Seeding test data to $SERVER"
echo ""

send() {
    curl -sf -X POST "$SERVER/api/v1/events" \
        -H "Content-Type: application/json" \
        -d "$1" > /dev/null && echo "  OK: $2" || echo "  FAIL: $2"
}

# ---- SSH brute force (50+ failures from same IP) ----
echo "[1/6] SSH brute force events..."
for i in $(seq 1 55); do
    send "{
        \"timestamp\": \"$(date -u +%Y-%m-%dT%H:%M:%SZ)\",
        \"host\": \"server01\",
        \"source\": \"auth\",
        \"event_type\": \"login_failed\",
        \"severity\": \"medium\",
        \"message\": \"Failed password for admin from 1.2.3.4 port 2200$i ssh2\",
        \"fields\": {\"src_ip\": \"1.2.3.4\", \"username\": \"admin\", \"src_port\": \"2200$i\"}
    }" "ssh_fail #$i"
done

# ---- Successful logins ----
echo "[2/6] Successful logins..."
for user in alice bob carol; do
    send "{
        \"timestamp\": \"$(date -u +%Y-%m-%dT%H:%M:%SZ)\",
        \"host\": \"server01\",
        \"source\": \"auth\",
        \"event_type\": \"login_success\",
        \"severity\": \"info\",
        \"message\": \"Accepted publickey for $user from 10.0.0.5 port 443 ssh2\",
        \"fields\": {\"src_ip\": \"10.0.0.5\", \"username\": \"$user\"}
    }" "login $user"
done

# ---- Sudo / privilege escalation ----
echo "[3/6] Sudo events..."
for i in $(seq 1 6); do
    send "{
        \"timestamp\": \"$(date -u +%Y-%m-%dT%H:%M:%SZ)\",
        \"host\": \"server01\",
        \"source\": \"auth\",
        \"event_type\": \"sudo\",
        \"severity\": \"medium\",
        \"message\": \"bob : TTY=pts/0 ; PWD=/home/bob ; USER=root ; COMMAND=/bin/bash\",
        \"fields\": {\"username\": \"bob\"}
    }" "sudo #$i"
done

# ---- Firewall blocks (port scan simulation) ----
echo "[4/6] Firewall deny events (port scan)..."
for port in $(seq 20 140); do
    send "{
        \"timestamp\": \"$(date -u +%Y-%m-%dT%H:%M:%SZ)\",
        \"host\": \"fw-edge\",
        \"source\": \"network_syslog\",
        \"event_type\": \"firewall_deny\",
        \"severity\": \"medium\",
        \"message\": \"DENY SRC=5.5.5.5 DST=10.0.0.1 DPT=$port\",
        \"fields\": {\"src_ip\": \"5.5.5.5\", \"dst_ip\": \"10.0.0.1\", \"dst_port\": \"$port\"}
    }" "fw_deny port $port"
done

# ---- New user created ----
echo "[5/6] User account created..."
send "{
    \"timestamp\": \"$(date -u +%Y-%m-%dT%H:%M:%SZ)\",
    \"host\": \"server02\",
    \"source\": \"auth\",
    \"event_type\": \"user_change\",
    \"severity\": \"high\",
    \"message\": \"useradd: new user: name=backdoor, UID=1337, GID=1337\",
    \"fields\": {\"username\": \"backdoor\"}
}" "useradd"

# ---- Critical system events ----
echo "[6/6] Critical syslog events..."
for svc in sshd kernel firewalld; do
    send "{
        \"timestamp\": \"$(date -u +%Y-%m-%dT%H:%M:%SZ)\",
        \"host\": \"server0$RANDOM\",
        \"source\": \"syslog\",
        \"event_type\": \"syslog\",
        \"severity\": \"high\",
        \"message\": \"$svc: critical error detected - service may be compromised\",
        \"fields\": {\"service\": \"$svc\"}
    }" "critical $svc"
done

echo ""
echo "Done! Refresh the dashboard at http://localhost:3000"
echo ""
echo "Checking stats..."
curl -s "$SERVER/api/v1/events/stats" | python3 -m json.tool 2>/dev/null || \
curl -s "$SERVER/api/v1/events/stats"
