package handlers

import (
	"context"
	"school_project_grpc/internals/repositories"
	pb "school_project_grpc/proto/gen"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Add students
func (s *Server) AddStudents(ctx context.Context, req *pb.Students) (*pb.Students, error) {

	// Validate: ID must be empty on create
	for _, student := range req.Students {
		if student.Id != "" {
			return nil, status.Error(codes.InvalidArgument,
				"request is incorrect format: non-empty ID fields are not allowed.")
		}
	}

	addedStudent, err := repositories.AddStudentsDBHandler(ctx, req.GetStudents())
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.Students{Students: addedStudent}, nil
}
