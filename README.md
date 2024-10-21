https://github.com/open-telemetry/opentelemetry-go-contrib/tree/main/examples/otel-collector

# Prerequisites

Ensure you have [Docker Compose V2](https://docs.docker.com/compose/) installed before proceeding with this demo.

# Deploying with Docker Compose

Use the following command to start up the OpenTelemetry Collector, Jaeger, and Prometheus, while exposing the necessary ports for accessing the data:

```bash
docker compose up -d
```

# Running the Application

The complete code for this example is located in the [main.go](./main.go) file. To run it, make sure you have a recent version of Go (preferably >= 1.13), then execute:

```bash
go run main.go
```

This example simulates an application performing tasks for ten seconds before finishing.

# Viewing Instrumentation Data

Now for the exciting part—let's explore the telemetry data produced by the sample application!

## Jaeger UI

To view the generated traces, open your web browser and go to [http://localhost:16686](http://localhost:16686).

## Prometheus

To view the generated metrics, such as `testapp_run_total`, open your browser and navigate to [http://localhost:9090](http://localhost:9090).

# Shutting Down

To stop and clean up the example, run:

```bash
docker compose down
```

# Deploy collector

https://opentelemetry.io/docs/collector/installation/#manual-linux-installation