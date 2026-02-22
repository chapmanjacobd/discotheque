package utils

import (
	"context"
	"fmt"
	"io"
	"log/slog"
)

type PlainHandler struct {
	Level slog.Leveler
	Out   io.Writer
	Attrs []slog.Attr
}

func (h *PlainHandler) Enabled(_ context.Context, level slog.Level) bool {
	return level >= h.Level.Level()
}

func (h *PlainHandler) Handle(_ context.Context, r slog.Record) error {
	msg := r.Message
	for _, a := range h.Attrs {
		msg += fmt.Sprintf("\n    %s=%v", a.Key, a.Value.Any())
	}
	r.Attrs(func(a slog.Attr) bool {
		msg += fmt.Sprintf("\n    %s=%v", a.Key, a.Value.Any())
		return true
	})
	_, err := fmt.Fprintln(h.Out, msg)
	return err
}

func (h *PlainHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &PlainHandler{
		Level: h.Level,
		Out:   h.Out,
		Attrs: append(h.Attrs, attrs...),
	}
}

func (h *PlainHandler) WithGroup(name string) slog.Handler {
	// Not implementing groups for now
	return h
}
