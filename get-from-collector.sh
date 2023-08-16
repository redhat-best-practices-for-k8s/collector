#!/bin/bash

partnerName="$1"
decoded_password="$2"
endpoint="http://localhost:8080"

curl "$endpoint?partner_name=$partnerName&decoded_password=$decoded_password"