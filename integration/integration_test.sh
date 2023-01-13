#!/bin/sh

set -e

echo 'Ensure both app and database are running first!'

# creating required data
alias psqlcmd='psql -h localhost -p 5434 -d baas -U liberdade -c' 

export PGPASSWORD='password'
lua_script=`cat test_action_script.lua`

psqlcmd "INSERT INTO clients(email, password, is_admin) VALUES('test@go.dev','password','off') ON CONFLICT DO NOTHING;"
psqlcmd "INSERT INTO apps(owner_id,name) VALUES(1,'go test app') ON CONFLICT DO NOTHING;"
psqlcmd "INSERT INTO app_memberships(app_id,client_id,role) VALUES(1,1,'admin') ON CONFLICT DO NOTHING RETURNING *;"
psqlcmd "INSERT INTO users(app_id,email,password) VALUES(1,'test@go.dev','password') ON CONFLICT DO NOTHING;"
psqlcmd "INSERT INTO actions(app_id,name,script) VALUES (1,'Test Action','') ON CONFLICT DO NOTHING;"
psqlcmd "UPDATE actions SET script='$lua_script' WHERE id='1';"

psqlcmd "SELECT * FROM clients;"
psqlcmd "SELECT * FROM apps;"
psqlcmd "SELECT * FROM users;"
psqlcmd "SELECT * FROM actions;"

curl -X POST \
     -H 'Content-Type: application/json' \
     -d '{"action_name":"Test Action","app_id":1,"user_id":1,"params":{"name":"Joe","age":28}}' \
     http://localhost:8080/actions/run

