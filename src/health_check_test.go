package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHealthCheckHandler(t *testing.T) {
	// Test with a healthy metric value
	mockMetrics := `
    # HELP some_metric Some metric we're monitoring
    # TYPE some_metric gauge
    some_metric 3 
    # Other metrics...
    `
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, mockMetrics)
	}))
	defer server.Close()

	req, err := http.NewRequest("GET", "/health", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		healthCheckHandler(w, r, server.URL, 5, "some_metric") // Use mock server, threshold 5, and metric name "some_metric"
	})

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	expected := "Healthy\n"
	if rr.Body.String() != expected {
		t.Errorf("handler returned unexpected body: got %v want %v", rr.Body.String(), expected)
	}

	// Test with an unhealthy metric value
	mockMetrics = `
    # HELP some_metric Some metric we're monitoring
    # TYPE some_metric gauge
    some_metric 7 
    # Other metrics...
    `
	server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, mockMetrics)
	}))
	defer server.Close()

	req, err = http.NewRequest("GET", "/health", nil)
	if err != nil {
		t.Fatal(err)
	}
	rr = httptest.NewRecorder()
	handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		healthCheckHandler(w, r, server.URL, 5, "some_metric")
	})
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusServiceUnavailable {
		t.Errorf("handler returned wrong status code for unhealthy metric: got %v want %v", status, http.StatusServiceUnavailable)
	}

	expected = "Unhealthy\n"
	if rr.Body.String() != expected {
		t.Errorf("handler returned unexpected body for unhealthy metric: got %v want %v", rr.Body.String(), expected)
	}

	// Test with null metrics response
	server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	req, err = http.NewRequest("GET", "/health", nil)
	if err != nil {
		t.Fatal(err)
	}
	rr = httptest.NewRecorder()
	handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		healthCheckHandler(w, r, server.URL, 5, "some_metric")
	})
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code for null metrics: got %v want %v", status, http.StatusOK)
	}

	expected = "Healthy\n"
	if rr.Body.String() != expected {
		t.Errorf("handler returned unexpected body for null metrics: got %v want %v", rr.Body.String(), expected)
	}
}
