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
- `make build-image-collector`

Run the application via container:
- `make run-collector` or `make run-collector-headless`

Cleanup after:
- `make stop-collector`

# Instructions for running send-to-collector.sh

From collector's repo root directory, use the following command:

`./scripts/send-to-collector.sh "enter_endpoint" "path/to/claim.json" "enter_executed_by" "enter_partner_name(optional)" "enter_password(optinal)"`

# Instructions for running get-from-collector.sh

From collector's repo root directory, use the following command:

`./scripts/get-from-collector.sh "enter_endpoint" "enter_partner_name" "enter_password"`

# Instructions for running sanity-check.sh

From collector's repo root directory, use the following command:

`ENDPOINT=enter_endpoint COLLECTOR_USERNAME=enter_partner_name COLLECTOR_PASSWORD=enter_password ./scripts/sanity-check.sh`


