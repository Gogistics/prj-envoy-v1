package grpchandlers

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"io"
	"io/ioutil"
	"log"
	"sync"
	"time"

	"github.com/Gogistics/prj-envoy-v1/services/grpc-v1/protos"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
)

type gRPCAPIClient struct {
	serverAddr      string
	caCert          string
	skipHealthCheck bool
	serverName      string
}
type GRPCClientWrapper struct {
	Client *gRPCAPIClient
}

var (
	GRPCWrapper = GRPCClientWrapper{
		Client: &(gRPCAPIClient{
			serverAddr:      "172.10.0.200:443",
			skipHealthCheck: true,
			caCert:          "certs/atai-envoy.com.crt", // grpc need atai-envoy.com.crt
			serverName:      "atai-envoy.com",
		})}
)

func (gac *gRPCAPIClient) runUnary(client protos.ServiceIPMappingClient) {
	log.Println("run unary call...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	clusterInfo := protos.ClusterInfo{}
	clusterInfo.Type = "internal"
	clusterInfo.Group = "db"
	resp, err := client.QueryServiceUnary(ctx, &protos.ServiceRequest{Cluster: &clusterInfo})
	if err != nil {
		log.Fatalln("Error: failed to query cluster info.")
	}
	log.Printf("Unary response => type: %s ; group: %s ; name: %s ; ip: %s", resp.Cluster.Type, resp.Cluster.Group, resp.Name, resp.Ip)
}

func (gac *gRPCAPIClient) runClientStream(client protos.ServiceIPMappingClient) {
	log.Println("run client stream call...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	clientStream, err := client.QueryServiceClientStream(ctx)
	if err != nil {
		log.Fatalln("Error: failed to set up client stream. ", err)
	}

	// TODO: replace for loop
	for ith := 1; ith < 7; ith++ {
		clusterInfo := protos.ClusterInfo{}
		clusterInfo.Type = "internal"
		clusterInfo.Group = "db"
		if err := clientStream.Send(&protos.ServiceRequest{Cluster: &clusterInfo}); err != nil {
			if err == io.EOF {
				break
			}
			log.Fatalf("Error: %v failed to send request. %v", clientStream, err)
		}
	}

	clientReply, err := clientStream.CloseAndRecv()
	if err != nil {
		log.Fatalf("Error: %v CloseAndRecv() received an error %v", clientStream, err)
	} else {
		log.Printf("Received response of QueryServiceClientStream => type: %s ; group: %s", clientReply.Cluster.Type, clientReply.Cluster.Group)
	}
}

func (gac *gRPCAPIClient) runServerStream(client protos.ServiceIPMappingClient) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	clusterInfo := protos.ClusterInfo{}
	clusterInfo.Type = "internal"
	clusterInfo.Group = "db"
	serverStream, err := client.QueryServiceServerStream(ctx, &protos.ServiceRequest{Cluster: &clusterInfo})
	if err != nil {
		log.Fatalln("Error: failed to query through server stream")
	}
	for {
		service, err := serverStream.Recv()
		if err != nil {
			if err == io.EOF {
				trailer := serverStream.Trailer()
				log.Println("Server stream trailer: ", trailer)
				break
			}
			log.Fatalln("Error: QueryServiceServerStream failed to receive service; ", err)
		} else {
			log.Printf("Service => type: %s ; group: %s", service.Cluster.Type, service.Cluster.Group)
		}
	}
}

func (gac *gRPCAPIClient) runBiStream(client protos.ServiceIPMappingClient) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	done := make(chan bool)
	biStream, err := client.QueryServiceBiStream(ctx)
	if err != nil {
		log.Fatalln("Error: failed to open stream; ", err)
	}
	ctxOfBiStream := biStream.Context()
	var wg sync.WaitGroup

	go func() {
		for ith := 1; ith < 7; ith++ {
			wg.Add(1)

			go func(count int) {
				defer wg.Done()
				clusterInfo := protos.ClusterInfo{}
				clusterInfo.Type = "internal"
				clusterInfo.Group = "db"
				req := protos.ServiceRequest{Cluster: &clusterInfo}
				if err := biStream.SendMsg(&req); err != nil {
					log.Fatalln("Error: failed to send message; ", err)
				}
			}(ith)
		}

		wg.Wait()
		if err := biStream.CloseSend(); err != nil {
			log.Println("Failed to close", err)
		}
	}()

	go func() {
		for {
			resp, err := biStream.Recv()
			if err != nil {
				if err == io.EOF {
					close(done)
					return
				}
				log.Fatalln("Error: failed to receive response; ", err)
			}
			log.Printf("Response: type: %s ; group: %s ; name: %s ; ip: %s", resp.Cluster.Type, resp.Cluster.Group, resp.Name, resp.Ip)
		}
	}()

	go func() {
		<-ctxOfBiStream.Done()
		if err := ctxOfBiStream.Err(); err != nil {
			log.Println("Context err: ", err)
		}
	}()

	// Note: check if done has been closed
	_, ok := <-done
	if ok {
		close(done)
	}
}

func (wrapper *GRPCClientWrapper) RunGRPCCalls() {
	// set up connection to the server
	var err error
	var opts []grpc.DialOption

	// set tls
	var tlsCfg tls.Config
	rootCAs := x509.NewCertPool()
	pem, err := ioutil.ReadFile(wrapper.Client.caCert)
	if err != nil {
		log.Fatalf("failed to load root CA certificates  error=%v", err)
	}
	if !rootCAs.AppendCertsFromPEM(pem) {
		log.Fatalf("no root CA certs parsed from file ")
	}
	tlsCfg.RootCAs = rootCAs
	tlsCfg.ServerName = wrapper.Client.serverName
	creds := credentials.NewTLS(&tlsCfg)

	opts = append(opts, grpc.WithTransportCredentials(creds))
	opts = append(opts, grpc.WithBlock())
	log.Println("Dialing RPC server...")
	conn, err := grpc.Dial(wrapper.Client.serverAddr, opts...)
	if err != nil {
		log.Fatalf("fail to dial: %v", err)
	}
	defer conn.Close()

	client := protos.NewServiceIPMappingClient(conn)
	// runUnary
	wrapper.Client.runUnary(client)

	// Client stream
	wrapper.Client.runClientStream(client)

	// Server stream
	wrapper.Client.runServerStream(client)

	// bi-directional stream
	wrapper.Client.runBiStream(client)

	// perform healthcheck request
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// healthcheck is optional
	if !wrapper.Client.skipHealthCheck {
		resp, err := healthpb.NewHealthClient(conn).Check(ctx, &healthpb.HealthCheckRequest{Service: "protos.ServiceIPMapping"})
		if err != nil {
			log.Fatalln("Error: HealthCheck failed ", err)
		}
		if resp.GetStatus() != healthpb.HealthCheckResponse_SERVING {
			log.Fatalln("Error: service is not serving, ", resp.GetStatus().String())
		}
		log.Println("gRPC HealthCheckStatus: ", resp.GetStatus())
	}
}
