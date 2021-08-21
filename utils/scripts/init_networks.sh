#!/bin/bash

CWD=$(pwd)

trap "finish" INT TERM

finish() {
    # do something while things go wrong
    local existcode=$?
    exit $existcode
}

FRONT_ENVOY_NETWORK="atai_envoy"
FRONT_ENVOY_NETWORK_INSPECTION=$(docker network inspect $FRONT_ENVOY_NETWORK)
FRONT_ENVOY_NETWORK_INSPECTION=$?

if [ $FRONT_ENVOY_NETWORK_INSPECTION -ne 0 ]
then
    echo "Creating $FRONT_ENVOY_NETWORK network..."
    docker network create \
        --driver="bridge" \
        --subnet="172.10.0.0/24" \
        --gateway="172.10.0.1" \
        $FRONT_ENVOY_NETWORK
else
    echo "${FRONT_ENVOY_NETWORK} already exists"
fi

