package i18n

import (
	"embed"
	"encoding/json"
	"strings"
)

//go:embed locales/*.json
var localeFS embed.FS

var catalogs map[string]map[string]string

func init() {
	catalogs = make(map[string]map[string]string)
	for _, loc := range Supported {
		data, err := localeFS.ReadFile("locales/" + loc + ".json")
		if err != nil {
			panic("i18n: missing locale " + loc + ": " + err.Error())
		}
		var nested map[string]any
		if err := json.Unmarshal(data, &nested); err != nil {
			panic("i18n: invalid locale " + loc + ": " + err.Error())
		}
		catalogs[loc] = flatten("", nested)
	}
}

func flatten(prefix string, nested map[string]any) map[string]string {
	out := make(map[string]string)
	for k, v := range nested {
		key := k
		if prefix != "" {
			key = prefix + "." + k
		}
		switch val := v.(type) {
		case string:
			out[key] = val
		case map[string]any:
			for fk, fv := range flatten(key, val) {
				out[fk] = fv
			}
		}
	}
	return out
}

func T(locale, key string, vars map[string]string) string {
	locale = NormalizeLocale(locale)
	msg := lookup(locale, key)
	if msg == "" {
		msg = lookup("fr", key)
	}
	if msg == "" {
		return key
	}
	for k, v := range vars {
		msg = strings.ReplaceAll(msg, "{"+k+"}", v)
	}
	return msg
}

func lookup(locale, key string) string {
	if cat, ok := catalogs[locale]; ok {
		return cat[key]
	}
	return ""
}
