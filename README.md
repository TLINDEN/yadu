[![Go Report Card](https://goreportcard.com/badge/github.com/tlinden/yadu)](https://goreportcard.com/report/github.com/tlinden/yadu) 
[![Actions](https://github.com/tlinden/yadu/actions/workflows/ci.yaml/badge.svg)](https://github.com/tlinden/yadu/actions)
[![Go Coverage](https://github.com/tlinden/yadu/wiki/coverage.svg)](https://raw.githack.com/wiki/tlinden/yadu/coverage.html)
![GitHub License](https://img.shields.io/github/license/tlinden/yadu)
[![GoDoc](https://godoc.org/github.com/tlinden/yadu?status.svg)](https://godoc.org/github.com/tlinden/yadu)

# yadu - a human readable yaml based slog.Handler

## Introduction

Package yadu provides a handler for the log/slog logging framework.

It generates a  mixture of text lines containing  the timestamp and
log message and a YAML dump of the provided attibutes.

## Log format

The log format generated by yadu looks like this:

```
2023-04-02T10:50.09 EDT LEVEL Message text
    foo: value
    bar: 12345
```

## Example

```go
logger := slog.New(yadu.NewHandler(os.Stdout, nil))

type body string

type Ammo struct {
        Forweapon string
        Impact    int
        Cost      int
        Range     int
}

type Enemy struct {
    Alive  bool
    Health int
    Name   string
    Body   body `yaml:"-"` // not printed
    Ammo   []Ammo
}

e := &Enemy{Alive: true, Health: 10, Name: "Bodo", Body: "body\nbody\n",
    Ammo: []Ammo{{Forweapon: "Railgun", Range: 400, Impact: 100, Cost: 100000}},
}

slog.Info("info", "enemy", e, "spawn", 199)
```

Output:

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
```

See `example/example.go` for usage.

## Installation

Execute this to add the module to your project:
```sh
go get github.com/tlinden/yadu
```

## Configuration

You can tweak the behavior of the handler as any other handler by using the Options struct:

```go
func removeTime(_ []string, a slog.Attr) slog.Attr {
        if a.Key == slog.TimeKey {
                return slog.Attr{}
        }
        return a
}

opts := &yadu.Options{
           Level: slog.LevelDebug,
           ReplaceAttr: removeTime,
        }
```

Pass this object to `yadu.NewHandler()`.

Because you can pass whole structs  to the logger which will be dumped
using YAML, there's also a way to exclude fields from being printed:

```go
type User struct {
  Id int
  User string
  Pass string `yaml:"-"`
}
```

If you're already using YAML tags for other purposes you can also just
add a  `LogValue()` method  to your  struct, which  will be  called by
slog. Refer to the slog documentation how to use it.

You can also modify the time format using `yadu.Options.TimeFormat`.

## Acknowledgements

I  wrote  most  of  the  code  with  the  help  of  the  [humane  slog
handler][humane]. Also helpfull was the  [guide to writing `slog` handlers][guide].

+ [humane slog handler][humane]
+ [A Guide to Writing `slog` Handlers][guide]
+ [A Comprehensive Guide to Structured Logging in Go][betterstack]

[humane]: https://github.com/telemachus/humane/tree/main
[guide]: https://github.com/golang/example/tree/master/slog-handler-guide
[betterstack]: https://betterstack.com/community/guides/logging/logging-in-go/

## LICENSE

This  module  is  published  under  the  terms  of  the  BSD  3-Clause
License. Please read the file LICENSE for details.

## Author

Thomas von Dein `<git |AT| daemon.de>`

