package main

import (
	"log/slog"
	"os"

	"github.com/tlinden/yadu"
)

type body string

type Ammo struct {
	Forweapon string
	Impact    int
	Cost      int
	Range     int
}

func (a *Ammo) LogValue() slog.Value {
	return slog.GroupValue(
		slog.String("Forweapon", a.Forweapon),
	)
}

type Enemy struct {
	Alive  bool
	Health int
	Name   string
	Body   body `yaml:"-"`
	Ammo   []Ammo
}

func (e *Enemy) LogValue() slog.Value {
	return slog.GroupValue(
		slog.String("name", e.Name),
	)
}

func removeTime(_ []string, a slog.Attr) slog.Attr {
	if a.Key == slog.TimeKey {
		return slog.Attr{}
	}
	return a
}

func main() {
	opts := &yadu.Options{
		Level:       slog.LevelDebug,
		ReplaceAttr: removeTime,
		AddSource:   true,
	}

	logger := slog.New(yadu.NewHandler(os.Stdout, opts))

	slog.SetDefault(logger)

	e := &Enemy{Alive: true, Health: 10, Name: "Bodo", Body: "body\nbody\n",
		Ammo: []Ammo{{Forweapon: "Railgun", Range: 400, Impact: 100, Cost: 100000}},
	}

	slog.Info("info", "enemy", e, "spawn", 199)
	slog.Info("info", "ammo", &Ammo{Forweapon: "axe", Impact: 1})
	slog.Info("connecting", "enemies", 100, "players", 2, "world", "600x800")
	slog.Debug("debug text")
	slog.Error("error")
}
