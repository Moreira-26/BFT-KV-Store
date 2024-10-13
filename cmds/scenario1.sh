#!/bin/bash

./cmds/connect.sh localhost 8079
echo "Connected 8089 with 8079"

KEY=`./cmds/create_counter.sh`
echo "Created counter with key \"${KEY}\""

echo
echo "FROM 8089"
echo `./sendmsg.sh "/get {\"key\": \"${KEY}\"}"`
echo "FROM 8079"
echo "/get {\"key\": \"${KEY}\"}" | netcat localhost 8079
