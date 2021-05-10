package i18n

import (
	"context"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"
)

type languageTagKey struct{}
type localizerKey struct{}

var languageTagCtxKey = languageTagKey{}
var localizerCtxKey = localizerKey{}

// WithLanguageTag returns a new `context.Context` that holds a `language.Tag`
func WithLanguageTag(ctx context.Context, languageTag language.Tag) context.Context {
	return context.WithValue(ctx, languageTagCtxKey, languageTag)
}

// LanguageTagFromContext returns the `language.Tag` previously associated with `ctx`, or
// `defaultLanguage` if no such `language.Tag` could be found.
func LanguageTagFromContext(ctx context.Context) language.Tag {
	if languageTag, ok := ctx.Value(languageTagCtxKey).(language.Tag); ok {
		return languageTag
	}

	return defaultLanguage
}

// WithLocalizer returns a new `context.Context` that holds a reference to
// the localizer
func WithLocalizer(ctx context.Context, localizer *i18n.Localizer) context.Context {
	return context.WithValue(ctx, localizerCtxKey, localizer)
}

// LocalizerFromContext returns the `*i18n.Localizer` previously associated with `ctx`, or
// `defaultLocalizer` if no such `*i18n.Localizer` could be found.
func LocalizerFromContext(ctx context.Context) *i18n.Localizer {
	if localizer, ok := ctx.Value(localizerCtxKey).(*i18n.Localizer); ok {
		return localizer
	}

	return defaultLocalizer
}
