package api

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestManagementAPI_HealthCheck_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/-/healthy" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	m, err := newManagementClient(srv.URL)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := m.HealthCheck(context.Background()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestManagementAPI_HealthCheck_NotOK(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
		w.Write([]byte("not healthy"))
	}))
	defer srv.Close()

	m, err := newManagementClient(srv.URL)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := m.HealthCheck(context.Background()); err == nil {
		t.Fatal("expected error for non-200 status")
	}
}

func TestManagementAPI_HealthCheck_ConnError(t *testing.T) {
	m, err := newManagementClient("http://127.0.0.1:1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := m.HealthCheck(context.Background()); err == nil {
		t.Fatal("expected error for connection failure")
	}
}

func TestManagementAPI_ReadinessCheck_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	m, err := newManagementClient(srv.URL)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := m.ReadinessCheck(context.Background()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestManagementAPI_ReadinessCheck_NotOK(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer srv.Close()

	m, err := newManagementClient(srv.URL)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := m.ReadinessCheck(context.Background()); err == nil {
		t.Fatal("expected error for non-200 status")
	}
}

func TestManagementAPI_Reload_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Fatalf("expected PUT, got %s", r.Method)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	m, err := newManagementClient(srv.URL)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := m.Reload(context.Background()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestManagementAPI_Reload_NotOK(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusMethodNotAllowed)
	}))
	defer srv.Close()

	m, err := newManagementClient(srv.URL)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := m.Reload(context.Background()); err == nil {
		t.Fatal("expected error for non-200 status")
	}
}

func TestManagementAPI_Quit_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Fatalf("expected PUT, got %s", r.Method)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	m, err := newManagementClient(srv.URL)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := m.Quit(context.Background()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestManagementAPI_Quit_NotOK(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusMethodNotAllowed)
	}))
	defer srv.Close()

	m, err := newManagementClient(srv.URL)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := m.Quit(context.Background()); err == nil {
		t.Fatal("expected error for non-200 status")
	}
}

func newManagementClient(addr string) (ManagementAPI, error) {
	cli, err := New(addr, http.DefaultClient, nil)
	if err != nil {
		return nil, err
	}
	return cli, nil
}
