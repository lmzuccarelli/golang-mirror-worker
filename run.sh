#!/bin/bash

set -euxo pipefail 

PROJECT=golang-mirror-worker
LOG_LEVEL=trace
SERVER_PORT=9900
CALLBACK_URL=http://192.168.0.19:9000/api/v1/echo

export LOG_LEVEL SERVER_PORT CALLBACK_URL PROJECT

if [ "$#" -ne 1 ];
then
  echo -e "usage run.sh <start|stop>"
  exit 1
fi

case ${1} in
  start)
    echo -e "version v1.0.0"
    echo -e "project $PROJECT"
    rm -rf logs/console.log
    ./mirror-worker 1> logs/console.log 2>&1 & 
    exit 0
  ;;
  stop)
    rm -rf semaphore.txt
    ID=$(ps -ef | grep -v 'grep' | grep mirror-worker | awk '{print $2}')
    echo -e "stopping service with PID $ID"
    kill -9 $ID
    exit 0
  ;;
esac

