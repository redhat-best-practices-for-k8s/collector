#!/bin/bash

endpoint="$1"
partnerName="$2"
decodedPassword="$3"

curl "$endpoint?partner_name=$partnerName&decoded_password=$decodedPassword"