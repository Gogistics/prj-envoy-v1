# gRPC

```sh
# generate *pb.go file for local development; *.pb.go has been generated at ./bazel-bin/services/grpc-v1/protos/protos_go_proto_/github.com/Gogistics/prj-envoy-v1/services/grpc-v1/protos/service_ip_mapping.pb.go
$ bazel build //services/grpc-v1/protos:protos

# build images
$ bazel build --platforms=@io_bazel_rules_go//go/toolchain:linux_amd64 \
    //services/grpc-v1/server:all
$ bazel run --platforms=@io_bazel_rules_go//go/toolchain:linux_amd64 \
    //services/grpc-v1/server:grpc-query-server-v0.0.0

$ bazel build --platforms=@io_bazel_rules_go//go/toolchain:linux_amd64 \
    //services/grpc-v1/client:all
$ bazel run --platforms=@io_bazel_rules_go//go/toolchain:linux_amd64 \
    //services/grpc-v1/client:grpc-query-client-v0.0.0

# run containers
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
```

```sh
# 1. create grpc bridge proxy
$ docker run -d \
      --name atai_envoy_grpc_client \
      --network atai_envoy \
      --ip "172.10.0.200" \
      --log-opt mode=non-blocking \
      --log-opt max-buffer-size=5m \
      --log-opt max-size=100m \
      --log-opt max-file=5 \
      alantai/prj-envoy-v1/envoys:grpc-client-envoy-v0.0.0

# Or 2. create a container and run it later
$ docker create -it \
      --name atai_envoy_grpc_client \
      --network atai_envoy \
      --ip "172.10.0.200" \
      --log-opt mode=non-blocking \
      --log-opt max-buffer-size=5m \
      --log-opt max-size=100m \
      --log-opt max-file=5 \
      alantai/prj-envoy-v1/envoys:grpc-client-envoy-v0.0.0

$ docker network connect atai_grpc atai_envoy_grpc_client

# for 2. only
$ docker start atai_envoy_grpc_client
```

## Referneces
- https://grpc.io/docs/languages/go/basics/
