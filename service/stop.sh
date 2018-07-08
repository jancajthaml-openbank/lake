#!/bin/bash

killables=$(ps -ef | awk '$8=="/openbank/services/lake/entrypoint" {print $2}')

if [[ -z "${killables// }" ]] ; then
  echo "Not running"
  exit 0
fi

for pid in $(awk '{print $2}' <<< "${killables}") ; do
  for signal in TERM TERM TERM KILL ; do
    echo "killing ${pid} with ${signal} ..."
    if ! kill -s ${signal} ${pid} ; then
      echo "${pid} killed"
      break
    fi
    sleep 1
  done
done
