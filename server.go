package main

import (
	"log"
	"net"
	"net/http"
	"strings"
	"time"

	"context"

	"github.com/improbable-eng/grpc-web/go/grpcweb"
	echo "github.com/quanghung97/grpcexample/echo"
	pingpong "github.com/quanghung97/grpcexample/pingpong"
	pingpong2 "github.com/quanghung97/grpcexample/pingpong2"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

type Server struct {
	pingpong.PingPongServer
	echo.EchoServiceServer
}

type Server2 struct {
	pingpong2.PingPong2Server
}

type grpcMultiplexer struct {
	wrap1 *grpcweb.WrappedGrpcServer
	wrap2 *grpcweb.WrappedGrpcServer
}

// Ping fullfills the requirement for PingPong Server interface
func (s *Server) Ping(ctx context.Context, ping *pingpong.PingRequest) (*pingpong.PongResponse, error) {
	return &pingpong.PongResponse{
		Ok: true,
	}, nil
}

// Echo fullfills the requirement for PingPong Server interface
func (s *Server) Echo(ctx context.Context, ping *echo.EchoRequest) (*echo.EchoResponse, error) {
	return &echo.EchoResponse{
		Ok: true,
	}, nil
}

// Ping fullfills the requirement for PingPong2 Server2 interface
func (s *Server2) Ping2(ctx context.Context, ping *pingpong2.Ping2Request) (*pingpong2.Pong2Response, error) {
	return &pingpong2.Pong2Response{
		Ok: true,
	}, nil
}

// Handler is used to route requests to either grpc or to regular http
func (m *grpcMultiplexer) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if m.wrap1.IsGrpcWebRequest(r) && strings.Split(strings.Split(r.URL.Path, "/")[1], ".")[0] == "service" {
			m.wrap1.ServeHTTP(w, r)
			return
		}

		if m.wrap2.IsGrpcWebRequest(r) && strings.Split(strings.Split(r.URL.Path, "/")[1], ".")[0] == "service2" {
			m.wrap2.ServeHTTP(w, r)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// GenerateTLSApi will load TLS certificates and key and create a grpc server with those.
func GenerateTLSApi(pemPath, keyPath string) (*grpc.Server, error) {
	cred, err := credentials.NewServerTLSFromFile(pemPath, keyPath)
	if err != nil {
		return nil, err
	}

	s := grpc.NewServer(
		grpc.Creds(cred),
	)
	return s, nil
}

func main() {

	// We Generate a TLS grpc API
	// certFile := "/go/src/ssl/server-cert.pem"
	// keyFile := "/go/src/ssl/server-key.pem"
	apiserver, err := GenerateTLSApi("ssl/server-cert.pem", "ssl/server-key.pem")

	if err != nil {
		log.Fatal(err)
	}
	apiserver2, err2 := GenerateTLSApi("ssl/server-cert.pem", "ssl/server-key.pem")
	if err2 != nil {
		log.Fatal(err2)
	}

	// Start listening on a TCP Port
	lis, err := net.Listen("tcp", "127.0.0.1:9990")
	if err != nil {
		log.Fatal(err)
	}
	lis2, err2 := net.Listen("tcp", "127.0.0.1:9992")
	if err2 != nil {
		log.Fatal(err2)
	}
	// We need to tell the code WHAT TO do on each request, ie. The business logic.
	// In GRPC cases, the Server is acutally just an Interface
	// So we need a struct which fulfills the server interface
	// see server.go
	s := &Server{}
	s2 := &Server2{}
	// Register the API server as a PingPong Server
	// The register function is a generated piece by protoc.
	pingpong.RegisterPingPongServer(apiserver, s)
	echo.RegisterEchoServiceServer(apiserver, s)

	pingpong2.RegisterPingPong2Server(apiserver2, s2)
	// Start serving in a goroutine to not block
	go func() {
		log.Fatal(apiserver.Serve(lis), 1111111)
	}()

	go func() {
		log.Fatal(apiserver2.Serve(lis2), 2222222)
	}()
	// Wrap the GRPC Server in grpc-web and also host the UI
	grpcWebServer := grpcweb.WrapServer(apiserver)
	grpcWebServer2 := grpcweb.WrapServer(apiserver2)
	// Lets put the wrapped grpc server in our multiplexer struct so
	// it can reach the grpc server in its handler
	multiplex := grpcMultiplexer{
		grpcWebServer,
		grpcWebServer2,
	}

	// We need a http router
	r := http.NewServeMux()
	// Load the static webpage with a http fileserver
	webapp := http.FileServer(http.Dir("ui/pingpongapp/build"))
	// Host the Web Application at /, and wrap it in the GRPC Multiplexer
	// This allows grpc requests to transfer over HTTP1. then be
	// routed by the multiplexer
	r.Handle("/", multiplex.Handler(webapp))
	// Create a HTTP server and bind the router to it, and set wanted address
	srv := &http.Server{
		Handler:      r,
		Addr:         "localhost:8080",
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}
	// Serve the webapp over TLS
	log.Fatal(srv.ListenAndServeTLS("ssl/server-cert.pem", "ssl/server-key.pem"))
}
