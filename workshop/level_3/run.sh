#!/bin/bash

SLOWDOWN_IN_SECONDS=${SLOWDOWN_IN_SECONDS:-0.5}
URL=${URL:-"http://recipe-sample:8080/liveness"}

while true;
do
MAX=$(($(($RANDOM%30))+10))

for i in `seq 1 ${MAX}`
do
timeout 5 curl -s "${URL}" & 
done

wait
 
sleep ${SLOWDOWN_IN_SECONDS}
done