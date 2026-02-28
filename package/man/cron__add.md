#### Description

Add a new scheduled task. The task is immediately scheduled if the scheduler is running. The entry is persisted to `.cron.json`.

#### Usage

```bash
aux4 cron add --name <name> --every <expr> --run <command>
aux4 cron add --name <name> --every <expr> --at <HH:MM> --run <command>
```

#### Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `--port` | Server port | `8421` |
| `--name` | Task name | (required) |
| `--every` | Schedule expression | (required) |
| `--at` | Time of day (HH:MM) | |
| `--run` | Command to execute | (required) |

#### Example

```bash
aux4 cron add --name backup --every "1 day" --at "02:00" --run "aux4 backup run"
```
```text
{"name":"backup","every":"1 day","at":"02:00","run":"aux4 backup run","state":"active"}
```

```bash
aux4 cron add --name heartbeat --every 30s --run "curl -s http://localhost/health"
```
```text
{"name":"heartbeat","every":"30s","run":"curl -s http://localhost/health","state":"active"}
```
