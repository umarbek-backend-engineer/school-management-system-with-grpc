package main

import (
	"log"
	"net"
	"os"
	"time"

	"school_project_grpc/internals/api/handlers"
	itc "school_project_grpc/internals/api/interceptors"
	"school_project_grpc/pkg/utils"
	pb "school_project_grpc/proto/gen"

	"github.com/joho/godotenv"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {

	//loading  the /env file
	err := godotenv.Load("./cmd/grpcapi/.env")
	if err != nil {
		log.Fatal("Failed to load .env file: ", err)
	}

	log.Println("🎉 .env file successfully loaded")

	port := os.Getenv("GRPC_SERVER_PORT")
	// cert := os.Getenv("CERT_FILE")
	// key := os.Getenv("KEY_FILE")

	// creds, err := credentials.NewClientTLSFromFile(cert, key)
	// if err != nil {
	// 	log.Fatalf("Failed to load TLS cert files")
	// }

	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatal("Failed to make listerer: ", err)
	}
	defer lis.Close()

	// registering grpcServer, this is essential to run the server
	// grpcServer := grpc.NewServer(grpc.ChainUnaryInterceptor(itc.NewRateLimiter(20, time.Second*10).RateLimitIntercepter, itc.ResponseTimeIntercepter, itc.Authentication_Intercepter), grpc.Creds(creds))
	grpcServer := grpc.NewServer(grpc.ChainUnaryInterceptor(itc.NewRateLimiter(20, time.Second*10).RateLimitIntercepter, itc.ResponseTimeIntercepter, itc.Authentication_Intercepter))

	// registering rpcs
	pb.RegisterExecsServiceServer(grpcServer, &handlers.Server{})
	pb.RegisterStudentsServiceServer(grpcServer, &handlers.Server{})
	pb.RegisterTeachersServiceServer(grpcServer, &handlers.Server{})

	// this function is responsible to skip the proto file when testing in postman, it is only used in production period to test
	reflection.Register(grpcServer)

	go utils.JwtStore.CleanUpExpiredTokens()

	log.Println("Server is running on port", port)
	log.Println("--------------------------------------")
	log.Print("--------------------------------------\n\n")

	// running the server
	err = grpcServer.Serve(lis)
	if err != nil {
		log.Fatal("Failed to run the grpc server")
	}
}
