package main

import (
	"context"
	"time"
)

type locationKey struct{}

var locationCtxKey = locationKey{}

// WithLocation returns a new `context.Context` that holds a `*time.Location`
func WithLocation(ctx context.Context, location *time.Location) context.Context {
	return context.WithValue(ctx, locationCtxKey, location)
}

// LocationFromContext returns the `*time.Location` previously associated with `ctx`, or
// `defaultLocation` if no such `*time.Location` could be found.
func LocationFromContext(ctx context.Context) *time.Location {
	if location, ok := ctx.Value(locationCtxKey).(*time.Location); ok {
		return location
	}

	return defaultLocation
}
