package logger

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"
)

const modulePrefix = "file_share/"

var defaultTagName tagName = "default"

type tagName string

type fields struct {
	f map[string]string
}

type contextHandler struct {
	slog.Handler
}

func newContextHandler(h slog.Handler) *contextHandler {
	return &contextHandler{Handler: h}
}

func (h *contextHandler) Handle(ctx context.Context, r slog.Record) error {
	if r.PC != 0 {
		fnName, source := resolveCaller(r.PC)
		if fnName != "" {
			r.AddAttrs(slog.String("func", fnName))
		}
		if source != "" {
			r.AddAttrs(slog.String("source", source))
		}
	}

	if ctx != nil {
		if t, ok := ctx.Value(defaultTagName).(fields); ok && len(t.f) > 0 {
			attrs := make([]slog.Attr, 0, len(t.f))
			for name, value := range t.f {
				attrs = append(attrs, slog.String(name, value))
			}
			r.AddAttrs(attrs...)
		}
	}

	return h.Handler.Handle(ctx, r)
}

func (h *contextHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &contextHandler{Handler: h.Handler.WithAttrs(attrs)}
}

func (h *contextHandler) WithGroup(name string) slog.Handler {
	return &contextHandler{Handler: h.Handler.WithGroup(name)}
}

type Logger struct {
	*slog.Logger
}

func New() *Logger {
	return NewWithWriter(os.Stdout)
}

func NewWithWriter(w io.Writer) *Logger {
	if w == nil {
		w = os.Stdout
	}
	jsonHandler := slog.NewJSONHandler(w, nil)
	ctxHandler := newContextHandler(jsonHandler)
	return &Logger{Logger: slog.New(ctxHandler)}
}

func (l *Logger) WithFields(ctx context.Context, f map[string]string) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}

	existing, _ := ctx.Value(defaultTagName).(fields)
	merged := make(map[string]string, len(existing.f)+len(f))
	for k, v := range existing.f {
		merged[k] = v
	}
	for k, v := range f {
		merged[k] = v
	}

	return context.WithValue(ctx, defaultTagName, fields{f: merged})
}

func (l *Logger) Info(ctx context.Context, message string, args ...any) {
	l.log(ctx, slog.LevelInfo, message, args...)
}

func (l *Logger) Warn(ctx context.Context, message string, args ...any) {
	l.log(ctx, slog.LevelWarn, message, args...)
}

func (l *Logger) Debug(ctx context.Context, message string, args ...any) {
	l.log(ctx, slog.LevelDebug, message, args...)
}

func (l *Logger) Error(ctx context.Context, err error, args ...any) {
	msg := "<nil>"
	if err != nil {
		msg = err.Error()
	}
	l.log(ctx, slog.LevelError, msg, args...)
}

func (l *Logger) log(ctx context.Context, level slog.Level, message string, args ...any) {
	if ctx == nil {
		ctx = context.Background()
	}
	if !l.Enabled(ctx, level) {
		return
	}

	var pcs [1]uintptr
	// Skip: runtime.Callers (0), (*Logger).log (1), Info/Warn/Debug/Error (2) -> real caller (3).
	runtime.Callers(3, pcs[:])

	formattedMsg, attrs := normalizeLogArgs(message, args)

	r := slog.NewRecord(time.Now(), level, formattedMsg, pcs[0])
	if len(attrs) > 0 {
		r.AddAttrs(attrs...)
	}

	_ = l.Handler().Handle(ctx, r)
}

func resolveCaller(pc uintptr) (fnName string, source string) {
	frames := runtime.CallersFrames([]uintptr{pc})
	frame, _ := frames.Next()
	if frame.Function != "" {
		fnName = cleanFuncName(frame.Function)
	}
	if frame.File != "" {
		source = cleanSourcePath(frame.File, frame.Line)
	}
	return fnName, source
}

func cleanFuncName(full string) string {
	return strings.TrimPrefix(full, modulePrefix)
}

func cleanSourcePath(file string, line int) string {
	slashed := filepath.ToSlash(file)
	if idx := strings.LastIndex(slashed, "/"+modulePrefix); idx >= 0 {
		slashed = slashed[idx+len("/"+modulePrefix):]
	} else {
		parts := strings.Split(slashed, "/")
		if len(parts) > 3 {
			slashed = strings.Join(parts[len(parts)-3:], "/")
		}
	}
	if line > 0 {
		return slashed + ":" + strconv.Itoa(line)
	}
	return slashed
}

func normalizeLogArgs(message string, args []any) (string, []slog.Attr) {
	if len(args) == 0 {
		return message, nil
	}

	if hasPrintfVerbs(message) && !containsSlogAttr(args) {
		return fmt.Sprintf(message, args...), nil
	}

	attrs := make([]slog.Attr, 0, len(args))
	argIdx := 0
	detailIdx := 0

	for i := 0; i < len(args); {
		switch v := args[i].(type) {
		case slog.Attr:
			attrs = append(attrs, v)
			i++
		case string:
			if i+1 < len(args) && isValidAttrKey(v) {
				attrs = append(attrs, slog.Any(v, args[i+1]))
				i += 2
			} else {
				argIdx++
				attrs = append(attrs, slog.String(indexedKey("arg", argIdx), v))
				i++
			}
		case error:
			detailIdx++
			if v != nil {
				attrs = append(attrs, slog.String(indexedKey("detail", detailIdx), v.Error()))
			}
			i++
		default:
			argIdx++
			attrs = append(attrs, slog.Any(indexedKey("arg", argIdx), v))
			i++
		}
	}

	return message, attrs
}

func indexedKey(base string, idx int) string {
	if idx <= 1 {
		return base
	}
	return base + "_" + strconv.Itoa(idx)
}

func containsSlogAttr(args []any) bool {
	for _, a := range args {
		if _, ok := a.(slog.Attr); ok {
			return true
		}
	}
	return false
}

func hasPrintfVerbs(s string) bool {
	for i := 0; i < len(s)-1; i++ {
		if s[i] != '%' {
			continue
		}
		next := s[i+1]
		if next == '%' {
			i++
			continue
		}
		if next == 'v' || next == 's' || next == 'd' || next == 'q' || next == 'w' ||
			next == 'f' || next == 't' || next == 'x' || next == 'X' || next == 'T' ||
			next == '+' || next == '#' || next == '.' || (next >= '0' && next <= '9') {
			return true
		}
	}
	return false
}

func isValidAttrKey(s string) bool {
	if s == "" || len(s) > 48 {
		return false
	}
	for i := 0; i < len(s); i++ {
		c := s[i]
		if (c >= 'a' && c <= 'z') ||
			(c >= 'A' && c <= 'Z') ||
			(c >= '0' && c <= '9') ||
			c == '_' || c == '-' || c == '.' {
			continue
		}
		return false
	}
	return true
}
