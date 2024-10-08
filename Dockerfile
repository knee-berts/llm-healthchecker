# Build stage
FROM golang:1.22-alpine AS build

WORKDIR /app

COPY ./src ./

RUN go build -v -o health_check

# Runtime stage
FROM alpine:latest

WORKDIR /app

COPY --from=build /app/health_check .

# Set default values for environment variables
# ENV METRICS_ENDPOINT="http://localhost:8080/metrics"
# ENV METRIC_THRESHOLD=15
# ENV APP_PORT=8081
# ENV METRIC_TO_CHECK="tgi_queue_size"

# Expose the port the application will listen on
EXPOSE $APP_PORT

# Command to run the application with environment variables
CMD ["./health_check"]