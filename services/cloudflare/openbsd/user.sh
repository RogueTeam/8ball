#!/bin/ksh
#
# Script to create the _cloudflared user for the http2socks service.
# This script must be run as root.

# Define user and group name
USER_NAME="_cloudflared"
COMMENT="Cloudflared service user" # Updated comment
HOME_DIR="/var/empty"
SHELL="/sbin/nologin"

echo "Attempting to create user: $USER_NAME"

# Check if the user already exists
if id "$USER_NAME" >/dev/null 2>&1; then
    echo "User '$USER_NAME' already exists. Skipping user creation."
else
    # Create the user with specified home directory, shell, and comment
    useradd -c "$COMMENT" -d "$HOME_DIR" -s "$SHELL" "$USER_NAME"

    # Check if useradd was successful
    if [ $? -eq 0 ]; then
        echo "User '$USER_NAME' created successfully."
        echo "Verify user details: id $USER_NAME"
    else
        echo "Error: Failed to create user '$USER_NAME'."
        exit 1
    fi
fi

echo "User creation script finished."
