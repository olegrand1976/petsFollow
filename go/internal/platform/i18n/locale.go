package i18n

import (
	"context"
	"strings"
)

type ctxKey struct{}

var Supported = []string{"fr", "nl", "en"}

func NormalizeLocale(raw string) string {
	raw = strings.TrimSpace(strings.ToLower(raw))
	if raw == "" {
		return "fr"
	}
	if idx := strings.IndexAny(raw, ",;"); idx >= 0 {
		raw = raw[:idx]
	}
	raw = strings.TrimSpace(raw)
	if i := strings.Index(raw, "-"); i >= 0 {
		raw = raw[:i]
	}
	for _, loc := range Supported {
		if raw == loc {
			return loc
		}
	}
	return "fr"
}

func ParseAcceptLanguage(header string) string {
	for _, part := range strings.Split(header, ",") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		if i := strings.Index(part, ";"); i >= 0 {
			part = strings.TrimSpace(part[:i])
		}
		return NormalizeLocale(part)
	}
	return "fr"
}

func WithLocale(ctx context.Context, locale string) context.Context {
	return context.WithValue(ctx, ctxKey{}, NormalizeLocale(locale))
}

func FromContext(ctx context.Context) string {
	if v, ok := ctx.Value(ctxKey{}).(string); ok && v != "" {
		return NormalizeLocale(v)
	}
	return "fr"
}
