#!/bin/bash


echo `python3 ./cmds/sendcmd.py "${1}" "${2}" "CONN" "{\"Address\": \"${3}\", \"Port\": \"${4}\"}"`
