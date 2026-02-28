# aux4/cron

User-friendly cron scheduler with aux4/jobs integration.

## Install

```bash
aux4 install aux4/cron
```

## Quick Start

```bash
# Start the scheduler
aux4 cron start

# Add a task that runs every 5 minutes
aux4 cron add --name cleanup --every "5 min" --run "aux4 cleanup run"

# Add a daily task at 2am
aux4 cron add --name backup --every "1 day" --at "02:00" --run "aux4 backup run"

# List all tasks
aux4 cron list

# View execution history
aux4 cron history --name backup

# Stop the scheduler
aux4 cron stop
```

## Commands

### Start the scheduler

```bash
aux4 cron start
aux4 cron start --port 9000
aux4 cron start --dir /var/data
```

### Stop the scheduler

```bash
aux4 cron stop
aux4 cron stop --port 9000
```

### Add a scheduled task

```bash
aux4 cron add --name cleanup --every "5 min" --run "rm -rf /tmp/cache/*"
aux4 cron add --name backup --every "1 day" --at "02:00" --run "aux4 backup run"
aux4 cron add --name report --every monday --at "09:00" --run "aux4 report generate"
aux4 cron add --name heartbeat --every 30s --run "curl -s http://localhost/health"
```

### Remove a task

```bash
aux4 cron remove --name cleanup
```

### Pause a task

```bash
aux4 cron pause --name backup
```

### Resume a task

```bash
aux4 cron resume --name backup
```

### List all tasks

```bash
aux4 cron list
```

### View execution history

```bash
aux4 cron history --name backup
aux4 cron history --name backup --limit 20
```

## Time Expressions

| Expression | Type | Meaning |
|---|---|---|
| `10s` | interval | Every 10 seconds |
| `30s` | interval | Every 30 seconds |
| `5 min` or `5min` | interval | Every 5 minutes |
| `15 min` | interval | Every 15 minutes |
| `2 hours` or `2h` | interval | Every 2 hours |
| `1 day` | daily | Every day (use `--at` for specific time, default midnight) |
| `monday` | weekly | Every Monday (use `--at` for time) |
| `tuesday`...`sunday` | weekly | Every specific weekday |
| `weekday` | weekly | Monday through Friday |
| `weekend` | weekly | Saturday and Sunday |
| `1 month` | monthly | Every month on the 1st |

Short forms: `10s`, `5min`, `2h`, `1d`
Long forms: `10 seconds`, `5 minutes`, `2 hours`, `1 day`
Singular/plural: `1 minute` = `1 min`
Day names are case-insensitive.

## Integration with aux4/jobs

When a cron entry triggers, it executes the command via `aux4 jobs run "<command>"`. This provides:

- Background execution
- Output capture (stdout/stderr)
- Job status tracking
- Job ID for each execution

View job details with:

```bash
aux4 jobs status <jobId>
aux4 jobs output <jobId>
```

## Persistence

- `.cron.json` stores all cron entries (created in the working directory)
- `.cron-history.json` stores execution history (last 1000 entries)
- On restart, the scheduler loads existing entries and resumes scheduling
