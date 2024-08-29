package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHealthCheckHandler(t *testing.T) {
	// Test with a healthy queue depth
	mockMetrics := `
    # HELP tgi_queue_size Current queue depth for inference requests.
    # TYPE tgi_queue_size gauge
    tgi_queue_size 5 
    # Other metrics...
    `
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, mockMetrics)
	}))
	defer server.Close()

	// Create a request to the /health endpoint
	req, err := http.NewRequest("GET", "/health", nil)
	if err != nil {
		t.Fatal(err)
	}

	// Record the response using a ResponseRecorder
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		healthCheckHandler(w, r, server.URL, 10) // Use the mock server URL and threshold
	})

	handler.ServeHTTP(rr, req)

	// Check the status code
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	// Check the response body    1.  stackoverflow.com stackoverflow.com
	expected := "Healthy\n"
	if rr.Body.String() != expected {
		t.Errorf("handler returned unexpected body: got %v want %v",
			rr.Body.String(), expected)
	}

	// Test with an unhealthy queue depth
	mockMetrics = `
    # HELP tgi_queue_size Current queue depth for inference requests.
    # TYPE tgi_queue_size gauge
    tgi_queue_size 15 
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
		healthCheckHandler(w, r, server.URL, 10)
	})
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusServiceUnavailable {
		t.Errorf("handler returned wrong status code for unhealthy queue: got %v want %v",
			status, http.StatusServiceUnavailable)
	}

	expected = "Unhealthy\n"
	if rr.Body.String() != expected {
		t.Errorf("handler returned unexpected body for unhealthy queue: got %v want %v",
			rr.Body.String(), expected)
	}

	// Test with null metrics response
	server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK) // Set the status code to 200 OK (empty body)
	}))
	defer server.Close()

	req, err = http.NewRequest("GET", "/health", nil)
	if err != nil {
		t.Fatal(err)
	}
	rr = httptest.NewRecorder()
	handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		healthCheckHandler(w, r, server.URL, 10)
	})
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code for null metrics: got %v want %v",
			status, http.StatusOK)
	}

	expected = "Healthy\n" // Adjust if you have different behavior for null metrics
	if rr.Body.String() != expected {
		t.Errorf("handler returned unexpected body for null metrics: got %v want %v",
			rr.Body.String(), expected)
	}
}
