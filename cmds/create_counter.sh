#!/bin/bash

RES=`./sendmsg.sh "/new {\"type\": \"counter\"}"`
KEY=`echo ${RES[@]:4} | jq -r ".key"`

echo "${KEY}"
