#!/bin/sh

set -e

echo 'Ensure app is running first!'

curl -X POST -H 'Content-Type: application/json' \
     -d '{"app_name":"Shiny App","age":28}' \
     http://localhost:8080/actions/run

