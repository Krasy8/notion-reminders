#!/bin/bash
# Restore Notion Reminders mako configuration
# Run this if Omarchy update overwrites your mako config

echo "Restoring Notion Reminders mako configuration..."

# Check if backup exists
if [ ! -f ~/.config/omarchy/current/theme/mako.ini.backup ]; then
    echo "Error: Backup file not found!"
    exit 1
fi

# Check if Notion Reminders section exists in current config
if grep -q "\[app-name=\"Notion Reminders\"\]" ~/.config/omarchy/current/theme/mako.ini; then
    echo "✓ Notion Reminders configuration already exists in mako.ini"
    exit 0
fi

# Append Notion Reminders configuration
echo "" >> ~/.config/omarchy/current/theme/mako.ini
echo "[app-name=\"Notion Reminders\"]" >> ~/.config/omarchy/current/theme/mako.ini
echo "default-timeout=0" >> ~/.config/omarchy/current/theme/mako.ini
echo "on-button-left=exec sh -c 'omarchy-launch-webapp https://www.notion.so/2a9d40cb309b80c1a7adc2896f8d0713; makoctl dismiss'" >> ~/.config/omarchy/current/theme/mako.ini

echo "✓ Configuration restored!"
echo "Reloading mako..."
makoctl reload

echo "✓ Done! Test with: notion-reminder"
