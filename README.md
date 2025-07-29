# command-exporter

A lightweight Prometheus exporter that runs shell commands at fixed intervals and exposes their numeric output as metrics.

## Features

- Executes arbitrary shell commands
- Parses `stdout` for float values
- Exposes metrics with `command`, `exit_code`, `stdout_summary`, `stderr_summary` labels
- Logs full execution details
- Configurable interval and output summary truncation

## Configuration

Create a `config.yaml` file:

```yaml
commands:
  - name: process_count
    command: ["bash", "-c", "ps aux | wc -l"]
    interval: "10s"
    stdout_limit: 50

  - name: load_avg
    command: ["bash", "-c", "cat /proc/loadavg | awk '{print $1}'"]
    interval: "5s"
    stdout_limit: 40
```

## Running

```bash
go build -o command-exporter ./cmd/command-exporter
./command-exporter
```

Exporter listens by default on `:9860`.

## Prometheus Scrape Config

Add this to your `prometheus.yml`:

```yaml
scrape_configs:
  - job_name: 'command-exporter'
    static_configs:
      - targets: ['localhost:9860']
```

## Metrics Output

Sample:

```
# HELP command_result Result of command execution
# TYPE command_result gauge
command_result{command="load_avg", exit_code="0", stdout_summary="0.13", stderr_summary=""} 0.13
command_result{command="process_count", exit_code="0", stdout_summary="142", stderr_summary=""} 142
```

## Docker

```Dockerfile
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY . .
RUN go build -o command-exporter ./cmd/command-exporter

FROM alpine:latest
WORKDIR /root/
COPY --from=builder /app/command-exporter .
COPY config.yaml .
EXPOSE 9860
ENTRYPOINT ["./command-exporter"]
```

## Notes

- Only the first float token from stdout is used as the metric value.
- If stdout cannot be parsed to float, `command_result` will return `1` and log a warning.
