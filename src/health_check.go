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

// func keepAlive() {
// 	// Get the app port from APP_PORT
// 	appPortStr := os.Getenv("APP_PORT")
// 	appPort, err := strconv.Atoi(appPortStr)
// 	if err != nil || appPort == 0 {
// 		appPort = 8080 // Fallback if not set or invalid
// 	}

// 	// Construct the inference server URL using appPort
// 	inferenceServerURL := fmt.Sprintf("http://localhost:%d", appPort)

// 	// Add "/generate" to construct the full URL
// 	generateURL := inferenceServerURL + "/generate"

// 	for {
// 		// Construct the JSON payload
// 		payload := map[string]interface{}{
// 			"inputs": "What is a keep alive?",
// 			"parameters": map[string]interface{}{
// 				"max_new_tokens": 100,
// 			},
// 		}
// 		jsonPayload, err := json.Marshal(payload)
// 		if err != nil {
// 			log.Println("Error marshalling JSON:", err)
// 			return
// 		}

// 		// Make the HTTP POST request
// 		resp, err := http.Post(generateURL, "application/json", bytes.NewBuffer(jsonPayload))
// 		if err != nil {
// 			log.Println("Error making keep-alive request:", err)
// 			return
// 		}
// 		defer resp.Body.Close()

// 		// Read and log the response (optional)
// 		body, err := io.ReadAll(resp.Body)
// 		if err != nil {
// 			log.Println("Error reading keep-alive response:", err)
// 			return
// 		}
// 		log.Println("Keep-alive response:", string(body))

// 		// Sleep for 30 minutes
// 		time.Sleep(30 * time.Minute)
// 	}
// }

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

	// go keepAlive()

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

	// Handle null or empty response
	if len(body) == 0 || string(body) == "null" {
		log.Println("Metrics endpoint returned null or empty response. Assuming healthy for now.")
		fmt.Fprintln(w, "Healthy") // Or you might choose to return a different status/message
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
	// Handle null or empty response directly in this function
	if len(metricsResponse) == 0 || metricsResponse == "null" {
		return 0, nil // Or any default value you prefer for queue size when metrics are null
	}

	for _, line := range strings.Split(metricsResponse, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(strings.ToLower(line), "tgi_queue_size ") {
			parts := strings.Fields(line)
			if len(parts) != 2 {
				return 0, fmt.Errorf("invalid tgi_queue_size line format: %s", line)
			}
			return strconv.Atoi(parts[1])
		}
	}
	return 0, fmt.Errorf("tgi_queue_size not found in metrics")
}
