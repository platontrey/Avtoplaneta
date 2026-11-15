package env

import (
	"os"
	"testing"
	"time"
)

func TestEnvString(t *testing.T) {
	os.Setenv("TEST_ENV_STR", "hello")
	defer os.Unsetenv("TEST_ENV_STR")

	if val := String("TEST_ENV_STR", "default"); val != "hello" {
		t.Errorf("expected 'hello', got '%s'", val)
	}
	if val := String("TEST_ENV_UNSET", "default"); val != "default" {
		t.Errorf("expected 'default', got '%s'", val)
	}
}

func TestEnvFirstNonEmpty(t *testing.T) {
	os.Setenv("TEST_ENV_SECOND", "second_val")
	defer os.Unsetenv("TEST_ENV_SECOND")

	val := FirstNonEmpty([]string{"TEST_ENV_FIRST", "TEST_ENV_SECOND"}, "fallback")
	if val != "second_val" {
		t.Errorf("expected 'second_val', got '%s'", val)
	}

	valFallback := FirstNonEmpty([]string{"UNSET1", "UNSET2"}, "fallback")
	if valFallback != "fallback" {
		t.Errorf("expected 'fallback', got '%s'", valFallback)
	}
}

func TestEnvInt(t *testing.T) {
	os.Setenv("TEST_ENV_INT", "42")
	defer os.Unsetenv("TEST_ENV_INT")

	if val := Int("TEST_ENV_INT", 10); val != 42 {
		t.Errorf("expected 42, got %d", val)
	}
	if val := Int("TEST_ENV_UNSET", 10); val != 10 {
		t.Errorf("expected 10, got %d", val)
	}
	os.Setenv("TEST_ENV_INVALID_INT", "not_a_number")
	defer os.Unsetenv("TEST_ENV_INVALID_INT")
	if val := Int("TEST_ENV_INVALID_INT", 10); val != 10 {
		t.Errorf("expected 10 on invalid, got %d", val)
	}
}

func TestEnvBool(t *testing.T) {
	os.Setenv("TEST_ENV_BOOL", "true")
	defer os.Unsetenv("TEST_ENV_BOOL")

	if val := Bool("TEST_ENV_BOOL", false); !val {
		t.Errorf("expected true, got %v", val)
	}
	if val := Bool("TEST_ENV_UNSET", true); !val {
		t.Errorf("expected true, got %v", val)
	}
}

func TestEnvDuration(t *testing.T) {
	os.Setenv("TEST_ENV_DUR", "5s")
	defer os.Unsetenv("TEST_ENV_DUR")

	if val := Duration("TEST_ENV_DUR", time.Second); val != 5*time.Second {
		t.Errorf("expected 5s, got %v", val)
	}
	if val := Duration("TEST_ENV_UNSET", time.Second); val != time.Second {
		t.Errorf("expected 1s, got %v", val)
	}
}
