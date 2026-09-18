package userdirectory

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestNameReturnsCurrentValue(t *testing.T) {
	directory := New(func(context.Context) (map[int64]string, error) {
		return map[int64]string{7: "Иван Петров"}, nil
	}, time.Minute)

	if name, ok := directory.Name(context.Background(), 7); !ok || name != "Иван Петров" {
		t.Fatalf("получено %q, %v", name, ok)
	}
	if _, ok := directory.Name(context.Background(), 8); ok {
		t.Fatal("незнакомый идентификатор не должен считаться известным")
	}
	if _, ok := directory.Name(context.Background(), 0); ok {
		t.Fatal("нулевой идентификатор не должен считаться известным")
	}
}

func TestNamesAreCachedForTTL(t *testing.T) {
	calls := 0
	clock := time.Now()
	directory := New(func(context.Context) (map[int64]string, error) {
		calls++
		return map[int64]string{1: "старое"}, nil
	}, time.Minute)
	directory.now = func() time.Time { return clock }

	directory.Names(context.Background())
	directory.Names(context.Background())
	if calls != 1 {
		t.Fatalf("в пределах ttl должен быть один поход, было %d", calls)
	}

	clock = clock.Add(2 * time.Minute)
	directory.Names(context.Background())
	if calls != 2 {
		t.Fatalf("после истечения ttl нужен новый поход, было %d", calls)
	}
}

// Auth-service недоступен — это не повод терять имена в списке заказов.
func TestFetchFailureKeepsLastKnownNames(t *testing.T) {
	fail := false
	clock := time.Now()
	directory := New(func(context.Context) (map[int64]string, error) {
		if fail {
			return nil, errors.New("auth недоступен")
		}
		return map[int64]string{1: "Иван"}, nil
	}, time.Minute)
	directory.now = func() time.Time { return clock }

	directory.Names(context.Background())
	fail = true
	clock = clock.Add(2 * time.Minute)

	if name, ok := directory.Name(context.Background(), 1); !ok || name != "Иван" {
		t.Fatalf("после сбоя ожидалось последнее известное имя, получено %q, %v", name, ok)
	}
}

func TestNilDirectoryIsSafe(t *testing.T) {
	var directory *Directory
	if _, ok := directory.Name(context.Background(), 1); ok {
		t.Fatal("нулевой справочник не должен ничего знать")
	}
}
