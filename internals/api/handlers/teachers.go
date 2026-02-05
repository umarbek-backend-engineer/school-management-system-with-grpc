package handlers

import (
	"context"

	"school_project_grpc/internals/repositories"
	pb "school_project_grpc/proto/gen"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *Server) AddTeachers(ctx context.Context, req *pb.Teachers) (*pb.Teachers, error) {

	for _, teacher := range req.Teachers {
		if teacher.Id != "" {
			return nil, status.Error(codes.InvalidArgument, "request is incorrect format: nun-empty field ID fields are not allowed.")
		}
	}

	addedTeacher, err := repositories.AddTeachersDBHandler(ctx, req.GetTeachers())
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.Teachers{Teachers: addedTeacher}, nil
}
