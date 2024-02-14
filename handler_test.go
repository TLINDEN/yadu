package yadu_test

import (
	"bytes"
	"log/slog"
	"strings"
	"testing"
	"time"

	"github.com/fatih/color"
	"github.com/tlinden/yadu"
)

type body string

type Ammo struct {
	Forweapon string
	Impact    int
	Cost      int
	Range     float32
}

func (a *Ammo) LogValue() slog.Value {
	return slog.GroupValue(
		slog.String("Forweapon", "Use weapon: "+a.Forweapon),
	)
}

type Enemy struct {
	Alive  bool
	Health int
	Name   string
	Body   body `yaml:"-"`
	Ammo   []Ammo
}

type Point struct {
	y, Y, yes, n, N, no, True, False, on, off int
}

type Tests struct {
	name    string
	want    string
	negate  bool
	opts    yadu.Options
	with    slog.Attr
	haswith bool
}

const testTimeFormat = "03:04.05"

var tests = []Tests{
	{
		name:   "has-railgun",
		want:   "forweapon: Railgun",
		negate: false,
	},
	{
		name:   "has-ammo",
		want:   "ammo:",
		negate: false,
	},
	{
		name:   "has-ammo-logvaluer",
		want:   "Use weapon: Axe",
		negate: false,
	},
	{
		name:   "has-ammo-logvaluer-does-resolve",
		want:   "impact: 50", // should NOT appear
		negate: true,
	},
	{
		name:   "has-alive",
		want:   "alive: true",
		negate: false,
	},
	{
		name:   "has-no-body",
		want:   "body:",
		negate: true,
	},
	{
		name:   "has-time",
		want:   time.Now().Format(yadu.DefaultTimeFormat),
		negate: false,
	},
	{
		name: "has-no-time",
		want: time.Now().Format(yadu.DefaultTimeFormat),
		opts: yadu.Options{
			ReplaceAttr: removeTime,
		},
		negate: true,
	},
	{
		name: "has-custom-time",
		want: time.Now().Format(testTimeFormat),
		opts: yadu.Options{
			TimeFormat: testTimeFormat,
		},
		negate: false,
	},
	{
		name:   "with-group",
		want:   "pid:",
		negate: false,
		with: slog.Group("program_info",
			slog.Int("pid", 1923),
			slog.Bool("alive", true),
		),
		haswith: true,
	},
	{
		name:   "has-debug",
		want:   "DEBUG",
		negate: false,
		opts: yadu.Options{
			Level: slog.LevelDebug,
		},
	},
	{
		name:   "has-warn",
		want:   "WARN",
		negate: false,
		opts: yadu.Options{
			Level: slog.LevelWarn,
		},
	},
	{
		name:   "has-error",
		want:   "ERROR",
		negate: false,
		opts: yadu.Options{
			Level: slog.LevelError,
		},
	},
	{
		name:   "has-source",
		want:   "handler_test.go",
		negate: false,
		opts: yadu.Options{
			AddSource: true,
		},
	},
	{
		// check if output is NOT colored when disabling it
		name:   "disable-color",
		want:   "\x1b[0m",
		negate: true,
		opts: yadu.Options{
			NoColor: true,
		},
	},
	{
		// check if output is colored
		name:   "enable-color",
		want:   "\x1b[0m",
		negate: false,
	},
}

func GetEnemy() *Enemy {
	return &Enemy{Alive: true, Health: 10, Name: "Bodo", Body: "body\nbody\n",
		Ammo: []Ammo{{Forweapon: "Railgun", Range: 400, Impact: 100, Cost: 100000}},
	}
}

func GetAmmo() *Ammo {
	return &Ammo{Forweapon: "Axe", Range: 50, Impact: 1, Cost: 50}
}

func GetPoint() *Point {
	return &Point{}
}
func removeTime(_ []string, a slog.Attr) slog.Attr {
	if a.Key == slog.TimeKey {
		return slog.Attr{}
	}
	return a
}

func TestLogger(t *testing.T) {
	t.Parallel()

	for _, tt := range tests {
		var buf bytes.Buffer

		logger := slog.New(yadu.NewHandler(&buf, &tt.opts))

		if !tt.with.Equal(slog.Attr{}) {
			logger = logger.With(tt.with)
		}

		if !tt.opts.NoColor {
			color.NoColor = false
		}

		slog.SetDefault(logger)

		switch tt.opts.Level {
		case slog.LevelDebug:
			logger.Debug("attack", "enemy", GetEnemy(), "ammo", GetAmmo())
		case slog.LevelWarn:
			logger.Warn("attack", "enemy", GetEnemy(), "ammo", GetAmmo())
		case slog.LevelError:
			logger.Error("attack", "enemy", GetEnemy(), "ammo", GetAmmo())
		default:
			logger.Info("attack", "enemy", GetEnemy(), "ammo", GetAmmo())
		}

		got := buf.String()

		if strings.Contains(got, tt.want) == tt.negate {
			t.Errorf("test %s failed.\n want:\n%s\n got: %s\n", tt.name, tt.want, got)
		}

		buf.Reset()
	}
}

func TestYamlCleaner(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	logger := slog.New(yadu.NewHandler(&buf, &yadu.Options{}))
	slog.SetDefault(logger)

	logger.Info("got a point", "point", GetPoint())

	got := buf.String()

	bools := []string{"y:", "n:", "true:", "false:"}
	for _, want := range bools {
		if !strings.Contains(got, want) {
			t.Errorf("test TestYamlCleaner failed.\n want: %s:\n got: %s\n", want, got)
		}
	}
}
