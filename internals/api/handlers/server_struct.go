package handlers

import (
	pb "school_project_grpc/proto/gen"
)

// this is the server struct which is used to implement the rpc services

type Server struct {
	pb.UnimplementedExecsServiceServer
	pb.UnimplementedStudentsServiceServer
	pb.UnimplementedTeachersServiceServer
}
