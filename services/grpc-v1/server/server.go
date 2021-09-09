package main

import (
	"bytes"
	"context"
	"flag"
	"io"
	"log"
	"net"
	"time"

	"github.com/Gogistics/prj-envoy-v1/services/grpc-v1/protos"
	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/status"
)

type healthServer struct{}
type grpcServer struct{}

var (
	// tls      = flag.Bool("tls", false, "Connection uses TLS if true, else plain TCP")
	certFile = flag.String("certFile", "", "The TLS cert file")
	keyFile  = flag.String("keyFile", "", "The TLS key file")
	// jsonFile = flag.String("json_file", "", "A json file containing a list of services")
	port = flag.String("port", ":20000", "The server port")
)

// ref: https://pkg.go.dev/google.golang.org/grpc/health/grpc_health_v1
func (hServer *healthServer) Check(ctx context.Context, hcRequest *healthpb.HealthCheckRequest) (*healthpb.HealthCheckResponse, error) {
	log.Println("Handling grpc check request: ", hcRequest.Service)
	return &healthpb.HealthCheckResponse{Status: healthpb.HealthCheckResponse_SERVING}, nil
}

func (hServer *healthServer) Watch(hcRequest *healthpb.HealthCheckRequest, hwServer healthpb.Health_WatchServer) error {
	return status.Error(codes.Unimplemented, "Watch is not implemented")
}

func (rpcServer *grpcServer) QueryServiceUnary(ctx context.Context, srvQuery *protos.ServiceRequest) (*protos.Service, error) {
	uid, _ := uuid.NewUUID()
	log.Println("Received unary request: ", uid)
	// TODO: replace hardcode values with values from JSON
	clusterInfo := protos.ClusterInfo{}
	clusterInfo.Type = "internal"
	clusterInfo.Group = "db"
	return &protos.Service{Cluster: &clusterInfo, Name: "hello-grpc", Ip: "172.11.0.11"}, nil
}

func (rpcServer *grpcServer) QueryServiceServerStream(srvQuery *protos.ServiceRequest, stream protos.ServiceIPMapping_QueryServiceServerStreamServer) error {
	log.Println("Received QueryServiceServerStream request: ")
	// TODO: send data from JSON
	for ith := 0; ith < 7; ith++ {
		uid, _ := uuid.NewUUID()
		var currType string
		if ith%2 == 0 {
			currType = "internal"
		} else {
			currType = "external"
		}
		var strBytes bytes.Buffer
		strBytes.WriteString("group")
		strBytes.WriteString(uid.String())
		currGroup := strBytes.String()
		clusterInfo := protos.ClusterInfo{}
		clusterInfo.Type = currType
		clusterInfo.Group = currGroup
		stream.Send(&protos.Service{Cluster: &clusterInfo, Name: uid.String(), Ip: "172.10.0.200"})
	}
	return nil
}

func (rpcServer *grpcServer) QueryServiceClientStream(stream protos.ServiceIPMapping_QueryServiceClientStreamServer) error {
	// TODO: replace hardcode with test data
	for {
		req, err := stream.Recv()
		if err == io.EOF {
			clusterInfo := protos.ClusterInfo{}
			clusterInfo.Type = "NA"
			clusterInfo.Group = "NA"
			return stream.SendAndClose(&protos.Service{Cluster: &clusterInfo, Name: "NA", Ip: "NA"})
		} else if err != nil {
			log.Println(err)
			return err
		}
		log.Println("Received QueryServiceClientStream request: ", req.String())
	}
}

func (rpcServer *grpcServer) QueryServiceBiStream(srvServer protos.ServiceIPMapping_QueryServiceBiStreamServer) error {
	for {
		req, err := srvServer.Recv()
		if err == io.EOF {
			return nil
		} else if err != nil {
			log.Println("Error occured! ", err)
			continue
		} else {
			log.Println("Received QueryServiceBiStream ", req.String())
			clusterInfo := protos.ClusterInfo{}
			clusterInfo.Type = "NA"
			clusterInfo.Group = "NA"
			resp := &protos.Service{Cluster: &clusterInfo, Name: "NA", Ip: "NA"}
			if err := srvServer.Send(resp); err != nil {
				log.Println("Error: failed to send response to client. ", err)
			}
		}
	}
}
func main() {
	flag.Parse()
	if *port == "" {
		flag.Usage()
		log.Fatalln("Missing -port flag (:20000)")
	}

	lis, err := net.Listen("tcp", *port)
	if err != nil {
		log.Fatalf("Error: failed to listen: %v", err)
	}

	var serverOptions []grpc.ServerOption

	if *certFile == "" {
		*certFile = "atai-envoy.com.crt"
	}
	if *keyFile == "" {
		*keyFile = "atai-envoy.com.key"
	}
	creds, err := credentials.NewServerTLSFromFile(*certFile, *keyFile)
	if err != nil {
		log.Fatalf("Failed to generate credentials %v", err)
	}

	/* Config. guide
	ref:
	- https://github.com/grpc/grpc-go/blob/v1.40.0/server.go#L143
	- https://lukexng.medium.com/grpc-keepalive-maxconnectionage-maxconnectionagegrace-6352909c57b8
	*/
	serverOptions = []grpc.ServerOption{
		grpc.Creds(creds),
		grpc.NumStreamWorkers(10),
		grpc.ConnectionTimeout(30 * time.Second),
		grpc.KeepaliveEnforcementPolicy(keepalive.EnforcementPolicy{
			MinTime:             5 * time.Second,
			PermitWithoutStream: true,
		}),
		grpc.KeepaliveParams(keepalive.ServerParameters{
			MaxConnectionAge:      10 * time.Second,
			MaxConnectionAgeGrace: 30 * time.Second}),
		grpc.MaxHeaderListSize(10240)}

	rpcServer := grpc.NewServer(serverOptions...)
	protos.RegisterServiceIPMappingServer(rpcServer, &grpcServer{})
	healthpb.RegisterHealthServer(rpcServer, &healthServer{})
	log.Println("Starting server ...")

	if err := rpcServer.Serve(lis); err != nil {
		log.Fatalf("Error: failed to serve: %v", err)
	}
}
