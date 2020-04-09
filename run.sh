#!/usr/bin/env bash
# postgres username
 export      DB_USER="postgres"
# user password
 export      DB_PASS="nightmare666"
# name of the database
 export      DB_NAME="location"
# host
 export      DB_HOST="localhost"
# ssl mode ("enable" or "disable")
 export      DB_SSL_MODE="disable"
# endpoint of the application Ex( /location )
 export      ENDPOINT="/location"
# server address Ex(localhost:8080)
 export      SERVER_ADDRESS=":8080"


DIR="$( cd "$(dirname "$0")" >/dev/null 2>&1 ; pwd -P )"
go run "${DIR}"/
