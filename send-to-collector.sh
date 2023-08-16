#!/bin/bash

claimFile="$1"
executedBy="$2"
partnerName="$3"
encoded_password="$4"
endpoint="http://localhost:8080"

curl -X POST $endpoint \
    -H "Content-Type: multipart/form-data" \
    -F "claimFile=@$claimFile" \
    -F "executed_by=$executedBy" \
<<<<<<< HEAD
    -F "partner_name=$partnerName"
    -F "encoded_password=$encoded_password"
=======
    -F "partner_name=$partnerName" \
    -F "decoded_password=$decoded_password"
>>>>>>> 68b4f47 (modifeid send-to-collector.sh)
