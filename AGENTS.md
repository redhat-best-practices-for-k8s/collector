# AGENTS.md

This file provides guidance to AI coding agents when working with code in this repository.

## Project Overview

The Collector is a Go-based backend service for collecting and storing CNF (Cloud Native Functions) certification test results from the [CNF Certification Suite](https://github.com/redhat-best-practices-for-k8s/certsuite). It provides:

- HTTP API for uploading claim.json files (POST) and retrieving stored results (GET)
- MySQL database for storing claim metadata and test results
- AWS S3 storage for archiving claim files
- Partner-based authentication with bcrypt password hashing
- Multi-tenant data access (partners see only their data; admin sees all)

## Commands

### Build and Test
```bash
make build              # Build the collector binary
make test               # Run unit tests with coverage (generates cover.out.tmp)
```

### Linting
```bash
make lint               # Run all linters (golangci-lint, hadolint, shfmt, typos, markdownlint, yamllint, shellcheck, checkmake)
make tool-precheck      # Check that all linting tools are installed
make install-mac-brew-tools  # Install linting tools on macOS via Homebrew
```

### Container Images
```bash
make build-image-collector   # Build Docker image
make push-image-collector    # Push image to registry
make run-collector           # Run collector locally in Docker (requires tnf-secrets)
make stop-running-collector-container  # Stop running container
```

### Database Setup
```bash
make run-initial-mysql-scripts      # Initialize local MySQL (requires tnf-secrets)
make run-initial-mysql-scripts-rds  # Initialize AWS RDS MySQL
```

### Client Scripts
```bash
# Send a claim file to the collector
./scripts/send-to-collector.sh "endpoint" "path/to/claim.json" "executed_by" "partner_name" "password"

# Retrieve stored claims
./scripts/get-from-collector.sh "endpoint" "partner_name" "password"
```

## Architecture

### Request Flow

**POST (Upload Claim):**
1. Parse multipart form (claim file, executed_by, partner_name, password)
2. Validate claim JSON structure
3. Authenticate/create partner credentials
4. Upload claim file to S3
5. Store claim metadata in `claim` table
6. Store test results in `claim_result` table

**GET (Retrieve Claims):**
1. Authenticate partner credentials
2. Query claims (filtered by partner, or all for admin)
3. Query related claim results
4. Return combined JSON response

### Package Structure

```
main.go           # Entry point: initializes storage and starts HTTP server
api/
  server.go       # HTTP server setup, routes POST→ParserHandler, GET→ResultsHandler
  parser_handler.go    # POST request handling (upload/store claims)
  result_handler.go    # GET request handling (retrieve claims)
  auth.go         # bcrypt password hashing and verification
  validator.go    # Request and claim JSON validation
  parser.go       # Claim file JSON parsing
  upload_handler.go    # S3 upload/delete operations
storage/
  mysql.go        # MySQL database connection
  s3.go           # S3 storage client initialization
types/
  types.go        # Data models: Claim, ClaimResult, ClaimCollector
util/
  constants.go    # SQL queries, error messages, form field names
  http_utils.go   # Environment variable handling, HTTP utilities
  db_utils.go     # Database transaction management
```

### Database Schema

Database: `cnf`

**Tables:**
- `claim` - Claim metadata (id, cnf_version, executed_by, upload_time, partner_name, s3_file_key)
- `claim_result` - Individual test results (id, claim_id, suite_name, test_id, test_status)
- `authenticator` - Partner credentials (partner_name, encoded_password)

### Environment Variables

**Server:**
- `SERVER_ADDR` - Listen address (e.g., `:80`)
- `SERVER_READ_TIMEOUT` - Read timeout in seconds
- `SERVER_WRITE_TIMEOUT` - Write timeout in seconds

**Database:**
- `DB_USER`, `DB_PASSWORD`, `DB_URL`, `DB_PORT`

**S3:**
- `AWS_ACCESS_KEY`, `AWS_SECRET_ACCESS_KEY`
- `S3_BUCKET_NAME`, `S3_BUCKET_REGION`

### Secrets Management

Production secrets are stored in a separate private repository (`tnf-secrets`). The Makefile clones this repository when needed for deployment tasks.

## Go Version

This project uses Go 1.25.5.
