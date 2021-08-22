#!/bin/sh


# ref: https://github.com/nginxinc/docker-nginx/blob/4e5332fa50a1f8f73657417c6bfe249bbb3b110d/Dockerfile
apk update &&
    mkdir /.ssh &&
    cp atai-envoy.com.key atai-envoy.com.crt /.ssh/ &&
    tar -xvzf angular-app.tar.gz -C / &&
    chmod go+r /.ssh/* /custom-nginx.conf &&
    /usr/sbin/nginx -c /custom-nginx.conf -g "daemon off;"