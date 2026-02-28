#### Description

Show execution history for a scheduled task. Each entry includes the job ID from aux4/jobs, timestamp, and trigger status.

#### Usage

```bash
aux4 cron history --name <name>
aux4 cron history --name <name> --limit 20
```

#### Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `--port` | Server port | `8421` |
| `--name` | Task name | (required) |
| `--limit` | Max entries to show | `10` |

#### Example

```bash
aux4 cron history --name backup | jq .
```
```json
[
  {
    "name": "backup",
    "jobId": "42",
    "timestamp": "2025-01-15T02:00:00Z",
    "status": "TRIGGERED"
  }
]
```
