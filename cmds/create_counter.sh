#!/bin/bash

RES=`./cmds/sendcmd.py '/new' '{\"type\": \"counter\"}'`
KEY=`echo ${RES[@]:4} | jq -r ".key"`

echo "${KEY}"
