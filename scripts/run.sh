#!/usr/bin/env bash
# postgres username
 export      DB_USER="geouser"
# user password
 export      DB_PASS="Wipaydb8##"
# name of the database
 export      DB_NAME="geodb"
# host
 export      DB_HOST="localhost"
# ssl mode ("enable" or "disable")
 export      DB_SSL_MODE="disable"
# endpoint of the application Ex( /location )
 export      ENDPOINT="/location"
# server address Ex(localhost:8080)
 export      SERVER_ADDRESS=":8080"
 # refresh token expiration (in hours)
 export     REFRESH_TOKEN_EXP="360h"
 # access token expiration(in minutes or in hours)
 export     TEMP_TOKEN_EXP="15m"
 # json web token secret
 export     SECRET="XXXXX"

echo "user = '${DB_USER}'"
if [ -z "${DB_PASS}" ] ;then echo "user password = 'empty'";else echo "user password = 'xxxxxxxx'";fi
echo "database name = '${DB_NAME}'"
echo "database host = '${DB_HOST}'"
echo "SSL mode = '${DB_SSL_MODE}'"
echo "application endpoint = '${ENDPOINT}'"
echo "application address = '${SERVER_ADDRESS}'"
echo "expiration time of refresh token = '${REFRESH_TOKEN_EXP}'"
echo "expiration time of access token = '${TEMP_TOKEN_EXP}'"
if [ -z "${SECRET}" ] ;then echo "secret = 'empty'";else echo "secret = 'xxxxxxxx'";fi

DIR="$( cd "$(dirname "$0")" >/dev/null 2>&1 ; pwd -P )"
echo "Running..."
go run "${DIR}"/../
