#!/bin/bash


RES=`echo "/new {\"type\": \"gset\"}" | netcat localhost 8079`

KEY=`echo ${RES[@]:4} | jq -r ".key"`

echo "Created gset with key \"${KEY}\""

sleep 0.5

echo "add 5 to 8079"
echo `echo "/add {\"key\": \"${KEY}\", \"value\": 5}" | netcat localhost 8079`
./cmds/connect.sh localhost 8079
echo "Connected 8089 with 8079"

sleep 0.1
echo "add 1 to 8089"
echo `./sendmsg.sh "/add {\"key\": \"${KEY}\", \"value\": 1}"`


sleep 0.5

echo
echo "FROM 8089"
echo `./sendmsg.sh "/get {\"key\": \"${KEY}\"}"`
echo "FROM 8079"
echo "/get {\"key\": \"${KEY}\"}" | netcat localhost 8079
