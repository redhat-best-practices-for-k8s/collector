#!/bin/bash

endpoint="$1"
claimFile="$2"
executedBy="$3"
partnerName="$4"
decodedPassword="$5"

if [ -z "$endpoint" ] || [ -z "$claimFile" ] || [ -z "$executedBy"]; then
    echo "Usage: $0 [ endpoint ] [ path/to/claim.json ] [ executed_by ] [ partner_name(optional ] [ password(optinal) ]"
    exit 1
fi

curl -X POST $endpoint \
    -H "Content-Type: multipart/form-data" \
    -F "claimFile=@$claimFile" \
    -F "executed_by=$executedBy" \
    -F "partner_name=$partnerName" \
    -F "decoded_password=$decodedPassword"
