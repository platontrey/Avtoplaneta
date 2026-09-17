package httpcache

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestMatches(t *testing.T) {
	cases := []struct {
		name        string
		ifNoneMatch string
		etag        string
		want        bool
	}{
		{"точное совпадение", `"v1"`, `"v1"`, true},
		{"другая версия", `"v1"`, `"v2"`, false},
		{"пустой заголовок", "", `"v1"`, false},
		{"звёздочка", "*", `"v1"`, true},
		{"список, совпадение во втором", `"v0", "v1"`, `"v1"`, true},
		{"список без совпадений", `"v0", "v2"`, `"v1"`, false},
		{"пробелы вокруг значений", `  "v1"  `, `"v1"`, true},
		{"клиент ослабил валидатор", `W/"v1"`, `"v1"`, true},
		{"сервер отдал слабый", `"v1"`, `W/"v1"`, true},
		{"оба слабые", `W/"v1"`, `W/"v1"`, true},
		{"слабый в списке", `"v0", W/"v1"`, `"v1"`, true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := Matches(tc.ifNoneMatch, tc.etag); got != tc.want {
				t.Fatalf("Matches(%q, %q) = %v, ожидалось %v", tc.ifNoneMatch, tc.etag, got, tc.want)
			}
		})
	}
}

func TestQuote(t *testing.T) {
	if got := Quote("v1"); got != `"v1"` {
		t.Fatalf("Quote не закавычил значение: %q", got)
	}
	if got := Quote(`"v1"`); got != `"v1"` {
		t.Fatalf("Quote закавычил дважды: %q", got)
	}
	if got := Quote(`W/"v1"`); got != `W/"v1"` {
		t.Fatalf("Quote испортил слабый валидатор: %q", got)
	}
}

func TestServeVersionedSetsHeaders(t *testing.T) {
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/", nil)

	if ServeVersioned(recorder, request, "v1", 5*time.Minute) {
		t.Fatal("без If-None-Match ответ не должен быть 304")
	}
	if got := recorder.Header().Get("ETag"); got != `"v1"` {
		t.Fatalf("ETag = %q", got)
	}
	if got := recorder.Header().Get("Cache-Control"); got != "public, max-age=300, must-revalidate" {
		t.Fatalf("Cache-Control = %q", got)
	}
}

func TestServeVersionedAnswers304(t *testing.T) {
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/", nil)
	request.Header.Set("If-None-Match", `W/"v1"`)

	if !ServeVersioned(recorder, request, "v1", time.Minute) {
		t.Fatal("совпавшая версия должна давать 304")
	}
	if recorder.Code != http.StatusNotModified {
		t.Fatalf("код ответа = %d", recorder.Code)
	}
	if recorder.Body.Len() != 0 {
		t.Fatalf("у 304 не должно быть тела, получено %d байт", recorder.Body.Len())
	}
}

func TestServeVersionedSkipsEmptyVersion(t *testing.T) {
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/", nil)
	request.Header.Set("If-None-Match", "*")

	if ServeVersioned(recorder, request, "", time.Minute) {
		t.Fatal("без версии условный ответ невозможен, даже по звёздочке")
	}
	if recorder.Header().Get("ETag") != "" {
		t.Fatal("пустой ETag выставлять нельзя")
	}
	if recorder.Header().Get("Cache-Control") != "" {
		t.Fatal("без версии не должно быть и Cache-Control")
	}
}
