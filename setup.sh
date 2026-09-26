#!/bin/bash
set -e

SCRIPT_NAME="notify"
DAEMON_NAME="agent-notifyd"
INSTALL_DIR="/usr/local/bin"
CONFIG_DIR="$HOME/.config/agent-notify"
ASSETS_DIR="$HOME/.local/share/agent-notify/assets"
REPO_DIR="$(cd "$(dirname "$0")" && pwd)"

echo "=========================================="
echo "  Agent Notify - Installation Script"
echo "=========================================="
echo ""

# Check if Go is available for building
if command -v go &> /dev/null; then
    echo "[1/5] Building from source..."
    go build -o "$REPO_DIR/$SCRIPT_NAME" "$REPO_DIR/cmd/agent-notify"
    go build -o "$REPO_DIR/$DAEMON_NAME" "$REPO_DIR/cmd/agent-notifyd"
    echo "      Build complete."
else
    echo "[1/5] Go not found. Using pre-built binaries..."
    if [ ! -f "$REPO_DIR/$SCRIPT_NAME" ]; then
        echo "ERROR: No binary found and Go is not installed."
        echo "Please install Go or provide a pre-built binary."
        exit 1
    fi
fi

# Install binaries
echo "[2/5] Installing binaries to $INSTALL_DIR..."
cp "$REPO_DIR/$SCRIPT_NAME" "$INSTALL_DIR/$SCRIPT_NAME"
cp "$REPO_DIR/$DAEMON_NAME" "$INSTALL_DIR/$DAEMON_NAME"
# Also install agent-notify for backward compatibility
cp "$REPO_DIR/$SCRIPT_NAME" "$INSTALL_DIR/agent-notify"
chmod +x "$INSTALL_DIR/$SCRIPT_NAME"
chmod +x "$INSTALL_DIR/$DAEMON_NAME"
chmod +x "$INSTALL_DIR/agent-notify"
echo "      Done."

# Install assets (sounds)
echo "[3/5] Installing sound assets..."
mkdir -p "$ASSETS_DIR/sounds"
if [ -d "$REPO_DIR/assets/sounds" ]; then
    cp "$REPO_DIR/assets/sounds"/*.wav "$ASSETS_DIR/sounds/" 2>/dev/null || true
fi
echo "      Done."

# Install config template
echo "[4/5] Installing configuration..."
mkdir -p "$CONFIG_DIR"
if [ -f "$REPO_DIR/configs/config.yaml.example" ]; then
    cp "$REPO_DIR/configs/config.yaml.example" "$CONFIG_DIR/config.yaml"
fi
echo "      Config at $CONFIG_DIR/config.yaml"

# Create config if it doesn't exist
if [ ! -f "$CONFIG_DIR/config.yaml" ] || [ -s "$CONFIG_DIR/config.yaml" ]; then
    if [ ! -f "$CONFIG_DIR/config.yaml" ]; then
        cp "$REPO_DIR/configs/config.yaml.example" "$CONFIG_DIR/config.yaml"
    fi
fi

# Verify installation
echo ""
echo "[5/5] Verifying installation..."
if command -v notify &> /dev/null; then
    echo "      notify: $(notify --version 2>/dev/null || echo 'installed')"
else
    echo "      ERROR: notify not found in PATH"
fi
if command -v agent-notifyd &> /dev/null; then
    echo "      agent-notifyd: $(agent-notifyd --version 2>/dev/null || echo 'installed')"
else
    echo "      ERROR: agent-notifyd not found in PATH"
fi

echo ""
echo "=========================================="
echo "  Installation Complete!"
echo "=========================================="
echo ""
echo "Usage:"
echo "  notify -t \"Hey there I am User\"    # TTS"
echo "  notify -s success                 # Play sound"
echo "  notify -n \"Title\" -m \"Message\"   # Desktop notification"
echo "  notify work_done                  # Predefined event"
echo "  notify --daemon-status            # Check daemon"
echo ""
echo "Daemon:"
echo "  agent-notifyd start               # Start daemon"
echo "  agent-notifyd status              # Check status"
echo ""
echo "Config:  $CONFIG_DIR/config.yaml"
echo "Assets:  $ASSETS_DIR/sounds/"
echo ""