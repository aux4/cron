#### Description

Start the cron scheduler as a background process. The scheduler loads any existing `.cron.json` file and resumes all active entries.

#### Usage

```bash
aux4 cron start
aux4 cron start --port 9000
aux4 cron start --dir /var/data
```

#### Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `--port` | Server port | `8421` |
| `--dir` | Working directory for cron files | `.` |

#### Example

```bash
aux4 cron start --port 8421
```
```text
cron scheduler started on port 8421
```
