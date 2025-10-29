package security

import (
	"encoding/json"
	"net/url"
	"strings"
)

// RedactSensitiveFieldsRaw takes a JSON payload and redacts sensitive fields.
// Any field commonly used for credentials (passwords, secrets, tokens, keys, usernames)
// is replaced with the literal string "redacted".
// Additionally, URL fields that may contain embedded credentials will have their userinfo redacted.
func RedactSensitiveFieldsRaw(raw json.RawMessage) json.RawMessage {
	if len(raw) == 0 {
		return raw
	}

	var data map[string]interface{}
	if err := json.Unmarshal(raw, &data); err != nil {
		// If it's not an object, return as-is
		return raw
	}

	// Known explicit keys to redact
	knownKeys := map[string]struct{}{
		"connection_password": {},
		"kafka_sasl_password": {},
		// Treat username as credential as per requirement
		"connection_username": {},
		"kafka_sasl_username": {},
		// Generic credential fields
		"password":        {},
		"secret":          {},
		"token":           {},
		"api_key":         {},
		"apikey":          {},
		"access_token":    {},
		"refresh_token":   {},
		"tls_secret_name": {}, // not a secret value itself but better avoid exposing exact name
	}

	// Helper to decide if a key is sensitive by substring
	isSensitiveKey := func(k string) bool {
		kLower := strings.ToLower(k)
		if _, ok := knownKeys[kLower]; ok {
			return true
		}
		// Substring heuristics
		substrings := []string{"password", "passwd", "secret", "token", "key", "credential", "username", "userinfo", "sasl"}
		for _, sub := range substrings {
			if strings.Contains(kLower, sub) {
				return true
			}
		}
		return false
	}

	// Redact fields
	for k, v := range data {
		// Redact simple sensitive fields
		if isSensitiveKey(k) {
			data[k] = "[REDACTED]"
			continue
		}

		// Special handling for URLs that might have embedded credentials
		// e.g., kong_admin_url: http://user:pass@host:8001
		if strings.HasSuffix(strings.ToLower(k), "url") {
			if str, ok := v.(string); ok && str != "" {
				if u, err := url.Parse(str); err == nil && u != nil {
					if u.User != nil {
						// Replace any user info with redacted values
						// If password is present, use UserPassword; else use User
						if pw, hasPw := u.User.Password(); hasPw {
							_ = pw // unused
							u.User = url.UserPassword("[REDACTED]", "[REDACTED]")
						} else {
							username := u.User.Username()
							_ = username // unused
							u.User = url.User("[REDACTED]")
						}
						data[k] = u.String()
					}
				}
			}
		}
	}

	// Marshal back to RawMessage
	b, err := json.Marshal(data)
	if err != nil {
		return raw
	}
	return json.RawMessage(b)
}
