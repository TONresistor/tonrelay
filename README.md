# tonrelay

CLI tool for TON tunnel relay operators. Wraps [tunnel-node](https://github.com/ton-blockchain/adnl-tunnel) with one-command installation, service management, live monitoring, and configuration.

## Install

```bash
# Download tonrelay
curl -sSL https://raw.githubusercontent.com/TONresistor/tonrelay/main/scripts/install.sh | sudo sh

# Set up the relay (downloads tunnel-node, creates service, starts automatically)
sudo tonrelay install
```

Or build from source:

```bash
git clone https://github.com/TONresistor/tonrelay.git
cd tonrelay
go build -o tonrelay ./cmd/tonrelay/
sudo mv tonrelay /usr/local/bin/
sudo tonrelay install
```

## What it does

`tonrelay install` handles everything in one command:

- Downloads the tunnel-node binary from upstream releases
- Creates a dedicated system user (`tonrelay`)
- Generates config with auto-detected external IP
- Writes a hardened systemd service
- Starts the relay and verifies health
- Reports ADNL ID and connection details

## Usage

```bash
tonrelay status              # quick status check
tonrelay status --live       # interactive TUI dashboard
tonrelay logs -f             # follow relay logs
tonrelay info                # show ADNL ID, IP, port
tonrelay share               # print shareable relay info
tonrelay config show         # display config (keys masked)
tonrelay config set-ip       # update external IP in config
```

### Service management

```bash
sudo tonrelay start
sudo tonrelay stop
sudo tonrelay restart
```

### Updates

```bash
sudo tonrelay update         # download latest tunnel-node, backup old, restart
```

### Removal

```bash
sudo tonrelay uninstall      # stop service, remove binaries, config, user
```

## System layout

```
/usr/local/bin/tunnel-node          upstream binary (managed by tonrelay)
/etc/tonrelay/config.json           relay configuration
/var/lib/tonrelay/                  runtime data
/etc/systemd/system/tonrelay.service  systemd unit
```

The service runs as the unprivileged `tonrelay` user with systemd security hardening (NoNewPrivileges, ProtectSystem=strict, ProtectHome).

## Requirements

- Linux (amd64 or arm64)
- systemd
- Root access for install/service commands
- Open UDP port (default 17330)

## License

MIT
