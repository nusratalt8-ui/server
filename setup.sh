#!/bin/bash
set -e

# ── Hollowpoint VPS setup ─────────────────────────────────────────────────────
# Run once on a fresh VPS. Idempotent — safe to re-run.
# Usage: bash setup.sh

TURN_SECRET="9fc07c3d5969e67bc1d3764f76814779ea4463ccf4868582c1813449abf6a3f8"

echo "[1/5] Installing coturn..."
apt-get update -qq && apt-get install -y -qq coturn

echo "[2/5] Discovering public IP..."
PUBLIC_IP=$(curl -sf https://api.ipify.org || curl -sf https://ifconfig.me/ip || curl -sf https://icanhazip.com)
if [ -z "$PUBLIC_IP" ]; then
  echo "ERROR: could not discover public IP" && exit 1
fi
echo "      Public IP: $PUBLIC_IP"

echo "[3/5] Writing coturn config..."
touch /var/log/turnserver.log
chown turnserver:turnserver /var/log/turnserver.log 2>/dev/null || true
cat > /etc/turnserver.conf << EOF
listening-port=3478
tls-listening-port=5349
listening-ip=0.0.0.0
external-ip=$PUBLIC_IP
relay-ip=$PUBLIC_IP
realm=$PUBLIC_IP
server-name=$PUBLIC_IP
fingerprint
lt-cred-mech
use-auth-secret
static-auth-secret=$TURN_SECRET
min-port=49152
max-port=65535
log-file=/var/log/turnserver.log
verbose
no-multicast-peers
EOF

echo "[4/5] Opening firewall ports..."
if command -v ufw &>/dev/null; then
  ufw allow 3478/udp
  ufw allow 3478/tcp
  ufw allow 5349/tcp
  ufw allow 49152:65535/udp
elif command -v iptables &>/dev/null; then
  iptables -I INPUT -p udp --dport 3478 -j ACCEPT
  iptables -I INPUT -p tcp --dport 3478 -j ACCEPT
  iptables -I INPUT -p tcp --dport 5349 -j ACCEPT
  iptables -I INPUT -p udp --dport 49152:65535 -j ACCEPT
fi

echo "[5/5] Enabling and starting coturn..."
sed -i 's/#TURNSERVER_ENABLED=1/TURNSERVER_ENABLED=1/' /etc/default/coturn 2>/dev/null || true
systemctl stop coturn 2>/dev/null || true
sleep 1
systemctl enable coturn
systemctl restart coturn
sleep 2

if systemctl is-active --quiet coturn; then
  echo ""
  echo "✓ coturn running on $PUBLIC_IP:3478"
else
  echo "ERROR: coturn failed to start"
  journalctl -u coturn --no-pager | tail -10
  exit 1
fi