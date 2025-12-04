#!/bin/bash

PID=$(pgrep -f 'python main.py')
kill "$PID"