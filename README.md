# LLM Healthchecker - Simulating Inference Server Capacity
This Go application serves as a health checker for LLM (Large Language Model) inference servers, with a specific focus on simulating capacity overload. It achieves this by monitoring the queue depth of incoming requests on the inference server.  When the queue depth exceeds a predefined threshold, the health checker signals an "unhealthy" status, triggering Kubernetes to remove the pod from the service's endpoints. This simulation helps demonstrate how the GKE multi-cluster gateway can intelligently route traffic to inference servers in multiple regions, ensuring optimal performance and high availability even under heavy load.

## Context

- **Hugging Face's efficiency:** Hugging Face's inference servers are highly efficient, continually accepting requests and queuing them locally for processing.
- **Capacity simulation:** This health checker allows us to artificially create a "full" scenario by monitoring the queue depth. If the queue grows beyond a certain limit, we consider the pod "full" and trigger its removal from the service endpoints.
- **Multi-cluster gateway demonstration:** This behavior showcases how the GKE multi-cluster gateway can seamlessly redirect traffic to inference servers in other regions when one region's capacity is saturated.

## Features
- **Flexible metric checking:** You can configure the metric to monitor and its threshold using environment variables.
- **Handles null metrics:** Gracefully handles cases where the metrics endpoint returns null or an empty response, indicating that the LLM service hasn't processed any requests yet.
- **Exposes a `/health` endpoint:** Provides a `/health` endpoint that Kubernetes or other systems can use for readiness probes.

## How to Use
1. Build the application:

```bash
go build health_check.go
```

2. Set environment variables:

`METRICS_ENDPOINT`: The URL of your LLM inference server's metrics endpoint (e.g., `http://localhost:8000/metrics`).
`METRIC_THRESHOLD`: The threshold value for the metric you want to check.
`METRIC_TO_CHECK`: The name of the metric to monitor (e.g., `tgi_queue_size`).
`APP_PORT` (optional): The port on which the health checker should listen (defaults to `8081`).

3. Run the application:

```bash
./health_check
```

4. Access the health check endpoint:

You can now access the health check status by making a GET request to `http://localhost:8081/health` (or the port you configured).

- If the metric value is within the threshold, the response will be `200 OK` with the body "Healthy".
- If the metric value exceeds the threshold, the response will be `503 Service Unavailable` with the body "Unhealthy".
- If the metrics endpoint returns null or an empty response, the response will be `200 OK` with the body "Healthy" (you can customize this behavior in the code).

## Example Usage in Kubernetes
Deploy this health checker as a sidecar container alongside your Hugging Face inference server. Configure the `inference-server` container's readiness probe to target the health checker's `/health` endpoint.

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: tgi-gemma-deployment
  namespace: inference
  labels: 
    app: gemma-server
spec:
  replicas: 1
  selector:
    matchLabels:
      app: gemma-server
  template:
    metadata:
      labels:
        app: gemma-server
        ai.gke.io/model: gemma-2b-1.1-it
        ai.gke.io/inference-server: text-generation-inference
        examples.ai.gke.io/source: user-guide
    spec:
      containers:
      - name: llm-healthcheck 
        image: us-docker.pkg.dev/fleet-dev-1/llm-healthcheck/llm-healthcheck-v0.0.9
        resources:
          requests:
            cpu: "100m"
            memory: "128Mi"
        env:
        - name: METRICS_ENDPOINT 
          value: "http://localhost:8080/metrics" 
        - name: METRICS_THRESHOLD
          value: "10"
        - name: METRIC_TO_CHECK 
          value: "tgi_queue_size" 
        - name: APP_PORT
          value: "8081" 
        ports:
        - name: healthcheck
          containerPort: 8081
      - name: inference-server
        image: us-docker.pkg.dev/vertex-ai/vertex-vision-model-garden-dockers/pytorch-hf-tgi-serve:20240328_0936_RC01
        readinessProbe:
          httpGet:
            path: /health
            port: 8081 # Port the llm-healthchecker is listening on
          initialDelaySeconds: 20 # Delay before first probe
          periodSeconds: 10           
        resources:
          requests:
            cpu: "2"
            memory: "7Gi"
            ephemeral-storage: "20Gi"
            nvidia.com/gpu: 1
          limits:
            cpu: "2"
            memory: "7Gi"
            ephemeral-storage: "20Gi"
            nvidia.com/gpu: 1
        args:
        - --model-id=$(MODEL_ID)
        - --num-shard=1
        env:
        - name: MODEL_ID
          value: google/gemma-1.1-2b-it
        - name: PORT
          value: "8080"
        - name: HUGGING_FACE_HUB_TOKEN
          valueFrom:
            secretKeyRef:
              name: hf-secret
              key: hf_api_token
        volumeMounts:
        - mountPath: /dev/shm
          name: dshm
        ports:
        - name: web
          containerPort: 8080
      volumes:
      - name: dshm
        emptyDir:
          medium: Memory
      nodeSelector:
        cloud.google.com/gke-accelerator: nvidia-tesla-a100
```

In this configuration, the `inference-server` container's readiness probe will check the `/health` endpoint of the `llm-healthchecker` container to determine its readiness.

## Important Notes
- In a real-world production scenario, you'd likely implement this queue depth monitoring directly within the inference server's code and expose a native health metric for Kubernetes to check
- This implementation focuses on simulating capacity issues for demonstration purposes
- Full Multi-Cluster Gateway demo is coming soon!

## License
MIT