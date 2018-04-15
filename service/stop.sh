#!/bin/bash

killables=$(ps aux | grep lake)

if [ ! "${killables}" == "" ] ; then
  echo "You are going to kill some process:"
  echo "${killables}"
else
  echo "No process with the pattern $1 found."
  exit 0
fi

for pid in $(echo "${killables}" | awk '{print $2}') ; do
  for signal in TERM TERM TERM KILL ; do
    echo "killing ${pid} with ${signal} ..."
    if ! pkill "-${signal}" -f -- "lake" ; then
      echo "${pid} killed"
      break
    fi
    sleep 1
  done
done
