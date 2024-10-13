#!/bin/bash

RES=`./sendmsg.sh "/new {\"type\": \"gset\"}"`
KEY=`echo ${RES[@]:4} | jq -r ".key"`

echo "New GSet created with id: ${KEY}"
echo

for var in "$@"
do
	if [ `./sendmsg.sh "/add {\"key\": \"${KEY}\", \"value\": \"${var}\"}"` = "R_OK" ]; then
		echo "${var} added to the set"
	else
		echo "Failed to add ${var}"
	fi
done

echo 
VALUE_GET=`./sendmsg.sh "/get {\"key\": \"${KEY}\"}"`
VALUE_READ=`echo ${VALUE_GET[@]:4} | jq`

echo "Response: ${VALUE_READ}"
