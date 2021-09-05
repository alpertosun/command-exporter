## Command Exporter

---

Command exporter runs commands with "name" parameters in shell.

### Command metrics example

---

http://127.0.0.1:9860/commands?target=any command name here


### Prometheus Config example
```
scrape_configs:
- job_name: 'command-exporter'
  metrics_path: /commands
  static_configs:
    - targets:
        - echo 1    # Target to run command with shell.
        - sleep 5    # Target to run command with shell.
  relabel_configs:
    - source_labels: [__address__]
      target_label: __param_target
    - source_labels: [__param_target]
      target_label: instance
    - target_label: __address__
      replacement: 127.0.0.1:9860 # The command exporter's real hostname:port.
```
