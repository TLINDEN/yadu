package yamldumphandler

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

const defaultTimeFormat = "2006-01-02T03:04.05 MST"

type YamlDumpHandler struct {
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

type YamlDumpHandlerOptions struct {
	Level       slog.Leveler
	ReplaceAttr func(groups []string, a slog.Attr) slog.Attr
	TimeFormat  string
	AddSource   bool
}

func (h *YamlDumpHandler) Handle(ctx context.Context, r slog.Record) error {
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

	//	h.l.Println(timeStr, level, msg, color.WhiteString(tree))

	return err
}

func (h *YamlDumpHandler) Postprocess(yamlstr []byte) string {
	return "\n    " + strings.TrimSpace(h.indenter.ReplaceAllString(string(yamlstr), "    "))
}

func NewYamlDumpHandler(out io.Writer, opts *YamlDumpHandlerOptions) *YamlDumpHandler {
	if opts == nil {
		opts = &YamlDumpHandlerOptions{}
	}

	h := &YamlDumpHandler{
		writer:      out,
		mu:          &sync.Mutex{},
		level:       opts.Level,
		timeFormat:  opts.TimeFormat,
		replaceAttr: opts.ReplaceAttr,
		addSource:   opts.AddSource,
		indenter:    regexp.MustCompile(`(?m)^`),
	}

	if h.timeFormat == "" {
		h.timeFormat = defaultTimeFormat
	}

	return h
}

// Enabled indicates whether the receiver logs at the given level.
func (h *YamlDumpHandler) Enabled(_ context.Context, l slog.Level) bool {
	return l >= h.level.Level()
}

// attributes plus attrs.
func (h *YamlDumpHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
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
func (h *YamlDumpHandler) WithGroup(name string) slog.Handler {
	if name == "" {
		return h
	}
	h2 := h.clone()
	h2.groups = append(h2.groups, name)
	return h2
}

func (h *YamlDumpHandler) clone() *YamlDumpHandler {
	return &YamlDumpHandler{
		writer:      h.writer,
		mu:          h.mu,
		level:       h.level,
		groups:      slices.Clip(h.groups),
		attrs:       h.attrs,
		timeFormat:  h.timeFormat,
		replaceAttr: h.replaceAttr,
		addSource:   h.addSource,
	}
}