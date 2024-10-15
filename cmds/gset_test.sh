#!/bin/bash

KEY=`python3 ./cmds/sendcmd.py 'localhost' '8089' '/new' '{"type": "gset"}'`
KEY=`echo ${KEY[@]:4} | jq -r ".key"`

echo "New GSet created with id: ${KEY}"
echo

for var in "$@"
do
	if [ `python3 ./cmds/sendcmd.py "localhost" "8089" "/add" "{\"key\": \"${KEY}\", \"value\": \"${var}\"}"` = "R_OK" ]; then
		echo "${var} added to the set"
	else
		echo "Failed to add ${var}"
	fi
done

echo 

VALUE_GET=`python3 ./cmds/sendcmd.py "localhost" "8089" "/get" "{\"key\": \"${KEY}\"}"`
VALUE_READ=`echo ${VALUE_GET[@]:4} | jq`

echo "Response: ${VALUE_READ}"
