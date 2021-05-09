package i18n

import (
	"context"
	"github.com/nicksnyder/go-i18n/v2/i18n"
)

type contextKey struct{}

var localizerKey = contextKey{}

// WithLocalizer returns a new `context.Context` that holds a reference to
// the localizer
func WithLocalizer(ctx context.Context, localizer *i18n.Localizer) context.Context {
	return context.WithValue(ctx, localizerKey, localizer)
}

// LocalizerFromContext returns the `*i18n.Localizer` previously associated with `ctx`, or
// `nil` if no such `*i18n.Localizer` could be found.
func LocalizerFromContext(ctx context.Context) *i18n.Localizer {
	if localizer, ok := ctx.Value(localizerKey).(*i18n.Localizer); ok {
		return localizer
	}

	return nil
}
