package main

import (
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
	"net/url"
	"testing"
	"time"
)

func TestRateLimiter(t *testing.T) {
	rateLimiter := NewRateLimiter(1, 2)

	rateLimiter.GetLimiter("1.2.3.4").Allow()
	rateLimiter.GetLimiter("1.2.3.4").Allow()
	response := rateLimiter.GetLimiter("1.2.3.4").Allow()

	if response != false {
		t.Errorf("got %t for the third request, wanted %t", response, !response)
	}
}

func TestCircuitBreaker(t *testing.T) {
	cb := NewCircuitBreaker(2, 100*time.Millisecond)

	cb.RecordFailure()
	cb.RecordFailure()

	if isOpen := cb.IsOpen(); isOpen != true {
		t.Errorf("got %t for circuit breaker status when expected %t", isOpen, !isOpen)
	}

	time.Sleep(time.Until(cb.openUntil))

	if isOpen := cb.IsOpen(); isOpen != false {
		t.Errorf("got %t for circuit breaker status when expected %t", isOpen, !isOpen)
	}

	cb.RecordSuccess()

	if isOpen := cb.IsOpen(); isOpen != false {
		t.Errorf("got %t for circuit breaker status when expected %t", isOpen, !isOpen)
	}
}

func TestLoggingMiddleware(t *testing.T) {
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte("created"))
	})

	handler := loggingMiddleware(inner)

	req := httptest.NewRequest("POST", "/test", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Errorf("got status %d, want %d", rec.Code, http.StatusCreated)
	}
	if rec.Body.String() != "created" {
		t.Errorf("got body %q, want %q", rec.Body.String(), "created")
	}
}

func TestCircuitBreakerIntegration(t *testing.T) {
	backendHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadGateway)
	})

	srv := httptest.NewServer(backendHandler)
	defer srv.Close()

	backendURL, _ := url.Parse(srv.URL)
	proxy := httputil.NewSingleHostReverseProxy(backendURL)

	cb := NewCircuitBreaker(2, 100*time.Millisecond)

	proxy.ModifyResponse = func(r *http.Response) error {
		if r.StatusCode < 500 {
			cb.RecordSuccess()
		} else {
			cb.RecordFailure()
		}
		return nil
	}

	handler := gatewayHandler(proxy, cb)

	for i := 0; i < 2; i++ {
		req := httptest.NewRequest("GET", "/", nil)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusBadGateway {
			t.Errorf("request %d: got status %d, want 502", i+1, rec.Code)
		}
	}

	// Third request: circuit is open, should get 503 without hitting backend
	req := httptest.NewRequest("GET", "/", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusServiceUnavailable {
		t.Errorf("got status %d, want 503 (circuit open)", rec.Code)
	}
}
