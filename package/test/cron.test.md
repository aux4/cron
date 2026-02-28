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
