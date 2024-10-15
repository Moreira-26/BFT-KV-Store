#!/bin/bash


KEY=`python3 ./cmds/sendcmd.py "localhost" "8089" "/new" "{\"type\": \"counter\"}"`
KEY=`echo ${KEY[@]:4} | jq -r ".key"`
echo "Created counter with key \"${KEY}\""

./cmds/connect.sh localhost 8089 localhost 8079
echo "Connected 8089 with 8079"


echo `python3 ./cmds/sendcmd.py "localhost" "8089" "/inc" "{\"key\": \"${KEY}\", \"value\": 5}"`

sleep 0.2

echo
echo "FROM 8089"
echo `python3 ./cmds/sendcmd.py "localhost" "8089" "/get" "{\"key\": \"${KEY}\"}"`
echo "FROM 8079"
echo `python3 ./cmds/sendcmd.py "localhost" "8079" "/get" "{\"key\": \"${KEY}\"}"`
