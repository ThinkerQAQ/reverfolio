package echo

import (
	"context"
	"encoding/json"
	"strings"
)

type rawCaptureKey struct{}

func captureRaw(ctx context.Context, endpoint string, data []byte) {
	recorder, ok := ctx.Value(rawCaptureKey{}).(map[string]json.RawMessage)
	if !ok {
		return
	}
	var value any
	if json.Unmarshal(data, &value) != nil {
		return
	}
	var clean func(any) any
	clean = func(v any) any {
		switch node := v.(type) {
		case map[string]any:
			for key, value := range node {
				lower := strings.ToLower(key)
				if strings.Contains(lower, "token") || strings.Contains(lower, "secret") || strings.Contains(lower, "password") || strings.Contains(lower, "cookie") || lower == "authorization" {
					node[key] = "[redacted]"
				} else {
					node[key] = clean(value)
				}
			}
		case []any:
			for i := range node {
				node[i] = clean(node[i])
			}
		}
		return v
	}
	if raw, err := json.Marshal(clean(value)); err == nil {
		recorder[endpoint] = raw
	}
}
