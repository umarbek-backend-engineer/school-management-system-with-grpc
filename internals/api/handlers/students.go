package handlers

import (
	"context"
	"school_project_grpc/internals/models"
	"school_project_grpc/internals/repositories"
	"school_project_grpc/pkg/utils"
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

func (s *Server) GetStudents(ctx context.Context, req *pb.GetStudentRequset) (*pb.Students, error) {

	// build filters
	filter, err := buildfilter(req.Student, &models.Student{})
	if err != nil {
		return nil, utils.ErrorHandler(err, "Internal error")
	}

	// build sortoptions
	sortOptions := buildSortOptions(req.GetSortBy())

	// fetch data from data base

	pageNumber := req.GetPageNum()
	pageSize := req.GetPageSize()

	if pageNumber < 1 {
		pageNumber = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}

	students, err := repositories.GetStudentsDBHandler(ctx, sortOptions, filter, pageSize, pageNumber)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.Students{Students: students}, nil
}
