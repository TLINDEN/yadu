# YamlDumpHandler - a human readable yaml based slog.Handler

Example output:

```sh
2024-01-18T02:57.41 CET INFO: info 
    enemy:
        alive: true
        health: 10
        name: Bodo
        ammo:
            - forweapon: Railgun
              impact: 100
              cost: 100000
              range: 400
    spawn: 199
2024-01-18T02:57.41 CET INFO: connecting 
    enemies: 100
    players: 2
    world: 600x800
2024-01-18T02:57.41 CET DEBUG: debug text 
2024-01-18T02:57.41 CET ERROR: error 
```

See `example.go` for usage.
