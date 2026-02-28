#### Description

Resume a paused scheduled task. The task will be rescheduled and start executing again.

#### Usage

```bash
aux4 cron resume --name <name>
```

#### Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `--port` | Server port | `8421` |
| `--name` | Task name | (required) |

#### Example

```bash
aux4 cron resume --name backup
```
```text
{"name":"backup","state":"active"}
```
