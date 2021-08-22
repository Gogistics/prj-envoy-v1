#!/bin/sh

chmod go+r /redis-envoy-config.yaml &&
  envoy -c /redis-envoy-config.yaml