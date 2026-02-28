#### Description

Pause a scheduled task. The task configuration is kept but scheduling stops until resumed.

#### Usage

```bash
aux4 cron pause --name <name>
```

#### Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `--port` | Server port | `8421` |
| `--name` | Task name | (required) |

#### Example

```bash
aux4 cron pause --name backup
```
```text
{"name":"backup","state":"paused"}
```
