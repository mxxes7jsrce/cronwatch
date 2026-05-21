# cronwatch

Lightweight daemon that monitors cron job execution and sends alerts on missed or failed runs.

## Installation

```bash
go install github.com/yourusername/cronwatch@latest
```

Or build from source:

```bash
git clone https://github.com/yourusername/cronwatch.git && cd cronwatch && make build
```

## Usage

Define your jobs in a `cronwatch.yaml` config file:

```yaml
jobs:
  - name: daily-backup
    schedule: "0 2 * * *"
    timeout: 30m
    alert:
      email: ops@example.com

  - name: hourly-sync
    schedule: "0 * * * *"
    timeout: 5m
    alert:
      slack: "#alerts"
```

Start the daemon:

```bash
cronwatch --config /etc/cronwatch/cronwatch.yaml
```

Wrap your cron commands to report execution status:

```bash
# In your crontab
0 2 * * * cronwatch exec --job daily-backup /usr/local/bin/backup.sh
```

cronwatch will send an alert if the job exceeds its timeout, exits with a non-zero status, or fails to run within the expected schedule window.

### Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--config` | `./cronwatch.yaml` | Path to config file |
| `--log-level` | `info` | Log verbosity (`debug`, `info`, `warn`) |
| `--dry-run` | `false` | Validate config without starting daemon |

## License

MIT © 2024 cronwatch contributors