# Tutorial of Envoy, etc.
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


## Demo
In this demo, we are going to spin up a full container topology illustrated in [demo diagram](https://drive.google.com/file/d/1vxLz5n-xSXl-OgHwXULtl5MUyzSk7Sdg/view?usp=sharing). All Docker images are stored at [DockerHub](https://hub.docker.com/repository/docker/alantai/prj-envoy-v1).

Basically, the container topology comprises the components as follows:
* Envoy proxies
* API applications written in Golang
* Nginx CDNs
* MongoDB
* Redis
* gRPC server and client

And the folder directories are as follows:
* ./databases/ contains a Bazel file for building database Docker images; in this demo, Redis and Mongo standalone will be built.
* ./envoys/ contains a Bazel file for building Envoy Docker images; in this demo, there are four kinds of Envoy proxies will be built.
* ./services/ contains api-v1/, grpc-v1/, nginx-v1/, web-frontend-angular/, and web-frontend-react/. api-v1/ is for developing API application in Golang and a Bazel file for building API Docker image, grpc-v1/ is for developing gRPC server and client in Golang and Bazel files for building their Docker images. nginx-v1/ contains a Bazel file for building Nginx Docker image.
* ./utils/ contains certs/, configs/, dockerfiles/, and scripts/


### Context
Based on the topology, what we want to achieve are as follows:
* End users interact with the remote API service through Envoy front proxy.
* The API service interacts with Redis, Mongo, and gRPC server throught different kinds of Envoy proxies.

### Prerequisites
Install Docker, Bazel, etc. on the Linux environment. I wrote this demo and built the whole topology on my Macbook Pro.

### Steps of building this demo from scratch
1. Draw the network/container topology
2. Create Docker network and certs
3. Write API application and gRPC server in Golang
4. Build Docker images
5. Bring up all the Docker containers and test the API service, gRPC service, and Nignx
6. If you are interested in CI pipeline, you can think of the CI pipeline design and the implementation, especially integrate Bazel builds into CI pipeline. [Here](https://drive.google.com/file/d/1PZAZ9NlV8nXpVqpNvGgmvvKje6SFzFQp/view?usp=sharing) is a typical pipeline flow.

Now, we are going through parts of steps:

- Create certs
Idealy for production applications, don't put the certs in the course code repository. If you're using Kubernetes, suggest reading sensitive data from [Secrets](https://kubernetes.io/docs/concepts/configuration/secret/). For development/demo, it's okay to generate certs and put them in the source code repository.

```sh
$ cd utils/

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
# \generate certificates for each proxy
```

- Develop and build API app in Golang [design diagram of the web application in Golang](https://drive.google.com/file/d/1Jw5kNCA2c-gVy2K7fdNEpEF7b-0X1sQM/view?usp=sharing)
```sh
# create a git repo. in cloud and come back to the project root to init a Golang mod
$ go mod init github.com/Gogistics/prj-envoy-v1

# add module requirements and sums by running the command below if needed
$ go mod tidy

# then, let's write a app in Golang and connect it to Redis and Mongo. See /services/api-v1 as reference.

# 1. bring up redis and mongo

# build and run redis and mongo by Bazel and Docker
# run the gazelle target specified in the BUILD file
$ bazel run //:gazelle

# update repos deps
$ bazel run //:gazelle -- update-repos -from_file=go.mod -to_macro=deps.bzl%go_dependencies

$ bazel build --platforms=@io_bazel_rules_go//go/toolchain:linux_amd64 \
    //databases:redis-standalone-v0.0.0
$ bazel run --platforms=@io_bazel_rules_go//go/toolchain:linux_amd64 \
    //databases:redis-standalone-v0.0.0

$ bazel build --platforms=@io_bazel_rules_go//go/toolchain:linux_amd64 \
    //envoys:mongo-envoy-v0.0.0
$ bazel run --platforms=@io_bazel_rules_go//go/toolchain:linux_amd64 \
    //envoys:mongo-envoy-v0.0.0

$ docker run -d \
    --name redis_standalone \
    --network atai_envoy \
    --ip "172.10.0.61" \
    alantai/prj-envoy-v1/databases:redis-standalone-v0.0.0
$ docker run -d \
    --name mongo_standalone \
    --network atai_envoy \
    --ip "172.10.0.71" \
    alantai/prj-envoy-v1/databases:mongo-standalone-v0.0.0

# 2. run the command below at api-v1/
$ docker run --name atai-go-dev \
    --network atai_envoy \
    --ip "172.10.0.3" \
    -v $(pwd):/prj \
    -w /prj \
    -it \
    --rm \
    golang:latest bash

# 3. write REST APIs

# 4. run golang app in dev mode if a flag, dev, set for running the app dev mode
$ go run main -dev

# 5. then exec into the same container from the other terminal to test the Golang app which you just brought up
# go into the container
$ docker exec -it atai-go-dev sh

# test the app by curl
$ curl -k https://0.0.0.0/api/v1
```

### Steps of bringing up the whole topology with the existing code
1. Create certs (optional because all certs have been generated under ./utils/certs/)

2. Create the Docker networks
```sh
# run the command to init two networks for future development
$ ./utils/scripts/init_networks.sh
```

3. Build by Bazel

Write WORKSPACE and its corresponding BUILD.bazel at the project root dir (optional because both files are ready to use). And then run the following commands
```sh
# 1. run the gazelle target specified in the BUILD file
$ bazel run //:gazelle

# 2. update repos
$ bazel run //:gazelle -- update-repos -from_file=go.mod -to_macro=deps.bzl%go_dependencies


# 3. test the API app from outside the container (optional)
# build container image
$ bazel build --platforms=@io_bazel_rules_go//go/toolchain:linux_amd64 \
    //services/api-v1:api-v0.0.0
$ bazel run --platforms=@io_bazel_rules_go//go/toolchain:linux_amd64 \
    //services/api-v1:api-v0.0.0

# after building the image, check if the image exists
# in my case, the image repository is alantai/api-app and the tag is atai-v0.0.0
$ docker images

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
    alantai/prj-envoy-v1/services/api-v1:api-v0.0.0

# test the app by cURL
$ curl -k -vvv https://0.0.0.0:8443/api/v1
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

# Once the testnig has been completed, remove the container by the following command
$ docker rm -f atai_envoy_service_api_v1
# \3. test the API app from outside the container

# 4. build all required Docker images
$ bazel build --platforms=@io_bazel_rules_go//go/toolchain:linux_amd64 //databases:all
$ bazel build --platforms=@io_bazel_rules_go//go/toolchain:linux_amd64 //envoys:all
$ bazel build --platforms=@io_bazel_rules_go//go/toolchain:linux_amd64 //services/api-v1:all
$ bazel build --platforms=@io_bazel_rules_go//go/toolchain:linux_amd64 //services/grpc-v1/server:all
$ bazel build --platforms=@io_bazel_rules_go//go/toolchain:linux_amd64 //services/grpc-v1/client:all
$ bazel build --platforms=@io_bazel_rules_go//go/toolchain:linux_amd64 //services/nginx-v1:all

# Or run build one by one and then remember to run bazel run to generate Docker images
$ bazel build --platforms=@io_bazel_rules_go//go/toolchain:linux_amd64 //databases:mongo-standalone-v0.0.0
$ bazel run --platforms=@io_bazel_rules_go//go/toolchain:linux_amd64 //databases:mongo-standalone-v0.0.0

$ bazel build --platforms=@io_bazel_rules_go//go/toolchain:linux_amd64 //databases:redis-standalone-v0.0.0
$ bazel run --platforms=@io_bazel_rules_go//go/toolchain:linux_amd64 //databases:redis-standalone-v0.0.0

$ bazel build --platforms=@io_bazel_rules_go//go/toolchain:linux_amd64 //envoys:front-envoy-v0.0.0
$ bazel run --platforms=@io_bazel_rules_go//go/toolchain:linux_amd64 //envoys:front-envoy-v0.0.0

$ bazel build --platforms=@io_bazel_rules_go//go/toolchain:linux_amd64 //envoys:redis-envoy-v0.0.0
$ bazel run --platforms=@io_bazel_rules_go//go/toolchain:linux_amd64 //envoys:redis-envoy-v0.0.0

$ bazel build --platforms=@io_bazel_rules_go//go/toolchain:linux_amd64 //envoys:mongo-envoy-v0.0.0
$ bazel run --platforms=@io_bazel_rules_go//go/toolchain:linux_amd64 //envoys:mongo-envoy-v0.0.0

$ bazel build --platforms=@io_bazel_rules_go//go/toolchain:linux_amd64 //envoys:grpc-client-envoy-v0.0.0
$ bazel run --platforms=@io_bazel_rules_go//go/toolchain:linux_amd64 //envoys:grpc-client-envoy-v0.0.0

$ bazel build --platforms=@io_bazel_rules_go//go/toolchain:linux_amd64 //services/api-v1:api-v0.0.0
$ bazel run --platforms=@io_bazel_rules_go//go/toolchain:linux_amd64 //services/api-v1:api-v0.0.0

$ bazel build --platforms=@io_bazel_rules_go//go/toolchain:linux_amd64 //services/grpc-v1/server:grpc-query-server-v0.0.0
$ bazel run --platforms=@io_bazel_rules_go//go/toolchain:linux_amd64 //services/grpc-v1/server:grpc-query-server-v0.0.0

$ bazel build --platforms=@io_bazel_rules_go//go/toolchain:linux_amd64 //services/grpc-v1/client:grpc-query-client-v0.0.0
$ bazel run --platforms=@io_bazel_rules_go//go/toolchain:linux_amd64 //services/grpc-v1/client:grpc-query-client-v0.0.0

$ bazel build --platforms=@io_bazel_rules_go//go/toolchain:linux_amd64 //services/nginx-v1:nginx-v0.0.0
$ bazel run --platforms=@io_bazel_rules_go//go/toolchain:linux_amd64 //services/nginx-v1:nginx-v0.0.0

# 5. login to the registry and push the docker images to the registry
$ docker login

$ bazel run //databases:push-mongo-standalone
$ bazel run //databases:push-redis-standalone

$ bazel run //envoys:push-front-envoy
$ bazel run //envoys:push-redis-envoy
$ bazel run //envoys:push-mongo-envoy
$ bazel run //envoys:push-grpc-client-envoy

$ bazel run //services/api-v1:push-api
$ bazel run //services/grpc-v1/server:push-grpc-query-server
$ bazel run //services/grpc-v1/client:push-grpc-query-client
$ bazel run //services/nginx-v1:push-nginx
# \5. login to the registry and push the docker images to the container registry

# 6. bring up all containers
# run redis
$ docker run -d \
    --name redis_standalone \
    --network atai_envoy \
    --ip "172.10.0.61" \
    alantai/prj-envoy-v1/databases:redis-standalone-v0.0.0

# run mongo
$ docker run -d \
    --name mongo_standalone \
    --network atai_envoy \
    --ip "172.10.0.71" \
    alantai/prj-envoy-v1/databases:mongo-standalone-v0.0.0

# run nginx to serve frontend static files
$ docker run -d \
    --name nginx_web_server_1 \
    --network atai_envoy \
    --ip "172.10.0.111" \
    alantai/prj-envoy-v1/services/nginx-v1:nginx-v0.0.0
$ docker run -d \
    --name nginx_web_server_2 \
    --network atai_envoy \
    --ip "172.10.0.112" \
    alantai/prj-envoy-v1/services/nginx-v1:nginx-v0.0.0

# run envoy front proxy
# note: -p 8001:8001 for demo /admin
$ docker run -d \
      --name atai_envoy_front \
      -p 80:80 -p 443:443 -p 8001:8001 \
      --network atai_envoy \
      --ip "172.10.0.10" \
      --log-opt mode=non-blocking \
      --log-opt max-buffer-size=5m \
      --log-opt max-size=100m \
      --log-opt max-file=5 \
      alantai/prj-envoy-v1/envoys:front-envoy-v0.0.0

# run envoy proxy of redis
$ docker run -d \
      --name atai_envoy_redis \
      --network atai_envoy \
      --ip "172.10.0.50" \
      --log-opt mode=non-blocking \
      --log-opt max-buffer-size=5m \
      --log-opt max-size=100m \
      --log-opt max-file=5 \
      alantai/prj-envoy-v1/envoys:redis-envoy-v0.0.0

# run envoy proxy of mongo
$ docker run -d \
      --name atai_envoy_mongo \
      --network atai_envoy \
      --ip "172.10.0.55" \
      --log-opt mode=non-blocking \
      --log-opt max-buffer-size=5m \
      --log-opt max-size=100m \
      --log-opt max-file=5 \
      alantai/prj-envoy-v1/envoys:mongo-envoy-v0.0.0

# run grpc proxy
$ docker run -d \
      --name atai_envoy_grpc_client \
      --network atai_envoy \
      --ip "172.10.0.200" \
      --log-opt mode=non-blocking \
      --log-opt max-buffer-size=5m \
      --log-opt max-size=100m \
      --log-opt max-file=5 \
      alantai/prj-envoy-v1/envoys:grpc-client-envoy-v0.0.0

# connect atai_envoy_grpc_client to the other network, atai_grpc
$ docker network connect atai_grpc atai_envoy_grpc_client
# run grpc proxy

# run grpc
$ docker run \
    -itd \
    --name atai_grpc_server \
    --network atai_grpc \
    --ip "172.11.0.11" \
    -p 20000:20000 \
    alantai/prj-envoy-v1/services/grpc-v1/server:grpc-query-server-v0.0.0 \
    --port ":20000" \
    --certFile "atai-envoy.com.crt" \
    --keyFile "atai-envoy.com.key"

$ docker run \
    -itd \
    --name atai_grpc_client \
    --network atai_grpc \
    --ip "172.11.0.12" \
    alantai/prj-envoy-v1/services/grpc-v1/client:grpc-query-client-v0.0.0 \
    --caCert "atai-envoy.com.crt" \
    --serverName "atai-envoy.com" \
    --serverAddr "172.11.0.11:20000"
# \run grpc

# run api service
$ docker run -d \
    --name atai_service_api_v1_1 \
    --network atai_envoy \
    --ip "172.10.0.21" \
    --log-opt mode=non-blocking \
    --log-opt max-buffer-size=5m \
    --log-opt max-size=100m \
    --log-opt max-file=5 \
    alantai/prj-envoy-v1/services/api-v1:api-v0.0.0

$ docker run -d \
    --name atai_service_api_v1_2 \
    --network atai_envoy \
    --ip "172.10.0.22" \
    --log-opt mode=non-blocking \
    --log-opt max-buffer-size=5m \
    --log-opt max-size=100m \
    --log-opt max-file=5 \
    alantai/prj-envoy-v1/services/api-v1:api-v0.0.0
# \run api service
```

### Testing
Since the error handlers have not been completely implemented yet, the invalid requests would break the api applications.

1. Test API service and Nginx
```sh
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

# post data to mongo and get data from mongo by running the commands below
$ curl -k -d "userName=alan" -X POST https://atai-envoy.com/api/v1/visitor
$ curl -k https://atai-envoy.com/api/v1/visitor

# TODO: test websocket by websocat or Chrome

# test nginx servers by running the command below or just visit the website from Chrome. If you encounter a warning related to self-signed certificate, type "thisisunsafe".
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

2. Test Envoy admin by visiting http://0.0.0.0:8001


## References
* Golang
Ref:
- [Golang code review](https://github.com/golang/go/wiki/CodeReviewComments)
- [Scheduling in Golang](https://www.ardanlabs.com/blog/2018/08/scheduling-in-go-part1.html)

* Docker
Ref:
- [Container Networking From Scratch](https://youtu.be/6v_BDHIgOY8)
- [CNI/CNM - Introducing Container Networking](https://youtu.be/QMNbgmxmB-M)

```sh
# remove images with tag <none>
$ docker rmi $(docker images --filter "dangling=true")
```

* Kubernetes
- [Kubernetes and Networks](https://youtu.be/GgCA2USI5iQ)
- [Communication Is Key - Understanding Kubernetes Networking](https://youtu.be/InZVNuKY5GY)
- [Kubernetes 的 Go 微服务实践](https://www.infoq.cn/article/gXEQ7HaBkoujLF1RRj7C)

* Build Angular by Bazel
Ref:
- [Angular 8 with Bazel](https://blog.bitsrc.io/angular-8-bazel-walkthrough-f7585bcaf282)

* MongoDB
Ref:
- [User guide of Mongo 4.0](https://docs.mongodb.com/v4.0/reference/)

