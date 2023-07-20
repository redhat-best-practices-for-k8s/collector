#!/bin/bash

claimFile="$1"
createdBy="$2"
partnerName="$3"
endpoint="http://localhost:8080"

curl -X POST $endpoint \
    -H "Content-Type: multipart/form-data" \
    -F "claimFile=@$claimFile" \
    -F "created_by=$createdBy" \
    -F "partner_name=$partnerName"