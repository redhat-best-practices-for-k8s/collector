#!/bin/bash

endpoint="$1"
partnerName="$2"
decodedPassword="$3"

if [ -z "$endpoint" ] || [ -z "$partnerName" ] || [ -z "$decodedPassword" ]; then
	echo "Usage: $0 [ endpoint ] [ partner_name ] [ password ]"
	exit 1
fi

curl "$endpoint?partner_name=$partnerName&decoded_password=$decodedPassword"
