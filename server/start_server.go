package server

import (
	"context"
	"fmt"
	"net"

	lo "main/libreoffice"
	proto "main/proto/gobre"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
)

type GobreServer struct {
	proto.UnimplementedGobreServer
}

func (s GobreServer) HandleFileRequest(
	ctx context.Context,
	param *proto.FileRequest,
) (*proto.FileResponse, error) {
	fmt.Println(
		"Received file conversion request ",
		param.OriginalFileType,
		" to ",
		param.NewFileType,
	)
	fileData, errors := lo.HandleConvertFile(
		param.OriginalFileType,
		param.NewFileType,
		param.FileData,
	)

	return &proto.FileResponse{FileData: fileData}, errors
}

func StartServer(ctx context.Context) {
	fmt.Println("Starting gRPC server on port :8081")
	listener, listenError := net.Listen("tcp", ":8081")
	if listenError != nil {
		fmt.Println("Server startup error: ", listenError)
		panic(listenError)
	}

	server := grpc.NewServer()
	reflection.Register(server) //Enabled for clients that support reflection

	grpc_health_v1.RegisterHealthServer(server, health.NewServer())

	proto.RegisterGobreServer(server, GobreServer{})

	//Start the server in a separate goroutine
	go func() {
		fmt.Println("gRPC server is running...")
		<-ctx.Done()
		fmt.Println("Shutting down gRPC server...")
		server.Stop()
		fmt.Println("gRPC server shut down complete.")
	}()

	//Serve the server
	serverError := server.Serve(listener)

	if serverError != nil {
		fmt.Println("gRPC server error: ", serverError)
		panic(serverError)
	}
}
