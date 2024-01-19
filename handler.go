package yadu

import (
	"bytes"
	"context"
	"io"
	"log/slog"
	"regexp"
	"slices"
	"strings"
	"sync"

	"gopkg.in/yaml.v3"

	"github.com/fatih/color"
)

// We use RFC datestring by default
const DefaultTimeFormat = "2006-01-02T03:04.05 MST"

// Default log level is INFO:
const defaultLevel = slog.LevelInfo

type Handler struct {
	writer      io.Writer
	mu          *sync.Mutex
	level       slog.Leveler
	groups      []string
	attrs       string
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
		fields[a.Key] = a.Value.Any()
		return true
	})

	tree := h.attrs

	if len(fields) > 0 {
		bytetree, err := yaml.Marshal(&fields)
		if err != nil {
			return err
		}

		tree = h.Postprocess(bytetree)
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
	buf.WriteString(color.WhiteString(tree))
	buf.WriteString("\n")

	h.mu.Lock()
	defer h.mu.Unlock()
	_, err := h.writer.Write(buf.Bytes())

	return err
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

	return h
}

// Enabled indicates whether the receiver logs at the given level.
func (h *Handler) Enabled(_ context.Context, l slog.Level) bool {
	return l >= h.level.Level()
}

// attributes plus attrs.
func (h *Handler) WithAttrs(attrs []slog.Attr) slog.Handler {
	if len(attrs) == 0 {
		return h
	}

	fields := make(map[string]interface{}, len(attrs))
	for _, a := range attrs {
		fields[a.Key] = a.Value.Any()
	}

	bytetree, err := yaml.Marshal(&fields)
	if err != nil {
		panic(err)
	}

	h2 := h.clone()

	h2.attrs += string(bytetree)
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
