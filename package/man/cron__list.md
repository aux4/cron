#### Description

List all scheduled tasks with their current state.

#### Usage

```bash
aux4 cron list
```

#### Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `--port` | Server port | `8421` |

#### Example

```bash
aux4 cron list | jq .
```
```json
[
  {
    "name": "backup",
    "every": "1 day",
    "at": "02:00",
    "run": "aux4 backup run",
    "state": "active"
  },
  {
    "name": "heartbeat",
    "every": "30s",
    "run": "curl -s http://localhost/health",
    "state": "paused"
  }
]
```
