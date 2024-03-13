#!/bin/bash

# Check if the required command 'jq' is installed
if ! command -v jq &>/dev/null; then
	echo "Error: 'jq' is not installed. Please install it to proceed."
	exit 1
fi

# Check if Environment variables exists
if [ -z "$ENDPOINT" ]; then
	echo "Error: Collector's endpoint must be specified."
	exit 1
fi

if [ -z "$COLLECTOR_USERNAME" ] || [ -z "$COLLECTOR_PASSWORD" ]; then
	echo "Error: COLLECTOR_USERNAME and COLLECTOR_PASSWORD env variables must be set."
	exit 1
fi

# Get results from collector
results=$(./scripts/get-from-collector.sh "$ENDPOINT" "$COLLECTOR_USERNAME" "$COLLECTOR_PASSWORD")
start_index=${results%%[*} start_index=$((${#start_index} + 1))
results="${results:start_index-1}"
results_claim=$(echo "$results" | jq -r '.[-1].Claim')

if [ -z "$results_claim" ]; then
	echo "User ${COLLECTOR_USERNAME} doesn't have any claims stored."
	echo "Collector's sanity check has failed!."
	exit 1
fi

echo "Collector's sanity check has passed."
