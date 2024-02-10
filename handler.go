package yadu

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log/slog"
	"regexp"
	"runtime"
	"slices"
	"strings"
	"sync"

	"gopkg.in/yaml.v3"

	"github.com/fatih/color"
)

const VERSION = "0.1.2"

// We use RFC datestring by default
const DefaultTimeFormat = "2006-01-02T03:04.05 MST"

// Default log level is INFO:
const defaultLevel = slog.LevelInfo

// holds attributes added with logger.With()
type attributes map[string]interface{}

type Handler struct {
	writer      io.Writer
	mu          *sync.Mutex
	level       slog.Leveler
	groups      []string
	attrs       attributes
	timeFormat  string
	replaceAttr func(groups []string, a slog.Attr) slog.Attr
	addSource   bool
	indenter    *regexp.Regexp
}

// Options are options for the Yadu [log/slog.Handler].
//
// Level sets the minimum log level.
//
// ReplaceAttr is a function you  can define to customize how supplied
// attrs are being handled. It is empty by default, so nothing will be
// altered.
//
// Loglevel and message cannot be altered using ReplaceAttr. Timestamp
// can only be removed, see example. Keep in mind that everything will
// be passed to yaml.Marshal() in the end.
type Options struct {
	Level       slog.Leveler
	ReplaceAttr func(groups []string, a slog.Attr) slog.Attr
	TimeFormat  string
	AddSource   bool
	NoColor     bool
}

func (h *Handler) Handle(ctx context.Context, r slog.Record) error {
	level := r.Level.String() + ":"

	switch r.Level {
	case slog.LevelDebug:
		level = color.MagentaString(level)
	case slog.LevelInfo:
		level = color.BlueString(level)
	case slog.LevelWarn:
		level = color.YellowString(level)
	case slog.LevelError:
		level = color.RedString(level)
	}

	fields := make(map[string]interface{}, r.NumAttrs())
	r.Attrs(func(a slog.Attr) bool {
		//fields[a.Key] = a.Value.Any()
		a.Value = a.Value.Resolve()
		wa := make(map[string]interface{})
		h.appendAttr(wa, a)
		fields[a.Key] = wa[a.Key]
		return true
	})

	tree := ""
	source := ""

	if h.addSource && r.PC != 0 {
		source = h.getSource(r.PC)
	}

	if len(h.attrs) > 0 {
		bytetree, err := yaml.Marshal(h.attrs)
		if err != nil {
			return err
		}
		tree = h.Postprocess(bytetree)
	}

	if len(fields) > 0 {
		bytetree, err := yaml.Marshal(&fields)
		if err != nil {
			return err
		}

		tree += h.Postprocess(bytetree)
	}

	timeStr := ""
	timeAttr := slog.Time(slog.TimeKey, r.Time)

	if h.replaceAttr != nil {
		timeAttr = h.replaceAttr(nil, timeAttr)
	}

	if !r.Time.IsZero() && !timeAttr.Equal(slog.Attr{}) {
		timeStr = r.Time.Format(h.timeFormat)
	}

	msg := color.CyanString(r.Message)

	buf := bytes.Buffer{}

	if len(timeStr) > 0 {
		buf.WriteString(timeStr)
		buf.WriteString(" ")
	}
	buf.WriteString(level)
	buf.WriteString(" ")
	buf.WriteString(msg)
	buf.WriteString(" ")
	buf.WriteString(source)
	buf.WriteString(" ")
	buf.WriteString(color.WhiteString(tree))
	buf.WriteString("\n")

	h.mu.Lock()
	defer h.mu.Unlock()
	_, err := h.writer.Write(buf.Bytes())

	return err
}

// report caller source+line as yaml string
func (h *Handler) getSource(pc uintptr) string {
	fs := runtime.CallersFrames([]uintptr{pc})
	source, _ := fs.Next()
	return fmt.Sprintf("%s: %d", source.File, source.Line)
}

func (h *Handler) Postprocess(yamlstr []byte) string {
	return "\n    " + strings.TrimSpace(h.indenter.ReplaceAllString(string(yamlstr), "    "))
}

// NewHandler returns a [log/slog.Handler] using the receiver's options.
// Default options are used if opts is nil.
func NewHandler(out io.Writer, opts *Options) *Handler {
	if opts == nil {
		opts = &Options{}
	}

	h := &Handler{
		writer:      out,
		mu:          &sync.Mutex{},
		level:       opts.Level,
		timeFormat:  opts.TimeFormat,
		replaceAttr: opts.ReplaceAttr,
		addSource:   opts.AddSource,
		indenter:    regexp.MustCompile(`(?m)^`),
	}

	if opts.Level == nil {
		h.level = defaultLevel
	}

	if h.timeFormat == "" {
		h.timeFormat = DefaultTimeFormat
	}

	if opts.NoColor {
		color.NoColor = true
	}

	return h
}

// Enabled indicates whether the receiver logs at the given level.
func (h *Handler) Enabled(_ context.Context, l slog.Level) bool {
	return l >= h.level.Level()
}

// attributes plus attrs.
func (h *Handler) appendAttr(wa map[string]interface{}, a slog.Attr) {
	a.Value = a.Value.Resolve()

	if a.Value.Kind() == slog.KindGroup {
		attrs := a.Value.Group()
		name := ""
		if len(attrs) == 0 {
			return
		}

		if a.Key != "" {
			name = a.Key
			h.groups = append(h.groups, a.Key)
		}

		innerwa := make(map[string]interface{})
		for _, a := range attrs {
			h.appendAttr(innerwa, a)
		}
		wa[name] = innerwa

		if a.Key != "" && len(h.groups) > 0 {
			h.groups = h.groups[:len(h.groups)-1]
		}

		return
	}

	if h.replaceAttr != nil {
		a = h.replaceAttr(h.groups, a)
	}

	if !a.Equal(slog.Attr{}) {
		wa[a.Key] = a.Value.Any()
	}
}

// sub logger is to be created, possibly with attrs, add them to h.attrs
func (h *Handler) WithAttrs(attrs []slog.Attr) slog.Handler {
	if len(attrs) == 0 {
		return h
	}

	h2 := h.clone()

	wa := make(map[string]interface{})

	for _, a := range attrs {
		h2.appendAttr(wa, a)
	}

	h2.attrs = wa

	return h2
}

// WithGroup returns a new [log/slog.Handler] with name appended to the
// receiver's groups.
func (h *Handler) WithGroup(name string) slog.Handler {
	if name == "" {
		return h
	}
	h2 := h.clone()
	h2.groups = append(h2.groups, name)
	return h2
}

func (h *Handler) clone() *Handler {
	return &Handler{
		writer:      h.writer,
		mu:          h.mu,
		level:       h.level,
		groups:      slices.Clip(h.groups),
		attrs:       h.attrs,
		timeFormat:  h.timeFormat,
		replaceAttr: h.replaceAttr,
		addSource:   h.addSource,
		indenter:    h.indenter,
	}
}
