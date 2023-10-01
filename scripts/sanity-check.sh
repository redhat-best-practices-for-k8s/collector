#!/bin/bash

# Check if the required command 'jq' is installed
if ! command -v jq &>/dev/null; then
    echo "Error: 'jq' is not installed. Please install it to proceed."
    exit 1
fi

# Get results from collector
results=$(./scripts/get-from-collector.sh "admin" "adminpa55")
echo `\n results: \n $results`
results_test_ids=($(echo $results | jq -r '.[-1].ClaimResults[].test_id'))
echo `\n results test ids: \n $results_test_ids\n` 

# Get generated policy requiredPassingTests ids
GENERATED_POLICY_RAW_URL="https://raw.githubusercontent.com/test-network-function/cnf-certification-test/main/generated_policy.json"
read -d '' -ra required_test_ids <<< "$(curl $GENERATED_POLICY_RAW_URL | jq -r '.grades.requiredPassingTests[].id')"

# Ensure all ids in ids_array are in results from collector
echo "Iterating over required passing tests ids..."
for test_id in "${required_test_ids[@]}"; do
    if [[ "${results_test_ids[*]}" == *"$test_id"* ]]; then
        echo "test $test_id exists in the collector"
    else
        echo "test $test_id does not exist in the collector"
        echo "Collector's sanity check has failed."
        exit 1
    fi
done

echo "Collector's sanity check has passed."