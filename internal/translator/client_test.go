package translator

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestTranslateSendsTextAndReturnsTranslation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/translate/yoda" {
			t.Errorf("path = %q, want /translate/yoda", r.URL.Path)
		}
		if r.Method != http.MethodPost {
			t.Errorf("method = %q, want POST", r.Method)
		}
		if got := r.FormValue("text"); got != "hello world" {
			t.Errorf("text = %q, want %q", got, "hello world")
		}
		w.Write([]byte(`{"contents": {"translated": "world hello, it is"}}`))
	}))
	defer server.Close()

	got, err := NewClient(server.URL, server.Client()).Translate(context.Background(), "hello world", Yoda)
	if err != nil {
		t.Fatal(err)
	}
	if got != "world hello, it is" {
		t.Errorf("translation = %q, want %q", got, "world hello, it is")
	}
}

func TestTranslateReturnsErrorWhenRateLimited(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
	}))
	defer server.Close()

	_, err := NewClient(server.URL, server.Client()).Translate(context.Background(), "hello", Shakespeare)
	if err == nil {
		t.Fatal("expected an error, got nil")
	}
}

func TestTranslateReturnsErrorOnEmptyTranslation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"contents": {"translated": ""}}`))
	}))
	defer server.Close()

	_, err := NewClient(server.URL, server.Client()).Translate(context.Background(), "hello", Shakespeare)
	if err == nil {
		t.Fatal("expected an error, got nil")
	}
}
