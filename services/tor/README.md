# Deployment Guide: Tor on OpenBSD

This guide outlines the steps to install and configure the Tor daemon on OpenBSD. It will configure Tor to listen on two SOCKS ports, one for local applications and one for a wider network, with specific access policies. It also includes instructions for setting up automatic updates for the Tor package using `weekly.local`.

---

### Prerequisites

* **Root Access**: All commands in this guide assume you have root privileges on your OpenBSD machine. If you're not logged in as root, you'll need to switch using `su -`.
* **Network Configuration**: Ensure `20.30.40.1` is an IP address configured on your OpenBSD machine, and that the `20.30.40.0/24` network is relevant to your setup. Adjust these as needed.
* **Packet Filter (`pf`)**: Familiarity with `pf.conf` is recommended as you will need to add firewall rules.

---

### Deployment Steps

Follow these steps to set up your Tor daemon:

1.  **Install Tor**
    Use the `pkg_add` command to install the Tor package from the OpenBSD repositories.

    ```bash
    pkg_add tor
    ```

2.  **Configure Tor (`/etc/tor/torrc`)**
    Edit the Tor configuration file, `/etc/tor/torrc`, to define its SOCKS listener ports and access policies.

    Open `/etc/tor/torrc` for editing:

    ```bash
    vi /etc/tor/torrc
    ```

    Insert the following lines into the file. It's often good practice to add these towards the end of the file, or in a logical section if the file already contains configuration.

    ```
    # --- SOCKS Proxy Configuration ---
    # Listen for SOCKS connections on the localhost interface.
    SOCKSPort 127.0.0.1:9050

    # Listen for SOCKS connections on a specific network interface.
    # Replace 20.30.40.1 with your actual IP address if different.
    SOCKSPort 20.30.40.1:9050

    # Allow connections from localhost to the SOCKS proxy.
    SOCKSPolicy accept 127.0.0.1/32

    # Allow connections from the specified network to the SOCKS proxy.
    # Replace 20.30.40.0/24 with your actual network range if different.
    SOCKSPolicy accept 20.30.40.0/24

    # Reject all other connections to the SOCKS proxy.
    SOCKSPolicy reject *
    # --- End SOCKS Proxy Configuration ---
    ```

    Save and exit the file.

3.  **Configure Packet Filter (`/etc/pf.conf`)**
    You need to add firewall rules to allow incoming connections to the Tor SOCKS ports and to allow the `_tor` user (Tor's unprivileged user) to make outgoing connections.

    Edit your `pf.conf` file:

    ```bash
    vi /etc/pf.conf
    ```

    Add the following rules. Replace `egress` with your actual external interface name (e.g., `em0`, `vio0`) and ensure `20.30.40.1` matches the IP address used in your `torrc` configuration.

    ```pf
    # --- Rules for Tor Daemon ---
    # Allow outgoing connections from the _tor user to the internet.
    # This is essential for Tor to build circuits.
    pass out quick on egress proto { tcp udp } user _tor

    # Allow incoming TCP connections to the local SOCKS port (for http2socks or other local apps).
    pass in quick on lo0 proto tcp to 127.0.0.1 port 9050

    # Allow traffic from 20.30.40.0/24 to your machine's 9050 on the specified interface
    # Using 'quick' here is good practice to ensure this rule is the final decision if it matches.
    pass in quick on egress proto tcp from 20.30.40.0/24 to 20.30.40.1 port 9050 keep state
    # --- End of Tor Rules ---
    ```

    Save and exit `pf.conf`.
    Reload `pf` to apply the new rules:

    ```bash
    pfctl -f /etc/pf.conf
    ```

4.  **Enable and Start the Tor Service**
    Use `rcctl` to enable Tor to start automatically at boot and to start it immediately.

    ```bash
    rcctl enable tor
    rcctl start tor
    ```

    You can check the status of the service using:

    ```bash
    rcctl status tor
    ```

---

### Setting up Auto-Updates for Tor using `weekly.local`

To ensure your Tor daemon remains up-to-date, you can use the `weekly.local` script, which is executed automatically by `cron.weekly` jobs.

1.  **Edit `/etc/weekly.local`**
    Open the `weekly.local` file for editing:

    ```bash
    vi /etc/weekly.local
    ```

    Add the following lines to the file. This script will update installed packages and then restart the Tor service. The `-I` flag for `pkg_add` ensures it only updates already installed packages.

    ```sh
    #!/bin/sh
    PATH="/bin:/usr/bin:/sbin:/usr/sbin:/usr/local/bin:/usr/local/sbin"

    # Update installed packages and restart Tor service
    pkg_add -u -I && rcctl restart tor
    ```

2.  **Make `weekly.local` Executable**
    Ensure the script has execute permissions so `cron.weekly` can run it:

    ```bash
    chmod +x /etc/weekly.local
    ```

3.  **Restart Cron Daemon**
    For changes to `weekly.local` to be recognized by the `cron` daemon, it's best to restart it.

    ```bash
    rcctl restart cron
    ```

---

After completing these steps, your Tor daemon should be running and accessible on the configured SOCKS ports according to the defined policies, and it will attempt to update itself automatically once a week.
