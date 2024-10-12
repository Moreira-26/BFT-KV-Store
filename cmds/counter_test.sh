#!/bin/bash

RES=`./sendmsg.sh "/new {\"type\": \"counter\"}"`
KEY=`echo ${RES[@]:4} | jq -r ".key"`

echo "New GSet created with id: ${KEY}"
echo

for var in "$@"
do
	if [ `./sendmsg.sh "/inc {\"key\": \"${KEY}\", \"value\": ${var}}"` = "R_OK" ]; then
		echo "${var} incremented"
	else
		echo "Failed to increment ${var}"
	fi
done

echo 
echo "Getting value of ${KEY}"
VALUE_GET=`./sendmsg.sh "/get {\"key\": \"${KEY}\"}"`
VALUE_READ=`echo ${VALUE_GET[@]:4} | jq`

echo "Response: ${VALUE_READ}"
