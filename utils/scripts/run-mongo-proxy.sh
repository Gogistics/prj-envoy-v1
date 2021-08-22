#!/bin/sh

chmod go+r /mongo-envoy-config.yaml &&
  envoy -c /mongo-envoy-config.yaml