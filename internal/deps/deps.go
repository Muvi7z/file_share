package deps

import "context"

type Logger interface {
	Info(ctx context.Context, message string, args ...any)
	Warn(ctx context.Context, message string, args ...any)
	Debug(ctx context.Context, message string, args ...any)
	Error(ctx context.Context, err error, args ...any)
}
