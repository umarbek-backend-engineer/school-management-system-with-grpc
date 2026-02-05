package main

import (
	"log"
	"net"
	"os"

	"school_project_grpc/internals/api/handlers"
	"school_project_grpc/internals/repositories/mongodb"
	pb "school_project_grpc/proto/gen"

	"github.com/joho/godotenv"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	_, err := mongodb.CreatMongoClient()
	if err != nil {
		log.Println("Failed to connect mongoDB: ", err)
		return
	}
	log.Println("🎉 mongoDB connected successfully")

	err = godotenv.Load("./cmd/grpcapi/.env")
	if err != nil {
		log.Fatal("Failed to load .env file: ", err)
	}

	log.Println("🎉 .env file successfully loaded")

	port := os.Getenv("GRPC_SERVER_PORT")

	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatal("Failed to make listerer: ", err)
	}
	defer lis.Close()

	grpcServer := grpc.NewServer()
	pb.RegisterExecsServiceServer(grpcServer, &handlers.Server{})
	pb.RegisterStudentsServiceServer(grpcServer, &handlers.Server{})
	pb.RegisterTeachersServiceServer(grpcServer, &handlers.Server{})

	reflection.Register(grpcServer)

	log.Println("Server is running on port", port)
	log.Println("--------------------------------------")
	log.Print("--------------------------------------\n\n")

	err = grpcServer.Serve(lis)
	if err != nil {
		log.Fatal("Failed to run the grpc server")
	}
}
