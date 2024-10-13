#!/bin/bash

RES=`./sendmsg.sh "/new {\"type\": \"gset\"}"`
KEY=`echo ${RES[@]:4} | jq -r ".key"`

echo "${KEY}"
