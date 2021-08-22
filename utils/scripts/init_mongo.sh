#!/bin/sh

# why execute the file by /bin/sh
# Ref: https://stackoverflow.com/questions/44803982/how-do-i-run-a-bash-script-in-an-alpine-docker-container
set -e

trap "finish" INT TERM

# finish
finish() {
    local existcode=$?
    exit $existcode
}

set +e

init_users_dbs() {
    # add more roles here
    echo "creating users for dbs..."
    mongo web --eval "db.createUser({ user: 'web_admin_user', pwd: 'web-1234567', roles: [ { role: 'userAdminAnyDatabase', db: 'admin' } ] });"
    mongo web --eval "db.createUser({ user: 'web_test_user', pwd: 'web-1234567', roles: [ { role: 'readWrite', db: 'web' } ] });"
}

# load mongo
apk update &&
apk add --no-cache mongodb &&
mkdir /data &&

# start mongo service
/usr/bin/mongod --dbpath /data --nojournal &
while ! nc -vz localhost 27017 ; do sleep 1; done

# init dbs
init_users_dbs &&

# stop
/usr/bin/mongod --dbpath /data --shutdown &&

# restart
sleep 2 &&
/usr/bin/mongod --dbpath /data --bind_ip "0.0.0.0"
set -e
