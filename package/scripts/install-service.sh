#!/bin/bash
set -e

INSTANCE="${1:-default}"
FRPC_BIN="${FRPC_BIN:-/usr/bin/frpc}"
DATA_DIR="/var/lib/frp/${INSTANCE}"
ENV_FILE="/etc/frp/${INSTANCE}.env"
SERVICE_NAME="frpc@${INSTANCE}"

echo "=== frpc Manager - Linux Service Installer ==="
echo "Instance: ${INSTANCE}"

# Check if running as root
if [ "$(id -u)" -ne 0 ]; then
    echo "ERROR: This script must be run as root (sudo)."
    exit 1
fi

# Create user if not exists
if ! id -u frp >/dev/null 2>&1; then
    echo "Creating frp user..."
    useradd --system --no-create-home --shell /sbin/nologin frp
fi

# Create directories
echo "Creating directories..."
mkdir -p /var/lib/frp
chown frp:frp /var/lib/frp
chmod 750 /var/lib/frp

mkdir -p "${DATA_DIR}"
chown frp:frp "${DATA_DIR}"
chmod 750 "${DATA_DIR}"

mkdir -p /etc/frp

# Create env file
if [ ! -f "${ENV_FILE}" ]; then
    cat > "${ENV_FILE}" << EOF
# frpc Manager environment configuration for instance "${INSTANCE}"
FRPC_WEB_PORT=7400
# FRPC_WEB_USER=admin
# FRPC_WEB_PASSWORD=changeme
EOF
    chown frp:frp "${ENV_FILE}"
    chmod 600 "${ENV_FILE}"
    echo "Created env file: ${ENV_FILE}"
fi

# Install binary
if [ -f "${FRPC_BIN}" ]; then
    cp "${FRPC_BIN}" /usr/bin/frpc
    chmod 755 /usr/bin/frpc
    echo "Installed frpc binary to /usr/bin/frpc"
else
    echo "WARNING: frpc binary not found at ${FRPC_BIN}"
    echo "Please copy the frpc binary to /usr/bin/frpc manually."
fi

# Install systemd unit
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
UNIT_SRC="${SCRIPT_DIR}/../systemd/frpc@.service"
if [ -f "${UNIT_SRC}" ]; then
    cp "${UNIT_SRC}" /etc/systemd/system/frpc@.service
else
    echo "ERROR: Cannot find frpc@.service template"
    exit 1
fi

# Reload and enable
systemctl daemon-reload
systemctl enable "${SERVICE_NAME}"

echo ""
echo "=== Installation Complete ==="
echo "Start the service:  systemctl start ${SERVICE_NAME}"
echo "Check status:       systemctl status ${SERVICE_NAME}"
echo "View logs:          journalctl -u ${SERVICE_NAME} -f"
echo ""
echo "Web UI will be available at: http://<server-ip>:7400"
echo "Edit environment:   ${ENV_FILE}"
