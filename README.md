# tonrelay

CLI for TON tunnel relay operators. One-command install, service management, live monitoring. Supports ADNL relay and optional HTTPS clearnet exit.

Built on [adnl-tunnel](https://github.com/TONresistor/adnl-tunnel).

## Quick start

```bash
curl -sSL https://raw.githubusercontent.com/TONresistor/tonrelay/main/scripts/install.sh | sudo sh
sudo tonrelay install
sudo tonrelay status --live
```

This installs the CLI, downloads tunnel-node, generates keys, creates a systemd service, and starts the relay automatically. The last command opens the live dashboard to verify everything is running. Open UDP port 17330 on your firewall.

## Install

```bash
curl -sSL https://raw.githubusercontent.com/TONresistor/tonrelay/main/scripts/install.sh | sudo sh
sudo tonrelay install
```

Options:

```
--clearnet-exit    Enable clearnet HTTPS exit mode (dual: relay + exit)
--ip <IP>          Set external IP (auto-detected if omitted)
--port <PORT>      Set UDP listen port (default: 17330)
--version <TAG>    Pin tunnel-node version (default: latest)
```

Build from source:

```bash
git clone https://github.com/TONresistor/tonrelay.git
cd tonrelay
go build -o tonrelay ./cmd/tonrelay/
sudo mv tonrelay /usr/local/bin/
sudo tonrelay install
```

## Clearnet exit

`--clearnet-exit` turns the relay into an HTTPS exit node. Clients running [tonutils-proxy](https://github.com/TONresistor/Tonutils-Proxy) with `--clearnet` route HTTPS traffic through the TON tunnel network.

The node continues relaying ADNL for .ton users. Existing clients are not affected.

Security (enforced by tunnel-node):
- HTTPS only (port 443), cleartext HTTP blocked
- DNS resolved at exit node, no client DNS leaks
- IP blacklist: RFC-1918, loopback, link-local, CGN, cloud metadata
- Fixed 1024-byte cell padding against traffic analysis
- Port 25 (SMTP) permanently blocked

## Usage

```bash
tonrelay status              # quick status check
tonrelay status --live       # interactive TUI dashboard
tonrelay info                # ADNL ID, IP, port, mode
tonrelay share               # shareable relay config (JSON)
tonrelay logs -f             # follow relay logs
tonrelay config show         # display config (keys masked)
tonrelay config set-ip <IP>  # update external IP
tonrelay version             # tonrelay and tunnel-node versions
```

```bash
sudo tonrelay start
sudo tonrelay stop
sudo tonrelay restart
sudo tonrelay update                  # download latest binary, restart
sudo tonrelay uninstall               # stop service, remove everything
sudo tonrelay uninstall --keep-config # preserve config on removal
```

Global flags:

```
--config <PATH>    Config file path (default: /etc/tonrelay/config.json)
--data-dir <PATH>  Data directory (default: /var/lib/tonrelay/)
--no-color         Disable colored output
--json             JSON output (status, info)
```

## Layout

```
/usr/local/bin/tunnel-node          binary (managed by tonrelay)
/etc/tonrelay/config.json           configuration
/var/lib/tonrelay/                  runtime data
/etc/systemd/system/tonrelay.service
```

Runs as unprivileged `tonrelay` user with systemd hardening (NoNewPrivileges, ProtectSystem=strict, ProtectHome).

## Requirements

- Linux (amd64 or arm64)
- systemd
- Root for install/service commands
- Open UDP port (default 17330)

## License

MIT
