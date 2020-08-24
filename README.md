# gomparator

gomparator compares HTTP GET JSON responses from different hosts by checking if they respond with the same json (deep equal ignoring order) and status code.

## Download and install

    go get -u github.com/emacampolo/gomparator

## Create a file with relative URL

    eg:

    /v1/payment_methods?client.id=1
    /v1/payment_methods?client.id=2
    /v1/payment_methods?client.id=3

## Run

```sh
$ gomparator -path "/path/to/file/with/urls" --host "http://host1.com" --host "http://host2.com" -H "X-Auth-Token: abc"
```
![](example.gif)

By Default, it will use 1 worker, and the rate limit will be 5 req/s

## Options

#### `--help, -h`
Shows a list of commands or help for one command

#### `--version, -v`
Print the version

#### `--path value`
Specifies the file from which to read targets. It should contain one column only with a rel path. eg: /v1/cards?query=123

#### `--host value`
Targeted hosts. Exactly 2 hosts must be specified. eg: --host 'http://host1.com --host 'http://host2.com'

#### `--header value, -H value`
Headers to be used in the http call

#### `--ratelimit value, -r value`
Operation rate limit per second (default: 67 - 4 krpm aprox.)

#### `--workers value, -w value`
Number of workers running concurrently (default: 1000)

#### `--status-code-only`
Whether or not it only compares status code ignoring response body

#### `--timeout value`
Request timeout (default: 30s)

#### `--duration value, -d value`
Duration of the comparison [0 = forever] (default: 0s)

#### `--exclude value`
Excludes a value from both json for the specified path. A [path](#path-syntax) is a series of keys separated by a dot or #. Default results.#.payer_costs.#.payment_method_option_id

#### `--parameter cutting`
Separate in files by the parameters parametrized. Default caller_id,display_filtered,differential_pricing_id,bins,public_key

## Path syntax

Given the following json input:

```json
{
  "name": {"first": "Tom", "last": "Anderson"},
  "friends": [
	{"first": "James", "last": "Murphy"},
	{"first": "Roger", "last": "Craig"}
  ]
}
```

<table>
<thead><tr><th>path</th><th>output</th></tr></thead>
<tbody>
<tr><td><b>"name.friends"</b></td><td>

```json

{
    "name": {"first": "Tom", "last": "Anderson"}
}

```
</td></tr>
<tr><td><b>"friends.#.last"</b></td><td>

```json

{
    "name": {"first": "Tom", "last": "Anderson"},
    "friends": [
        {"first": "James"},
        {"first": "Roger"}
    ]
}

```

</td></tr>

</tbody></table>

#### `Run consecutive comparisons`
Create a sh file with the next code and change vars values. ;) 

```console
#!/bin/bash

# sysinfo_page - A script to run all comparators

##### Constants

AUTH_TOKEN="8b2e19e9f224ce2fab8a682672e072fdef5c3aa0ea9752a680cf23529ef2b293"
SCOPE_1="https://read-batch_payment-methods.furyapps.io"
SCOPE_2="https://production-reader-testscope_payment-methods-read-v2.furyapps.io"
ARRAY_PATHS=(
	"/PATH/MELI/MLM/MLM.csv"
	"/PATH/MELI/MCO/MCO.csv"
	"/PATH/NONE/MLM/MLM.csv"
	"/PATH/NONE/MCO/MCO.csv"
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

```

#### `Terminal summary according to requested cuts`

![](termina-summary.png)