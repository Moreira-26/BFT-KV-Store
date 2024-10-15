#!/bin/bash


python3 ./cmds/sendcmd.py 'localhost' '8089' 'CONN' '{"Address": "localhost", "Port": "8079"}'
echo "Connected 8089 with 8079"

KEY=`python3 ./cmds/sendcmd.py 'localhost' '8089' '/new' '{"type": "counter"}'`
KEY=`echo ${KEY[@]:4} | jq -r ".key"`
echo "Created counter with key \"${KEY}\""

echo
echo "FROM 8089"
echo `python3 ./cmds/sendcmd.py "localhost" "8089" "/get" "{\"key\": \"${KEY}\"}"`
echo "FROM 8079"
echo `python3 ./cmds/sendcmd.py "localhost" "8079" "/get" "{\"key\": \"${KEY}\"}"`
