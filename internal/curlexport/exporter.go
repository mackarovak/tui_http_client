package curlexport

import (
	"fmt"
	"net/url"
	"strings"

	"htui/internal/types"
)


func Build(r types.SavedRequest) string {
	var sb strings.Builder

	finalURL := buildURL(r)


	if r.Method == "GET" {
		sb.WriteString(fmt.Sprintf("curl \"%s\"", finalURL))
	} else {
		sb.WriteString(fmt.Sprintf("curl -X %s \"%s\"", r.Method, finalURL))
	}

	
	if r.Auth.Type == types.AuthBearer && r.Auth.Token != "" {
		sb.WriteString(fmt.Sprintf(" \\\n  -H \"Authorization: Bearer %s\"", r.Auth.Token))
	}

	
	for _, h := range r.Headers {
		if !h.Enabled || h.Key == "" {
			continue
		}
		sb.WriteString(fmt.Sprintf(" \\\n  -H \"%s: %s\"", h.Key, h.Value))
	}

	
	switch r.BodyMode {
	case types.BodyModeJSON:
		if r.Body != "" {
			sb.WriteString(" \\\n  -H \"Content-Type: application/json\"")
			sb.WriteString(fmt.Sprintf(" \\\n  -d '%s'", r.Body))
		}
	case types.BodyModeRawText:
		if r.Body != "" {
			sb.WriteString(fmt.Sprintf(" \\\n  -d '%s'", r.Body))
		}
	case types.BodyModeForm:
		if r.Body != "" {
			sb.WriteString(fmt.Sprintf(" \\\n  --data-urlencode '%s'", r.Body))
		}
	}

	return sb.String()
}


func buildURL(r types.SavedRequest) string {
	raw := r.URL
	if raw == "" {
		return ""
	}
	if !strings.Contains(raw, "://") {
		raw = "https://" + raw
	}

	u, err := url.Parse(raw)
	if err != nil {
		return raw
	}

	q := u.Query()
	for _, p := range r.Params {
		if p.Enabled && p.Key != "" {
			q.Set(p.Key, p.Value)
		}
	}
	u.RawQuery = q.Encode()
	return u.String()
}