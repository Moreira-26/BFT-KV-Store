#!/bin/bash

PORT="$2"
if [ -n PORT ]; then 
	PORT="8089"
fi

echo $1 | netcat localhost $PORT

printf "\n"
