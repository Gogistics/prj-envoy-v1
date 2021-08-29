# Tutorial of Envoy, etc. (WIP)
This tutorial aims to share my experience of using Envoy. The demo is going to show you how to deploy Enovy proxies, web app in Golang, Redis, Mongo, and Nginx, and have them communicate with each other. With the basic understanding of the communicaiton between components, we will be able to debug the issues of service mesh more efficiently. Slides are [here](https://docs.google.com/presentation/d/1Pwcz2QOR7TnffP0VgZ8zxweUiakeF46V6VoPI_gt3rc/edit?usp=sharing).


## Introduction of Envoy
Envoy is a high-performance communication bus designed for large modern service-oriented architectures. And Envoy has been developed with the belief that: **The network should be transparent to applications. When network and application problems do occur it should be easy to determine the source of the problem.** Matt Klein and bunch of contributors already shared their knowledge and information about Envoy in many places, such as CNCF events and talks, YouTube, etc. Strongly suggest reading and watching those resources to know more about Envoy and enjoy developing new stuff with Envoy!


[Overview of Envoy architecture](https://drive.google.com/file/d/12I08q2M9WeuaVqIyk8ZwDFTYxJ8K0p4x/view?usp=sharing)

* Out of process architecture
* Modern C++ code base
* L3 & L4 filter architecture
* L7 filter architecture and routing
* First class HTTP/2 support
* HTTP/3 support (alpha)
* Service discovery and dynamic configuration
* Active/passive health checking
* Advanced load balancing
* Front/edge proxy support
* Best in class observability
* Scalability/extendability


## Demo (WIP)
[Demo diagram](https://drive.google.com/file/d/1vxLz5n-xSXl-OgHwXULtl5MUyzSk7Sdg/view?usp=sharing)

Steps of running the demo are as follows:

1. Create certs

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

$ openssl req -new -sha256 \
     -key certs/atai-envoy.com.key \
     -subj "/C=US/ST=CA/O=GOGISTICS, Inc./CN=atai-envoy.com" \
     -out certs/grpc.atai-envoy.com.csr
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

# for grpc
$ openssl x509 -req \
     -in certs/grpc.atai-envoy.com.csr \
     -CA certs/ca.crt \
     -CAkey certs/ca.key \
     -CAcreateserial \
     -extfile <(printf "subjectAltName=DNS:atai-envoy.com") \
     -out certs/grpc.atai-envoy.com.crt \
     -days 500 \
     -sha256
# Signature ok
# subject=/C=US/ST=CA/O=GOGISTICS, Inc./CN=atai-envoy.com
# Getting CA Private Key
# \generate certificates for each proxy
```


2. Build API app in Golang
```sh
# create a git repo. in cloud and come back to the project repo. to init Golang mod
$ go mod init github.com/Gogistics/prj-envoy-v1

# add module requirements and sums
$ go mod tidy
```

Notes of developing Golang locally
```sh
# bring up redis and mongo
$ docker run -d \
    --name redis_standalone \
    --network atai_envoy \
    --ip "172.10.0.61" \
    redis:alpine

# run mongo
$ bazel build --platforms=@io_bazel_rules_go//go/toolchain:linux_amd64 \
    //envoys:mongo-envoy-v0.0.0
$ bazel run --platforms=@io_bazel_rules_go//go/toolchain:linux_amd64 \
    //envoys:mongo-envoy-v0.0.0
$ docker run -d \
    --name mongo_standalone \
    --network atai_envoy \
    --ip "172.10.0.71" \
    alantai/databases:mongo-v0.0.0

# go to golang app dir and run the command below
$ docker run --name atai-envoy \
    --network atai_envoy \
    --ip "172.10.0.3" \
    -v $(pwd):/prj \
    -w /prj \
    -it \
    --rm \
    golang:latest bash

# run golang app in dev mode
$ go run main -dev

```


3. General setup and build by Bazel
Write WORKSPACE and its corresponding BUILD.bazel and run the following commands
```sh
# run the gazelle target specified in the BUILD file
$ bazel run //:gazelle

# update repos
$ bazel run //:gazelle -- update-repos -from_file=go.mod -to_macro=deps.bzl%go_dependencies

# build container
$ bazel build --platforms=@io_bazel_rules_go//go/toolchain:linux_amd64 \
    //services/api-v1:api-v0.0.0
$ bazel run --platforms=@io_bazel_rules_go//go/toolchain:linux_amd64 \
    //services/api-v1:api-v0.0.0

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
    alantai/services/api-v1:api-v0.0.0


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


4. Build Docker images and run all containers
```sh
$ bazel build --platforms=@io_bazel_rules_go//go/toolchain:linux_amd64 //envoys:all

# Or run build one by one

$ bazel build --platforms=@io_bazel_rules_go//go/toolchain:linux_amd64 //envoys:front-envoy-v0.0.0
$ bazel run --platforms=@io_bazel_rules_go//go/toolchain:linux_amd64 //envoys:front-envoy-v0.0.0

$ bazel build --platforms=@io_bazel_rules_go//go/toolchain:linux_amd64 //envoys:redis-envoy-v0.0.0
$ bazel run --platforms=@io_bazel_rules_go//go/toolchain:linux_amd64 //envoys:redis-envoy-v0.0.0

$ bazel build --platforms=@io_bazel_rules_go//go/toolchain:linux_amd64 //envoys:mongo-envoy-v0.0.0
$ bazel run --platforms=@io_bazel_rules_go//go/toolchain:linux_amd64 //envoys:mongo-envoy-v0.0.0

$ bazel build --platforms=@io_bazel_rules_go//go/toolchain:linux_amd64 //databases:mongo-v0.0.0
$ bazel run --platforms=@io_bazel_rules_go//go/toolchain:linux_amd64 //databases:mongo-v0.0.0

$ bazel build --platforms=@io_bazel_rules_go//go/toolchain:linux_amd64 //services/nginx-v1:nginx-v0.0.0
$ bazel run --platforms=@io_bazel_rules_go//go/toolchain:linux_amd64 //services/nginx-v1:nginx-v0.0.0

$ bazel build --platforms=@io_bazel_rules_go//go/toolchain:linux_amd64 //services/api-v1:api-v0.0.0
$ bazel run --platforms=@io_bazel_rules_go//go/toolchain:linux_amd64 //services/api-v1:api-v0.0.0

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

# run nginx to serve frontend static files
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

# run envoy front proxy
$ docker run -d \
      --name atai_envoy_front \
      -p 80:80 -p 443:443 -p 10000:10000 -p 8001:8001 \
      --network atai_envoy \
      --ip "172.10.0.10" \
      --log-opt mode=non-blocking \
      --log-opt max-buffer-size=5m \
      --log-opt max-size=100m \
      --log-opt max-file=5 \
      alantai/envoys:front-envoy-v0.0.0

# run envoy proxy of redis
$ docker run -d \
      --name atai_envoy_redis \
      --network atai_envoy \
      --ip "172.10.0.50" \
      --log-opt mode=non-blocking \
      --log-opt max-buffer-size=5m \
      --log-opt max-size=100m \
      --log-opt max-file=5 \
      alantai/envoys:redis-envoy-v0.0.0

# run envoy proxy of mongo
$ docker run -d \
      --name atai_envoy_mongo \
      --network atai_envoy \
      --ip "172.10.0.55" \
      --log-opt mode=non-blocking \
      --log-opt max-buffer-size=5m \
      --log-opt max-size=100m \
      --log-opt max-file=5 \
      alantai/envoys:mongo-envoy-v0.0.0


# run api service
$ docker run -d \
    --name atai_service_api_v1_1 \
    --network atai_envoy \
    --ip "172.10.0.21" \
    --log-opt mode=non-blocking \
    --log-opt max-buffer-size=5m \
    --log-opt max-size=100m \
    --log-opt max-file=5 \
    alantai/services/api-v1:api-v0.0.0

$ docker run -d \
    --name atai_service_api_v1_2 \
    --network atai_envoy \
    --ip "172.10.0.22" \
    --log-opt mode=non-blocking \
    --log-opt max-buffer-size=5m \
    --log-opt max-size=100m \
    --log-opt max-file=5 \
    alantai/services/api-v1:api-v0.0.0
# \run api service


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


# test nginx servers
$ curl -k -vvv https://atai-envoy.com
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
# * Using Stream ID: 1 (easy handle 0x7f885480d600)
# > GET / HTTP/2
# > Host: atai-envoy.com
# > User-Agent: curl/7.64.1
# > Accept: */*
# > 
# * Connection state changed (MAX_CONCURRENT_STREAMS == 2147483647)!
# < HTTP/2 200 
# < server: envoy
# < date: Sun, 22 Aug 2021 23:56:55 GMT
# < content-type: text/html; charset=utf-8
# < content-length: 546
# < last-modified: Sat, 01 Jan 2000 00:00:00 GMT
# < etag: "386d4380-222"
# < accept-ranges: bytes
# < x-envoy-upstream-service-time: 4
# < 
# <!DOCTYPE html><html lang="en"><head>
#   <meta charset="utf-8">
#   <title>WebFrontendAngular</title>
#   <base href="/">
#   <meta name="viewport" content="width=device-width, initial-scale=1">
#   <link rel="icon" type="image/x-icon" href="favicon.ico">
# <link rel="stylesheet" href="styles.31d6cfe0d16ae931b73c.css"></head>
# <body>
#   <app-root></app-root>
# <script src="runtime.b69dca14a0abc60a86b6.js" defer></script><script src="polyfills.717592f80c33ea26a5b6.js" defer></script><script src="main.ec81fdd375efa010230b.js" defer></script>

# * Connection #0 to host atai-envoy.com left intact
# </body></html>* Closing connection 0
```


### MIS
* Docker
```sh
# remove images with tag <none>
$ docker rmi $(docker images --filter "dangling=true")

```

* MongoDB

Ref:
- https://docs.mongodb.com/v4.0/reference/

