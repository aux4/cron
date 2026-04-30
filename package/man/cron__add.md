#### Description

Add a new scheduled task. The task is immediately scheduled if the scheduler is running. The entry is persisted to `.cron.json`.

#### Usage

```bash
aux4 cron add --name <name> --every <expr> --run <command>
aux4 cron add --name <name> --every <expr> --at <time> --run <command>
aux4 cron add --name <name> --at <time> --run <command>
aux4 cron add --name <name> --in <delay> --run <command>
aux4 cron add --name <name> --every <expr> --max <n> --run <command>
```

#### Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `--port` | Server port | `8421` |
| `--name` | Task name | (required) |
| `--every` | Schedule expression (e.g. 10s, 15 min, 1 day, monday) | |
| `--at` | Time of day (HH:MM, 2pm, 2:30pm). Standalone: runs once at that time | |
| `--in` | One-time delay (e.g. 2 min, 30s, 1 hour). Runs once then auto-removes | |
| `--max` | Max executions before auto-remove | |
| `--run` | Command to execute | (required) |

At least one of `--every`, `--at`, or `--in` is required.

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

```bash
aux4 cron add --name reminder --in "5 min" --run "echo time is up"
```
```text
{"name":"reminder","in":"5 min","run":"echo time is up","state":"active"}
```

```bash
aux4 cron add --name retry --every 10s --max 3 --run "curl -s http://localhost/health"
```
```text
{"name":"retry","every":"10s","max":3,"run":"curl -s http://localhost/health","state":"active"}
```

```bash
aux4 cron add --name alert --at "2pm" --run "echo lunch time"
```
```text
{"name":"alert","at":"2pm","run":"echo lunch time","state":"active"}
```
