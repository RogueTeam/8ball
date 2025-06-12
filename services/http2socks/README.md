# Deployment Guide: http2socks Proxy on OpenBSD

This guide outlines the steps to deploy the http2socks proxy service on OpenBSD, ensuring it runs as a dedicated, unprivileged user within a secure chroot environment and is managed by rc.d and Packet Filter (pf).

---

Prerequisites

* **http2socks Binary**: You have the source code for http2socks and a Go development environment.
* **Root Access**: All commands in this guide assume you have root privileges on your OpenBSD machine. If you're not logged in as root, you'll need to switch using `su -`.
* **IP Addresses**: Replace `20.30.40.1` with the actual public IP address on your OpenBSD machine where http2socks should listen. Replace `egress` with your actual external network interface (e.g., `em0`, `vio0`).

---

Deployment Steps

Follow these steps in order to set up your http2socks service:

1. **Build the http2socks Binary**
   Navigate to your http2socks source directory and build the binary specifically for OpenBSD. This command will create the executable at `./build/http2socks`.

   ```bash
   GOOS=openbsd go build -o build/http2socks ./cmd/http2socks
   ```

   After building, move the binary to `/usr/local/bin` and set its ownership and permissions:

   ```bash
   mv build/http2socks /usr/local/bin/http2socks
   chown _proxy:_proxy /usr/local/bin/http2socks
   chmod 755 /usr/local/bin/http2socks
   ```

2. **Create the Service File (`/etc/rc.d/http2socks`)**
   This script tells OpenBSD how to start, stop, and manage your http2socks service.

   Create the file `/etc/rc.d/http2socks` with the contents of the `http2socks.service` file.

   Make the script executable: `chmod 755 /etc/rc.d/http2socks`

   Make sure to modify `/etc/rc.conf.local` in order to include the line with:

   ```
   http2socks_flags="-listen 20.30.40.1"
   ```

3. **Insert Firewall Rules into `/etc/pf.conf`**
   These rules open the necessary ports and allow the `_proxy` user to make outgoing connections.

   Edit `/etc/pf.conf` and insert the following lines, replacing `egress` with your actual external interface name:

   ```pf
   # --- Rules for http2socks proxy service ---
   # Allow incoming TCP connections to the http2socks proxy.
   # This permits clients to connect to your proxy on the specified IP and port.
   pass in quick on egress proto tcp to 20.30.40.1 port 8080

   # Allow outgoing TCP connections from the _proxy user to the local SOCKS proxy.
   # This rule is essential for http2socks to connect to the SOCKS proxy on localhost.
   pass out quick on lo0 proto tcp from (_proxy) to 127.0.0.1 port 9050
   # --- End of http2socks rules ---
   ```

4. **Run the User Creation Script**
   This script creates the dedicated `_proxy` user.

   Create a file (e.g., `create_proxy_user.sh`) with the contents of `user.sh`.

   Make it executable: `chmod +x create_proxy_user.sh`
   Run the script: `./create_proxy_user.sh`

5. **Run the Service Setup Script**
   This script reloads your firewall and enables/starts the http2socks service.

   Create a file (e.g., `setup_proxy_service.sh`) with the following content and run it as root:

   ```bash
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
   ```

   Make it executable: `chmod +x setup_proxy_service.sh`
   Run the script: `./setup_proxy_service.sh`

---

After completing these steps, your http2socks proxy should be running, secured, and ready for use.
