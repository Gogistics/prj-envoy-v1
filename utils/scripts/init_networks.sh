#!/bin/bash

CWD=$(pwd)

trap "finish" INT TERM

finish() {
    # do something while things go wrong
    local existcode=$?
    exit $existcode
}

ENVOY_NETWORK="atai_envoy"
ENVOY_NETWORK_INSPECTION=$(docker network inspect $ENVOY_NETWORK)
ENVOY_NETWORK_INSPECTION=$?

if [ $ENVOY_NETWORK_INSPECTION -ne 0 ]
then
    echo "Creating $ENVOY_NETWORK network..."
    docker network create \
        --driver="bridge" \
        --subnet="172.10.0.0/24" \
        --gateway="172.10.0.1" \
        $ENVOY_NETWORK
else
    echo "${ENVOY_NETWORK} already exists"
fi

GRPC_NETWORK="atai_grpc"
GRPC_NETWORK_INSPECTION=$(docker network inspect $GRPC_NETWORK)
GRPC_NETWORK_INSPECTION=$?

if [ $GRPC_NETWORK_INSPECTION -ne 0 ]
then
    echo "Creating $GRPC_NETWORK network..."
    docker network create \
        --driver="bridge" \
        --subnet="172.11.0.0/24" \
        --gateway="172.11.0.1" \
        $GRPC_NETWORK
else
    echo "${GRPC_NETWORK} already exists"
fi

