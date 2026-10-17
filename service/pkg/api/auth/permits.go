package auth

import "strings"

const (
	OpRead  = "read"
	OpWrite = "write"
)

func ScopeForKey(key string) string {
	if strings.HasPrefix(key, "public:") {
		return "public"
	}
	if strings.HasPrefix(key, "private:") {
		return "private"
	}
	if strings.HasPrefix(key, "unsafe:") {
		return "unsafe"
	}
	return "unsafe"
}

func KeyScopeAccessAllowed(key, role, op string) (string, bool) {
	scope := ScopeForKey(key)
	if role != "backend" && role != "frontend" {
		return scope, false
	}

	switch op {
	case OpRead:
		if role == "backend" {
			return scope, true
		}
		return scope, scope != "private"
	case OpWrite:
		if role == "backend" {
			return scope, true
		}
		return scope, scope == "unsafe"
	default:
		return scope, false
	}
}
