# Monero Node

Uses a Linux VM

## UFW

```sh
sudo ufw reset
sudo ufw default deny incoming
sudo ufw default deny outgoing
sudo ufw allow out to 20.30.40.5 port 8080 proto tcp
sudo ufw allow out to 20.30.40.5 port 9999
sudo ufw allow out to 20.30.40.1 port 9999
sudo ufw allow from 20.30.40.1 to any port 22 proto tcp
sudo ufw enable
sudo ufw status verbose
```

## Creating user

- daemon user

```sh
sudo adduser --system --home /home/monero-daemon --group monero-daemon
```

- wallet user

```sh
sudo adduser --system --home /home/monero-wallet --group monero-wallet
```

## Service

- Monero daemon

Copy the `monerod.conf` to the target.

```
[Unit]
Description=Monero Daemon
After=network.target

[Service]
Type=forking
PIDFile=/home/monero-daemon/monerod.pid
ExecStart=/home/monero-daemon/monero-x86_64-linux-gnu-v0.18.4.0/monerod --non-interactive --detach --config-file /home/monero-daemon/monerod.conf --pidfile=/home/monero-daemon/monerod.pid
User=monero-daemon
Group=monero-daemon
WorkingDirectory=/home/monero-daemon
Restart=always
RestartSec=5

[Install]
WantedBy=multi-user.target
```

- Monero wallet

Copy the `monero-wallet.conf` to the target.

```
[Unit]
Description=Monero Wallet
After=network.target

[Service]
Type=forking
PIDFile=/home/monero-wallet/monerod.pid
ExecStart=/home/monero-wallet/monero-x86_64-linux-gnu-v0.18.4.0/monero-wallet-rpc --non-interactive --detach --config-file /home/monero-wallet/monero-wallet.conf --pidfile=/home/monero-wallet/monerod.pid
User=monero-wallet
Group=monero-wallet
WorkingDirectory=/home/monero-wallet
Restart=always
RestartSec=5

[Install]
WantedBy=multi-user.target
```
