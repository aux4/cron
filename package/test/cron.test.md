# cron

````beforeAll
rm -f .cron.json .cron-history.json
nohup aux4 cron start --port 18430 >/dev/null 2>&1 &
sleep 1
````

````afterAll
aux4 cron stop --port 18430
rm -f .cron.json .cron-history.json
````

## add

### should add a cron entry

````execute
aux4 cron add --name test-task --every 1s --run "echo hello" --port 18430 | jq .
````

````expect
{
  "name": "test-task",
  "every": "1s",
  "run": "echo hello",
  "state": "active"
}
````

### should fail to add duplicate entry

````execute
aux4 cron add --name test-task --every 1s --run "echo hello" --port 18430
````

````error:partial
already exists
````

## list

### should list cron entries

````execute
aux4 cron list --port 18430 | jq .
````

````expect
[
  {
    "name": "test-task",
    "every": "1s",
    "run": "echo hello",
    "state": "active"
  }
]
````

## pause

### should pause a cron entry

````execute
aux4 cron pause --name test-task --port 18430 | jq .
````

````expect
{
  "name": "test-task",
  "state": "paused"
}
````

### should show paused state in list

````execute
aux4 cron list --port 18430 | jq '.[0].state'
````

````expect
"paused"
````

## resume

### should resume a paused entry

````execute
aux4 cron resume --name test-task --port 18430 | jq .
````

````expect
{
  "name": "test-task",
  "state": "active"
}
````

### should show active state in list

````execute
aux4 cron list --port 18430 | jq '.[0].state'
````

````expect
"active"
````

## history

### should return empty history initially

````execute
aux4 cron history --name unknown-task --port 18430 | jq .
````

````expect
[]
````

## add with --in

### should add a one-time delayed task

````execute
aux4 cron add --name delayed-task --in "5 min" --run "echo delayed" --port 18430 | jq .
````

````expect
{
  "name": "delayed-task",
  "in": "5 min",
  "run": "echo delayed",
  "state": "active"
}
````

### should remove delayed task

````execute
aux4 cron remove --name delayed-task --port 18430 | jq .
````

````expect
{
  "name": "delayed-task",
  "status": "REMOVED"
}
````

## add with --max

### should add a task with max executions

````execute
aux4 cron add --name limited-task --every 1s --max 3 --run "echo limited" --port 18430 | jq .
````

````expect
{
  "name": "limited-task",
  "every": "1s",
  "max": 3,
  "run": "echo limited",
  "state": "active"
}
````

### should remove limited task

````execute
aux4 cron remove --name limited-task --port 18430 | jq .
````

````expect
{
  "name": "limited-task",
  "status": "REMOVED"
}
````

## add with --at standalone

### should add a one-time at task

````execute
aux4 cron add --name at-task --at "2pm" --run "echo at-time" --port 18430 | jq .
````

````expect
{
  "name": "at-task",
  "at": "2pm",
  "run": "echo at-time",
  "state": "active"
}
````

### should remove at task

````execute
aux4 cron remove --name at-task --port 18430 | jq .
````

````expect
{
  "name": "at-task",
  "status": "REMOVED"
}
````

## add validation

### should fail without schedule expression

````execute
aux4 cron add --name bad-task --run "echo fail" --port 18430
````

````error:partial
schedule expression is required
````

## remove

### should remove a cron entry

````execute
aux4 cron remove --name test-task --port 18430 | jq .
````

````expect
{
  "name": "test-task",
  "status": "REMOVED"
}
````

### should show empty list after remove

````execute
aux4 cron list --port 18430 | jq .
````

````expect
[]
````

### should fail to remove non-existent entry

````execute
aux4 cron remove --name test-task --port 18430
````

````error:partial
not found
````
