package handlers

import (
	pb "school_project_grpc/proto/gen"
)

type Server struct {
	pb.UnimplementedExecsServiceServer
	pb.UnimplementedStudentsServiceServer
	pb.UnimplementedTeachersServiceServer
}
