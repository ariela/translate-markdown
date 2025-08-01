package logging

import (
	"context"
	"log/slog"
	"sync"
)

// FingerCrossedHandlerは、指定されたレベル以上のログが出力されるまで
// ログメッセージをバッファリングするslog.Handlerです。
type FingerCrossedHandler struct {
	next         slog.Handler
	triggerLevel slog.Level
	buffer       []slog.Record
	triggered    bool
	mu           sync.Mutex
}

// NewFingerCrossedHandlerは新しいFingerCrossedHandlerを作成します。
func NewFingerCrossedHandler(next slog.Handler, triggerLevel slog.Level) *FingerCrossedHandler {
	return &FingerCrossedHandler{
		next:         next,
		triggerLevel: triggerLevel,
		buffer:       make([]slog.Record, 0, 100),
	}
}

// Enabledは、指定されたレベルのログが有効かどうかを返します。
// このハンドラは全てのレベルを一旦受け入れるため、常にtrueを返します。
func (h *FingerCrossedHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return true
}

// Handleはログレコードを処理します。
func (h *FingerCrossedHandler) Handle(ctx context.Context, r slog.Record) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.triggered {
		return h.next.Handle(ctx, r)
	}

	h.buffer = append(h.buffer, r)

	if r.Level >= h.triggerLevel {
		h.triggered = true
		// バッファリングされたログを全て出力
		for _, rec := range h.buffer {
			if err := h.next.Handle(ctx, rec); err != nil {
				return err
			}
		}
		// バッファをクリア
		h.buffer = nil
	}

	return nil
}

// WithAttrsは属性を持つ新しいハンドラを返します。
func (h *FingerCrossedHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &FingerCrossedHandler{
		next:         h.next.WithAttrs(attrs),
		triggerLevel: h.triggerLevel,
		buffer:       h.buffer,
		triggered:    h.triggered,
	}
}

// WithGroupはグループを持つ新しいハンドラを返します。
func (h *FingerCrossedHandler) WithGroup(name string) slog.Handler {
	return &FingerCrossedHandler{
		next:         h.next.WithGroup(name),
		triggerLevel: h.triggerLevel,
		buffer:       h.buffer,
		triggered:    h.triggered,
	}
}
