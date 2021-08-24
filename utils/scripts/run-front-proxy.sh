#!/bin/sh

apk add --update --no-cache ca-certificates &&
  # copy certs to envoy/certs
  mkdir -p /etc/envoy/certs/ &&
  cp atai-envoy.com.crt atai-envoy.com.key /etc/envoy/certs/ &&
  chmod 744 /etc/envoy/certs/* &&
  mkdir -p /usr/local/share/ca-certificates/extra/ &&
  cp atai-envoy.com.crt /usr/local/share/ca-certificates/ &&
  cat /usr/local/share/ca-certificates/atai-envoy.com.crt >> /etc/ssl/certs/ca-certificates.crt &&
  cp custom-ca-certificates.crt /usr/local/share/ca-certificates/ &&
  cp custom-ca-certificates.crt /usr/local/share/ca-certificates/extra/ &&
  update-ca-certificates &&
  chmod go+r /front-envoy-config.yaml &&
  envoy -c /front-envoy-config.yaml