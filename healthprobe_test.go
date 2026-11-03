package healthprobe

import (
	"context"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHTTP(t *testing.T) {
	ok := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.Write([]byte(`{"status":"ok"}`)) }))
	defer ok.Close()
	bad := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(503) }))
	defer bad.Close()
	if err := Check(context.Background(), ok.URL, Options{Contains: `"ok"`}); err != nil {
		t.Fatal(err)
	}
	if err := Check(context.Background(), ok.URL, Options{Contains: "nope"}); err == nil {
		t.Fatal("expected body mismatch")
	}
	if err := Check(context.Background(), bad.URL, Options{}); err == nil {
		t.Fatal("expected 503 failure")
	}
}

func TestTCP(t *testing.T) {
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	addr := l.Addr().String()
	if err := Check(context.Background(), "tcp://"+addr, Options{}); err != nil {
		t.Fatal(err)
	}
	l.Close()
	if err := Check(context.Background(), "tcp://"+addr, Options{}); err == nil {
		t.Fatal("expected closed port failure")
	}
}
