#!/bin/bash


KEY=`./cmds/create_counter.sh`
echo "Created counter with key \"${KEY}\""

./cmds/connect.sh localhost 8079
echo "Connected 8089 with 8079"


echo `./sendmsg.sh "/inc {\"key\": \"${KEY}\", \"value\": 5}"`

sleep 1

echo
echo "FROM 8089"
echo `./sendmsg.sh "/get {\"key\": \"${KEY}\"}"`
echo "FROM 8079"
echo "/get {\"key\": \"${KEY}\"}" | netcat localhost 8079
