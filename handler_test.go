package yadu_test

import (
	"bytes"
	"log/slog"
	"strings"
	"testing"
	"time"

	"github.com/tlinden/yadu"
)

type body string

type Ammo struct {
	Forweapon string
	Impact    int
	Cost      int
	Range     float32
}

type Enemy struct {
	Alive  bool
	Health int
	Name   string
	Body   body `yaml:"-"`
	Ammo   []Ammo
}

type Tests struct {
	name   string
	want   string
	negate bool
	opts   *yadu.Options
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
		opts: &yadu.Options{
			ReplaceAttr: removeTime,
		},
		negate: true,
	},
	{
		name: "has-custom-time",
		want: time.Now().Format(testTimeFormat),
		opts: &yadu.Options{
			TimeFormat: testTimeFormat,
		},
		negate: false,
	},
	// FIXME: add WithGroup + WithAttr tests
}

func GetEnemy() *Enemy {
	return &Enemy{Alive: true, Health: 10, Name: "Bodo", Body: "body\nbody\n",
		Ammo: []Ammo{{Forweapon: "Railgun", Range: 400, Impact: 100, Cost: 100000}},
	}

}

func removeTime(_ []string, a slog.Attr) slog.Attr {
	if a.Key == slog.TimeKey {
		return slog.Attr{}
	}
	return a
}

func Test(t *testing.T) {
	t.Parallel()

	for _, tt := range tests {
		var buf bytes.Buffer

		logger := slog.New(yadu.NewHandler(&buf, tt.opts))

		logger.Info("attack", "enemy", GetEnemy())
		got := buf.String()

		if strings.Contains(got, tt.want) == tt.negate {
			t.Errorf("test %s failed.\n want:\n%s\n got: %s\n", tt.name, tt.want, got)
		}

		buf.Reset()
	}
}
