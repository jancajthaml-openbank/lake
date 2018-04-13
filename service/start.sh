#!/bin/bash

killables=$(ps aux | grep lake)

if [ ! "${killables}" = "" ] ; then
  echo "Already running"
  return
fi

. /openbank/services/lake/params.conf

/openbank/services/lake/entrypoint
