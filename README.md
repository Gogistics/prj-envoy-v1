# Envoy Intro. (WIP)

### Create certs

```sh
$ cd infra/


# create a cert authority
$ openssl genrsa -out certs/ca.key 4096
# Generating RSA private key, 4096 bit long modulus
# .....++
# ...............................................................................++
# e is 65537 (0x10001)


$ openssl req -x509 -new -nodes -key certs/ca.key -sha256 -days 1024 -out certs/ca.crt
# You are about to be asked to enter information that will be incorporated
# into your certificate request.
# What you are about to enter is what is called a Distinguished Name or a DN.
# There are quite a few fields but you can leave some blank
# For some fields there will be a default value,
# If you enter '.', the field will be left blank.
# -----
# Country Name (2 letter code) []:TW
# State or Province Name (full name) []:Taiwan
# Locality Name (eg, city) []:Kaohsiung
# Organization Name (eg, company) []:Gogistics
# Organizational Unit Name (eg, section) []:infra
# Common Name (eg, fully qualified host name) []:
# Email Address []:gogistics@gogistics-tw.com
# Alans-MacBook-Pro-2:infra alantai$ 
# Alans-MacBook-Pro-2:infra alantai$ 
# Alans-MacBook-Pro-2:infra alantai$ 
# Alans-MacBook-Pro-2:infra alantai$ 
# Alans-MacBook-Pro-2:infra alantai$ openssl genrsa -out certs/atai-envoy.com.key 2048
# Generating RSA private key, 2048 bit long modulus
# ...............................................................+++
# ..............................................+++
# e is 65537 (0x10001)
# \create a cert authority

# create a domain key
$ openssl genrsa -out certs/atai-envoy.com.key 2048
# Generating RSA private key, 2048 bit long modulus
# ...............................................................+++
# ..............................................+++
# e is 65537 (0x10001)


# generate signing requests for proxy and app
$ openssl req -new -sha256 \
     -key certs/atai-envoy.com.key \
     -subj "/C=US/ST=CA/O=GOGISTICS, Inc./CN=atai-envoy.com" \
     -out certs/atai-envoy.com.csr

$ openssl req -new -sha256 \
     -key certs/atai-envoy.com.key \
     -subj "/C=US/ST=CA/O=GOGISTICS, Inc./CN=atai-envoy.com" \
     -out certs/dev.atai-envoy.com.csr
# \generate signing requests for proxy and app


# generate certificates for each proxy
$ openssl x509 -req \
     -in certs/atai-envoy.com.csr \
     -CA certs/ca.crt \
     -CAkey certs/ca.key \
     -CAcreateserial \
     -extfile <(printf "subjectAltName=DNS:atai-envoy.com") \
     -out certs/atai-envoy.com.crt \
     -days 500 \
     -sha256
# Signature ok
# subject=/C=US/ST=CA/O=GOGISTICS, Inc./CN=atai-envoy.com
# Getting CA Private Key

# for dev
$ openssl x509 -req \
     -in certs/dev.atai-envoy.com.csr \
     -CA certs/ca.crt \
     -CAkey certs/ca.key \
     -CAcreateserial \
     -extfile <(printf "subjectAltName=DNS:atai-envoy.com") \
     -out certs/dev.atai-envoy.com.crt \
     -days 500 \
     -sha256
# Signature ok
# subject=/C=US/ST=CA/O=GOGISTICS, Inc./CN=atai-envoy.com
# Getting CA Private Key
# \generate certificates for each proxy
```


### Golang
```sh
# create a git repo. in cloud and back to init Golang mod
$ go mod init github.com/Gogistics/prj-envoy-v1

# add module requirements and sums
$ go mod tidy
```


### General setup and build by Bazel
Write WORKSPACE and its corresponding BUILD.bazel and run the following commands
```sh
# run the gazelle target specified in the BUILD file
$ bazel run //:gazelle

# update repos
$ bazel run //:gazelle -- update-repos -from_file=go.mod -to_macro=deps.bzl%go_dependencies

# build container
$ bazel build --platforms=@io_bazel_rules_go//go/toolchain:linux_amd64 //services/api-v1:atai-envoy-api-v0.0.0
$ bazel run --platforms=@io_bazel_rules_go//go/toolchain:linux_amd64 //services/api-v1:atai-envoy-api-v0.0.0

# after building the image, check if the image exists
$ docker images # in my case, the image repository is alantai/api-app and the tag is atai-v0.0.0

# test image by spinning up an api container
$ docker run -d \
    -p 8443:443 \
    --name atai_envoy_service_api_v1 \
    --network atai_envoy \
    --ip "172.10.0.21" \
    --log-opt mode=non-blocking \
    --log-opt max-buffer-size=5m \
    --log-opt max-size=100m \
    --log-opt max-file=5 \
    alantai/services/api-v1:atai-envoy-api-v0.0.0


$ curl -k https://0.0.0.0:8443/api/v1 -vvv
# *   Trying 0.0.0.0...
# * TCP_NODELAY set
# * Connected to 0.0.0.0 (127.0.0.1) port 8443 (#0)
# * ALPN, offering h2
# * ALPN, offering http/1.1
# * successfully set certificate verify locations:
# *   CAfile: /etc/ssl/cert.pem
#   CApath: none
# * TLSv1.2 (OUT), TLS handshake, Client hello (1):
# * TLSv1.2 (IN), TLS handshake, Server hello (2):
# * TLSv1.2 (IN), TLS handshake, Certificate (11):
# * TLSv1.2 (IN), TLS handshake, Server key exchange (12):
# * TLSv1.2 (IN), TLS handshake, Server finished (14):
# * TLSv1.2 (OUT), TLS handshake, Client key exchange (16):
# * TLSv1.2 (OUT), TLS change cipher, Change cipher spec (1):
# * TLSv1.2 (OUT), TLS handshake, Finished (20):
# * TLSv1.2 (IN), TLS change cipher, Change cipher spec (1):
# * TLSv1.2 (IN), TLS handshake, Finished (20):
# * SSL connection using TLSv1.2 / ECDHE-RSA-AES128-GCM-SHA256
# * ALPN, server accepted to use h2
# * Server certificate:
# *  subject: C=US; ST=CA; O=GOGISTICS, Inc.; CN=atai-envoy.com
# *  start date: Aug 19 18:31:14 2021 GMT
# *  expire date: Jan  1 18:31:14 2023 GMT
# *  issuer: C=TW; ST=Taiwan; L=Kaohsiung; O=Gogistics; OU=infra; emailAddress=gogistics@gogistics-tw.com
# *  SSL certificate verify result: unable to get local issuer certificate (20), continuing anyway.
# * Using HTTP2, server supports multi-use
# * Connection state changed (HTTP/2 confirmed)
# * Copying HTTP/2 data in stream buffer to connection buffer after upgrade: len=0
# * Using Stream ID: 1 (easy handle 0x7f91b880d600)
# > GET /api/v1 HTTP/2
# > Host: 0.0.0.0:8443
# > User-Agent: curl/7.64.1
# > Accept: */*
# > 
# * Connection state changed (MAX_CONCURRENT_STREAMS == 250)!
# < HTTP/2 200 
# < content-type: applicaiton/json; charset=utf-8
# < content-length: 61
# < date: Thu, 19 Aug 2021 22:51:41 GMT
# < 
# * Connection #0 to host 0.0.0.0 left intact
# {"Name":"Alan","Hostname":"1cfccd1c9a0a","Hobbies":["workout","programming","driving"]}* Closing connection 0

# take down the container
$ docker rm -f atai_envoy_service_api_v1

# login container registry
$ docker login

```


## Envoy
```sh
$ bazel build --platforms=@io_bazel_rules_go//go/toolchain:linux_amd64 //envoys:api-envoy-v0.0.0
$ bazel run --platforms=@io_bazel_rules_go//go/toolchain:linux_amd64 //envoys:api-envoy-v0.0.0

$ bazel build --platforms=@io_bazel_rules_go//go/toolchain:linux_amd64 //envoys:redis-envoy-v0.0.0
$ bazel run --platforms=@io_bazel_rules_go//go/toolchain:linux_amd64 //envoys:redis-envoy-v0.0.0

$ bazel build --platforms=@io_bazel_rules_go//go/toolchain:linux_amd64 //envoys:mongo-envoy-v0.0.0
$ bazel run --platforms=@io_bazel_rules_go//go/toolchain:linux_amd64 //envoys:mongo-envoy-v0.0.0

$ bazel build --platforms=@io_bazel_rules_go//go/toolchain:linux_amd64 //databases:mongo-v0.0.0
$ bazel run --platforms=@io_bazel_rules_go//go/toolchain:linux_amd64 //databases:mongo-v0.0.0

$ bazel build --platforms=@io_bazel_rules_go//go/toolchain:linux_amd64 //services/nginx-v1:nginx-v0.0.0
$ bazel run --platforms=@io_bazel_rules_go//go/toolchain:linux_amd64 //services/nginx-v1:nginx-v0.0.0

# run redis
$ docker run -d \
    --name redis_standalone \
    --network atai_envoy \
    --ip "172.10.0.61" \
    redis:alpine

# run mongo
$ docker run -d \
    --name mongo_standalone \
    --network atai_envoy \
    --ip "172.10.0.71" \
    alantai/databases:mongo-v0.0.0

# run nginx
$ docker run -d \
    --name nginx_web_server_1 \
    --network atai_envoy \
    --ip "172.10.0.111" \
    alantai/services/nginx-v1:nginx-v0.0.0
$ docker run -d \
    --name nginx_web_server_2 \
    --network atai_envoy \
    --ip "172.10.0.112" \
    alantai/services/nginx-v1:nginx-v0.0.0

# run api service
$ docker run -d \
      --name atai_envoy_api_service \
      -p 80:80 -p 443:443 -p 10000:10000 -p 8001:8001 \
      --network atai_envoy \
      --ip "172.10.0.10" \
      --log-opt mode=non-blocking \
      --log-opt max-buffer-size=5m \
      --log-opt max-size=100m \
      --log-opt max-file=5 \
      alantai/envoys:api-envoy-v0.0.0

$ docker run -d \
      --name atai_envoy_redis \
      --network atai_envoy \
      --ip "172.10.0.50" \
      --log-opt mode=non-blocking \
      --log-opt max-buffer-size=5m \
      --log-opt max-size=100m \
      --log-opt max-file=5 \
      alantai/envoys:redis-envoy-v0.0.0

$ docker run -d \
      --name atai_envoy_mongo \
      --network atai_envoy \
      --ip "172.10.0.55" \
      --log-opt mode=non-blocking \
      --log-opt max-buffer-size=5m \
      --log-opt max-size=100m \
      --log-opt max-file=5 \
      alantai/envoys:mongo-envoy-v0.0.0

$ docker run -d \
    --name atai_envoy_service_api_v1_1 \
    --network atai_envoy \
    --ip "172.10.0.21" \
    --log-opt mode=non-blocking \
    --log-opt max-buffer-size=5m \
    --log-opt max-size=100m \
    --log-opt max-file=5 \
    alantai/services/api-v1:atai-envoy-api-v0.0.0

$ docker run -d \
    --name atai_envoy_service_api_v1_2 \
    --network atai_envoy \
    --ip "172.10.0.22" \
    --log-opt mode=non-blocking \
    --log-opt max-buffer-size=5m \
    --log-opt max-size=100m \
    --log-opt max-file=5 \
    alantai/services/api-v1:atai-envoy-api-v0.0.0


# run service proxy and the api containers to test the api service; edit /etc/hosts by adding 0.0.0.0 atai-envoy.com
$ curl -k -vvv https://atai-envoy.com/api/v1
# *   Trying 0.0.0.0...
# * TCP_NODELAY set
# * Connected to atai-envoy.com (127.0.0.1) port 443 (#0)
# * ALPN, offering h2
# * ALPN, offering http/1.1
# * successfully set certificate verify locations:
# *   CAfile: /etc/ssl/cert.pem
#   CApath: none
# * TLSv1.2 (OUT), TLS handshake, Client hello (1):
# * TLSv1.2 (IN), TLS handshake, Server hello (2):
# * TLSv1.2 (IN), TLS handshake, Certificate (11):
# * TLSv1.2 (IN), TLS handshake, Server key exchange (12):
# * TLSv1.2 (IN), TLS handshake, Request CERT (13):
# * TLSv1.2 (IN), TLS handshake, Server finished (14):
# * TLSv1.2 (OUT), TLS handshake, Certificate (11):
# * TLSv1.2 (OUT), TLS handshake, Client key exchange (16):
# * TLSv1.2 (OUT), TLS change cipher, Change cipher spec (1):
# * TLSv1.2 (OUT), TLS handshake, Finished (20):
# * TLSv1.2 (IN), TLS change cipher, Change cipher spec (1):
# * TLSv1.2 (IN), TLS handshake, Finished (20):
# * SSL connection using TLSv1.2 / ECDHE-RSA-CHACHA20-POLY1305
# * ALPN, server accepted to use h2
# * Server certificate:
# *  subject: C=US; ST=CA; O=GOGISTICS, Inc.; CN=atai-envoy.com
# *  start date: Aug 19 18:31:14 2021 GMT
# *  expire date: Jan  1 18:31:14 2023 GMT
# *  issuer: C=TW; ST=Taiwan; L=Kaohsiung; O=Gogistics; OU=infra; emailAddress=gogistics@gogistics-tw.com
# *  SSL certificate verify result: unable to get local issuer certificate (20), continuing anyway.
# * Using HTTP2, server supports multi-use
# * Connection state changed (HTTP/2 confirmed)
# * Copying HTTP/2 data in stream buffer to connection buffer after upgrade: len=0
# * Using Stream ID: 1 (easy handle 0x7ffb0e80d600)
# > GET /api/v1 HTTP/2
# > Host: atai-envoy.com
# > User-Agent: curl/7.64.1
# > Accept: */*
# > 
# * Connection state changed (MAX_CONCURRENT_STREAMS == 2147483647)!
# < HTTP/2 200 
# < content-type: applicaiton/json; charset=utf-8
# < content-length: 61
# < date: Fri, 20 Aug 2021 01:28:52 GMT
# < x-envoy-upstream-service-time: 1
# < server: envoy
# < 
# * Connection #0 to host atai-envoy.com left intact
# {"Name":"Alan","Hostname":"1cfccd1c9a0a","Hobbies":["workout","programming","driving"]}* Closing connection 0
```

