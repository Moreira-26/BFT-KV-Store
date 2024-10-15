#!/bin/bash


KEY=`python3 ./cmds/sendcmd.py "localhost" "8079" "/new" "{\"type\": \"gset\"}"`
KEY=`echo ${KEY[@]:4} | jq -r ".key"`
echo "Created gset with key \"${KEY}\""

sleep 0.5

echo "add 5 to 8079"
echo `python3 ./cmds/sendcmd.py "localhost" "8079" "/add" "{\"key\": \"${KEY}\", \"value\": 5}"`
echo
./cmds/connect.sh localhost 8089 localhost 8079
echo "Connected 8089 with 8079"

sleep 0.1
echo "add 1 to 8089"
echo `python3 ./cmds/sendcmd.py "localhost" "8089" "/add" "{\"key\": \"${KEY}\", \"value\": 1}"`


sleep 0.5

echo
echo "FROM 8089"
echo `python3 ./cmds/sendcmd.py "localhost" "8089" "/get" "{\"key\": \"${KEY}\"}"`
echo "FROM 8079"
echo `python3 ./cmds/sendcmd.py "localhost" "8079" "/get" "{\"key\": \"${KEY}\"}"`
