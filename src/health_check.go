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

	metricThresholdStr := os.Getenv("METRIC_THRESHOLD")
	metricThreshold, err := strconv.Atoi(metricThresholdStr)
	if err != nil || metricThreshold == 0 {
		metricThreshold = 15 // Fallback to default
	}

	appPortStr := os.Getenv("APP_PORT")
	appPort, err := strconv.Atoi(appPortStr)
	if err != nil || appPort == 0 {
		appPort = 8081 // Fallback to default
	}

	metricToCheck := os.Getenv("METRIC_TO_CHECK")
	if metricToCheck == "" {
		metricToCheck = "tgi_queue_size" // Fallback to default
	}

	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		healthCheckHandler(w, r, metricsEndpoint, metricThreshold, metricToCheck)
	})
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", appPort), nil))
}

func healthCheckHandler(w http.ResponseWriter, r *http.Request, metricsEndpoint string, metricThreshold int, metricToCheck string) {
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

	// Use the extractMetricValue function
	metricValue, err := extractMetricValue(string(body), metricToCheck)
	if err != nil {
		log.Println("Error extracting metric value:", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Handle null or empty response
	if len(body) == 0 || string(body) == "null" {
		log.Println("Metrics endpoint returned null or empty response. Assuming healthy for now.")
		fmt.Fprintln(w, "Healthy")
		return
	}

	// Log the metric value and the threshold
	log.Printf("Health check: %s = %d (threshold = %d)\n", metricToCheck, metricValue, metricThreshold)

	if metricValue <= metricThreshold {
		log.Println("Health check: Healthy")
		fmt.Fprintln(w, "Healthy")
	} else {
		log.Println("Health check: Unhealthy")
		http.Error(w, "Unhealthy", http.StatusServiceUnavailable)
	}
}

func extractMetricValue(metricsResponse, metricName string) (int, error) {
	// Handle null or empty response
	if len(metricsResponse) == 0 || metricsResponse == "null" {
		return 0, nil
	}

	for _, line := range strings.Split(metricsResponse, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(strings.ToLower(line), strings.ToLower(metricName)+" ") {
			parts := strings.Fields(line)
			if len(parts) != 2 {
				return 0, fmt.Errorf("invalid %s line format: %s", metricName, line)
			}
			return strconv.Atoi(parts[1])
		}
	}
	return 0, fmt.Errorf("%s not found in metrics", metricName)
}
