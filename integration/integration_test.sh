#!/bin/sh

set -e

echo 'Ensure app is running first!'

# TODO create owner client
# TODO create test app
# TODO create test user
# TODO create test action

curl -X POST \
     -H 'Content-Type: application/json' \
     -d '{"action_name":"Test Action","app_id":1,"user_id":2,"params":{"name":"Joe","age":28}}' \
     http://localhost:8080/actions/run

