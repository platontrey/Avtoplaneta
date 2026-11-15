package env

import (
	"os"
	"strconv"
	"strings"
	"time"
)

// String returns the environment variable value for the given key,
// or fallback if key is unset or empty.
func String(key, fallback string) string {
	if val, ok := os.LookupEnv(key); ok {
		val = strings.TrimSpace(val)
		if val != "" {
			return val
		}
	}
	return fallback
}

// FirstNonEmpty returns the first non-empty value among multiple environment variable keys,
// or fallback if none are set.
func FirstNonEmpty(keys []string, fallback string) string {
	for _, k := range keys {
		if val, ok := os.LookupEnv(k); ok {
			val = strings.TrimSpace(val)
			if val != "" {
				return val
			}
		}
	}
	return fallback
}

// Int returns the integer value of the environment variable,
// or fallback if unset, empty, or unparseable.
func Int(key string, fallback int) int {
	if val, ok := os.LookupEnv(key); ok {
		val = strings.TrimSpace(val)
		if val != "" {
			if n, err := strconv.Atoi(val); err == nil {
				return n
			}
		}
	}
	return fallback
}

// Bool returns the boolean value of the environment variable,
// or fallback if unset, empty, or unparseable.
func Bool(key string, fallback bool) bool {
	if val, ok := os.LookupEnv(key); ok {
		val = strings.TrimSpace(val)
		if val != "" {
			if b, err := strconv.ParseBool(val); err == nil {
				return b
			}
		}
	}
	return fallback
}

// Duration returns the duration value of the environment variable,
// or fallback if unset, empty, or unparseable.
func Duration(key string, fallback time.Duration) time.Duration {
	if val, ok := os.LookupEnv(key); ok {
		val = strings.TrimSpace(val)
		if val != "" {
			if d, err := time.ParseDuration(val); err == nil {
				return d
			}
		}
	}
	return fallback
}
