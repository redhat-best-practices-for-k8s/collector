#!/bin/bash

endpoint="$1"
partnerName="$2"
decoded_password="$3"

curl "$endpoint?partner_name=$partnerName&decoded_password=$decoded_password"