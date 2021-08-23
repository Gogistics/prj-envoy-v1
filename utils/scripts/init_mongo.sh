#!/bin/sh

# why execute the file by /bin/sh
# ref: https://stackoverflow.com/questions/44803982/how-do-i-run-a-bash-script-in-an-alpine-docker-container

trap "finish" INT TERM

# finish
finish() {
    local existcode=$?
    exit $existcode
}

load_mongo() {
    # load mongo
    apk update &&
    apk add --no-cache mongodb &&
    mkdir /data
}

start_mongo_for_init() {
    # start mongo service
    # ref: https://mrkaran.dev/posts/netcat-port/
    /usr/bin/mongod --dbpath /data --nojournal &
    while ! nc -vz localhost 27017 ; do sleep 1; done
}

init_users_dbs() {
    # init dbs
    # add more roles here
    echo "creating users for dbs..."
    mongo web --eval "db.createUser({ user: 'web_admin_user', pwd: 'web-1234567', roles: [ { role: 'userAdminAnyDatabase', db: 'admin' } ] });"
    mongo web --eval "db.createUser({ user: 'web_test_user', pwd: 'web-1234567', roles: [ { role: 'readWrite', db: 'web' } ] });"
}

set -e

load_mongo && \
start_mongo_for_init && \
init_users_dbs && \
/usr/bin/mongod --dbpath /data --shutdown && \
# restart mongo
sleep 2 && \
/usr/bin/mongod --dbpath /data --bind_ip "0.0.0.0"
