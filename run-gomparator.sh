#!/bin/bash

# sysinfo_page - A script to run all comparators

##### Constants

AUTH_TOKEN="f8f6450c0703b992c12ae4e923a669b9c168c519a0d6d6376bd0edf3b502778a"
SCOPE_1="https://read-batch_payment-methods.furyapps.io"
SCOPE_2="https://production-reader-testscope_payment-methods-read-v2.furyapps.io"
ARRAY_PATHS=(
	"/Users/mpons/Documents/comparator/payment-methods/v2/2_24-08-2020_28-08-2020/202008-10-15/MELI/MLM/MLM.error"
	"/Users/mpons/Documents/comparator/payment-methods/v2/2_24-08-2020_28-08-2020/202008-10-15/MELI/MCO/MCO.error"
	"/Users/mpons/Documents/comparator/payment-methods/v2/2_24-08-2020_28-08-2020/202008-10-15/NONE/MLM/MLM.error"
	"/Users/mpons/Documents/comparator/payment-methods/v2/2_24-08-2020_28-08-2020/202008-10-15/NONE/MCO/MCO.error"
	) 

#ARRAY_CHANNELS=("" "point" "splitter" "instore")
ARRAY_CHANNELS=("")

for i in "${ARRAY_PATHS[@]}"
do
	for j in "${ARRAY_CHANNELS[@]}"
	do
		gomparator -path "$i" -host "${SCOPE_1}" -host "${SCOPE_2}" -header "X-Auth-Token:${AUTH_TOKEN}" -header "X-Caller-Scopes:$j"
	done
done