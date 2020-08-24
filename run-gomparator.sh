#!/bin/bash

# sysinfo_page - A script to run all comparators

##### Constants

AUTH_TOKEN="b1911bf99c5c32b0c151095a9f956ece09801dc3eaaa966a0e2911e23df7c4a8"
SCOPE_1="https://read-batch_payment-methods.furyapps.io"
SCOPE_2="https://production-reader-testscope_payment-methods-read-v2.furyapps.io"
ARRAY_PATHS=(
	"/Users/mpons/Documents/comparator/payment-methods/v2/17-08-2020_21-08-2020/202008-10-15/MELI/MLM/MLM.csv"
	"/Users/mpons/Documents/comparator/payment-methods/v2/17-08-2020_21-08-2020/202008-10-15/MELI/MCO/MCO.csv"
	"/Users/mpons/Documents/comparator/payment-methods/v2/17-08-2020_21-08-2020/202008-10-15/NONE/MLM/MLM.csv"
	"/Users/mpons/Documents/comparator/payment-methods/v2/17-08-2020_21-08-2020/202008-10-15/NONE/MCO/MCO.csv"
	) 

ARRAY_CHANNELS=("point" "splitter" "instore")

for i in "${ARRAY_PATHS[@]}"
do
	gomparator -path "$i" -host "${SCOPE_1}" -host "${SCOPE_2}" -header "X-Auth-Token:${AUTH_TOKEN}"
done

for i in "${ARRAY_PATHS[@]}"
do
	for j in "${ARRAY_CHANNELS[@]}"
	do
		gomparator -path "$i" -host "${SCOPE_1}" -host "${SCOPE_2}" -header "X-Auth-Token:${AUTH_TOKEN}" -header "X-Caller-Scopes:$j"
	done
done