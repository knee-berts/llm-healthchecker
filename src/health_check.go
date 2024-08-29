package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
)

func main() {
	metricsEndpoint := os.Getenv("METRICS_ENDPOINT")
	if metricsEndpoint == "" {
		metricsEndpoint = "http://localhost:8080/metrics" // Fallback to default
	}

	queueDepthThresholdStr := os.Getenv("QUEUE_DEPTH_THRESHOLD")
	queueDepthThreshold, err := strconv.Atoi(queueDepthThresholdStr)
	if err != nil || queueDepthThreshold == 0 {
		queueDepthThreshold = 10 // Fallback to default
	}
	appPortStr := os.Getenv("APP_PORT")
	appPort, err := strconv.Atoi(appPortStr)
	if err != nil || appPort == 0 {
		appPort = 8081 // Fallback to default
	}

	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		healthCheckHandler(w, r, metricsEndpoint, queueDepthThreshold)
	})
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", appPort), nil))
}

func healthCheckHandler(w http.ResponseWriter, r *http.Request, metricsEndpoint string, queueDepthThreshold int) {
	resp, err := http.Get(metricsEndpoint)
	if err != nil {
		log.Println("Error fetching metrics:", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Println("Error reading metrics:", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Use the extractTGIQueueSize function
	queueDepth, err := extractTGIQueueSize(string(body))
	if err != nil {
		log.Println("Error extracting queue depth:", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Compare against the configured threshold
	if queueDepth <= queueDepthThreshold {
		fmt.Fprintln(w, "Healthy")
	} else {
		http.Error(w, "Unhealthy", http.StatusServiceUnavailable)
	}
}

func extractTGIQueueSize(metricsResponse string) (int, error) {
	for _, line := range strings.Split(metricsResponse, "\n") {
		line = strings.TrimSpace(line)                                   // Trim leading/trailing whitespace
		if strings.HasPrefix(strings.ToLower(line), "tgi_queue_size ") { // Case-insensitive comparison
			parts := strings.Fields(line)
			if len(parts) != 2 {
				return 0, fmt.Errorf("invalid tgi_queue_size line format: %s", line)
			}
			return strconv.Atoi(parts[1])
		}
	}
	return 0, fmt.Errorf("tgi_queue_size not found in metrics")
}
