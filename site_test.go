package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestUploadSendsOnlyTheCodeWithTheKey(t *testing.T) {
	var got map[string]any
	var auth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth = r.Header.Get("Authorization")
		_ = json.NewDecoder(r.Body).Decode(&got)
		_, _ = w.Write([]byte(`{"member":"Melktruck","results":[{"ok":true,"status":"imported","characters":[{"name":"Ostbur"}]}]}`))
	}))
	defer srv.Close()
	r, err := upload(srv.URL, "ffu_key", []string{"FF1:00000000:abc"}, "WOW1", 1793120400)
	if err != nil || r.Member != "Melktruck" || r.Results[0].Characters[0].Name != "Ostbur" {
		t.Fatalf("got %+v, %v", r, err)
	}
	if auth != "Bearer ffu_key" {
		t.Fatalf("auth %q", auth)
	}
	if len(got) != 3 || got["account"] != "WOW1" {
		t.Fatalf("sent %+v", got)
	}
}

func TestADisconnectedKeySaysSo(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(401)
	}))
	defer srv.Close()
	if _, err := upload(srv.URL, "ffu_gone", []string{"FF1:00000000:abc"}, "WOW1", 0); err != errDisconnected {
		t.Fatalf("got %v", err)
	}
	// Forgetting a key the site no longer knows still counts as done.
	if err := revoke(srv.URL, "ffu_gone"); err != nil {
		t.Fatalf("revoke: %v", err)
	}
}

func TestRevokeReportsOtherFailures(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/uploader/disconnect" {
			t.Errorf("path %s", r.URL.Path)
		}
		w.WriteHeader(500)
	}))
	defer srv.Close()
	if err := revoke(srv.URL, "ffu_key"); err == nil {
		t.Fatal("a 500 should be an error")
	}
}
