#!/bin/bash

cd ./modbus-slave || exit
source .venv/bin/activate
python main.py &
sleep 0.5