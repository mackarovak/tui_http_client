package requesteditor

import (
    "encoding/json"
    "net/url"

    "htui/internal/types"
)

type ValidationResult struct {
    Valid    bool
    Errors   []string
    Warnings []string
}

func Validate(r types.SavedRequest) ValidationResult {
    var res ValidationResult
    res.Valid = true

    if r.URL == "" {
        res.Valid = false
        res.Errors = append(res.Errors, "URL is required")
        return res
    }

    testURL := r.URL
    if len(testURL) > 0 && !hasScheme(testURL) {
        testURL = "https://" + testURL
    }
    if _, err := url.ParseRequestURI(testURL); err != nil {
        res.Valid = false
        res.Errors = append(res.Errors, "Invalid URL: "+err.Error())
    }

    if r.BodyMode == types.BodyModeJSON && r.Body != "" {
        if !json.Valid([]byte(r.Body)) {
            res.Valid = false
            res.Errors = append(res.Errors, "Invalid JSON in request body")
        }
    }

    if r.Auth.Type == types.AuthBearer && r.Auth.Token == "" {
        res.Warnings = append(res.Warnings, "Bearer token is empty")
    }

    return res
}

func hasScheme(u string) bool {
    return len(u) > 4 &&
        (u[:7] == "http://" || u[:8] == "https://")
}