#!/bin/bash
# Setup script for Notion Desktop Reminder (Go version)

set -e

echo "=== Notion Desktop Reminder Setup (Go) ==="
echo ""

# Check if Go is installed
if ! command -v go &> /dev/null; then
    echo "Error: Go is not installed"
    echo "Install it with: sudo pacman -S go"
    exit 1
fi

GO_VERSION=$(go version)
echo "✓ $GO_VERSION found"

# Install libnotify if not present
if ! command -v notify-send &> /dev/null; then
    echo ""
    echo "notify-send not found. Installing libnotify..."
    sudo pacman -S libnotify
fi

echo "✓ notify-send available"

# Create config directory
CONFIG_DIR="$HOME/.config/notion-reminder"
mkdir -p "$CONFIG_DIR"

echo "✓ Config directory created"

# Create binary directory
BIN_DIR="$HOME/.local/bin"
mkdir -p "$BIN_DIR"

# Build the Go binary
echo ""
echo "Building Go binary..."
go build -o notion-reminder main.go

if [ $? -eq 0 ]; then
    echo "✓ Build successful"
else
    echo "✗ Build failed"
    exit 1
fi

# Move binary to bin directory
mv notion-reminder "$BIN_DIR/notion-reminder"
chmod +x "$BIN_DIR/notion-reminder"

echo "✓ Binary installed to: $BIN_DIR/notion-reminder"

# Create config file if it doesn't exist
CONFIG_FILE="$CONFIG_DIR/config.conf"
if [ ! -f "$CONFIG_FILE" ]; then
    cat > "$CONFIG_FILE" << 'EOF'
# Notion Reminder Configuration
# 
# Get your integration token from: https://www.notion.so/my-integrations
# Get your database ID from the database URL (the 32-character code)

NOTION_TOKEN=your_integration_token_here
DATABASE_ID=your_database_id_here
EOF
    echo "✓ Config file created at: $CONFIG_FILE"
else
    echo "✓ Config file already exists at: $CONFIG_FILE"
fi

# Create systemd user service
SYSTEMD_DIR="$HOME/.config/systemd/user"
mkdir -p "$SYSTEMD_DIR"

cat > "$SYSTEMD_DIR/notion-reminder.service" << EOF
[Unit]
Description=Notion Desktop Reminder
After=graphical-session.target

[Service]
Type=oneshot
ExecStart=$BIN_DIR/notion-reminder
StandardOutput=journal
StandardError=journal

[Install]
WantedBy=default.target
EOF

echo "✓ Systemd service created"

echo ""
echo "=== Setup Complete! ==="
echo ""
echo "Next steps:"
echo "1. Edit the config file with your Notion credentials:"
echo "   nano $CONFIG_FILE"
echo ""
echo "2. Test the program manually:"
echo "   notion-reminder"
echo ""
echo "3. Enable automatic startup (optional):"
echo "   systemctl --user enable notion-reminder.service"
echo "   systemctl --user start notion-reminder.service"
echo ""
echo "Note: Make sure $BIN_DIR is in your PATH"
echo "Add this to your ~/.bashrc or ~/.zshrc if needed:"
echo "   export PATH=\"\$HOME/.local/bin:\$PATH\""
echo ""
echo "Binary size: $(du -h $BIN_DIR/notion-reminder | cut -f1)"
