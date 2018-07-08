#!/bin/bash

killables=$(ps -ef | awk '$8=="/openbank/services/lake/entrypoint" {print $2}')

if [[ -n "${killables// }" ]] ; then
  echo "Already running"
  exit 0
else
  echo "Starting"
fi

. /openbank/services/lake/params.conf

/openbank/services/lake/entrypoint
