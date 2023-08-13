#!/bin/bash

claimFile="$1"
executedBy="$2"
partnerName="$3"
endpoint="http://localhost:8080"

curl -X POST $endpoint \
    -H "Content-Type: multipart/form-data" \
    -F "claimFile=@$claimFile" \
    -F "executed_by=$executedBy" \
    -F "partner_name=$partnerName"