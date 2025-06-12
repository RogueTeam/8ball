#!/bin/ksh
#
# Script to reload Packet Filter (pf) rules and manage the http2socks service.
# This script must be run as root.

echo "--- Starting Firewall and Service Setup ---"

# --- 1. Reload Packet Filter (pf) rules ---
echo "Reloading Packet Filter rules..."
pfctl -f /etc/pf.conf
if [ $? -eq 0 ]; then
    echo "Packet Filter rules reloaded successfully."
else
    echo "Error: Failed to reload Packet Filter rules. Please check /etc/pf.conf for errors."
    exit 1
fi

# --- 2. Enable the http2socks service to start at boot ---
echo "Enabling http2socks service to start at boot..."
rcctl enable http2socks
if [ $? -eq 0 ]; then
    echo "http2socks service enabled."
else
    echo "Error: Failed to enable http2socks service."
    exit 1
fi

# --- 3. Start the http2socks service immediately ---
echo "Starting http2socks service..."
rcctl start http2socks
if [ $? -eq 0 ]; then
    echo "http2socks service started."
    echo "Verify service status: rcctl status http2socks"
else
    echo "Error: Failed to start http2socks service. Check logs (dmesg, /var/log/messages) for details."
    exit 1
fi

echo "--- Firewall and Service Setup Finished ---"
