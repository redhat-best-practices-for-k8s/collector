# collector
A Go-based endpoint for collecting TNF logs

# Goals for this project

- Go HTTP web backend application that will accept incoming requests from TNF runs.
- Parse the incoming requests (JSON payloads).
- Place the results into a database to aggregate results into statistics.

# Instructions for Running Locally

Use the following `Make` commands to build the collector container locally:

### Prerequisites
- Docker or Podman 

### Commands
Build the image:
- `make build-image-local` 

Run the application via container:
- `make run-collector` or `make run-collector-headless`

Cleanup after:
- `make stop-collector`


