#### Description

Remove a scheduled task. Stops the scheduler for the task and removes it from `.cron.json`.

#### Usage

```bash
aux4 cron remove --name <name>
```

#### Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `--port` | Server port | `8421` |
| `--name` | Task name | (required) |

#### Example

```bash
aux4 cron remove --name backup
```
```text
{"name":"backup","status":"REMOVED"}
```
