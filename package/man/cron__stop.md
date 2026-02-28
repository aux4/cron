#### Description

Stop a running cron scheduler. Reads the PID file and sends a termination signal to the scheduler process.

#### Usage

```bash
aux4 cron stop
aux4 cron stop --port 9000
```

#### Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `--port` | Server port | `8421` |

#### Example

```bash
aux4 cron stop --port 8421
```
```text
{"status":"STOPPED","port":"8421","pid":12345}
```
