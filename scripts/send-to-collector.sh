#!/bin/bash

endpoint="$1"
claimFile="$2"
executedBy="$3"
partnerName="$4"
decodedPassword="$5"

curl -X POST $endpoint \
    -H "Content-Type: multipart/form-data" \
    -F "claimFile=@$claimFile" \
    -F "executed_by=$executedBy" \
    -F "partner_name=$partnerName" \
    -F "decoded_password=$decodedPassword"
