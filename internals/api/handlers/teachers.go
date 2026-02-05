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

func (s *Server) GetTeachers(ctx context.Context, req *pb.GetTeacherRequset) (*pb.Teachers, error) {

	// filtering, getting filter from the requst, another function

	filter, err := buildfilter(req.Teacher, &models.Teacher{})
	if err != nil {
		return nil, utils.ErrorHandler(err, "internal err")
	}
	// sorting, getting sord option from the requst, another function
	sortOption := buildSortOptions(req.GetSortBy())
	//access the database to fetch data, another function

	teachers, err := repositories.GetTeachersDBhandler(ctx, sortOption, filter)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	return &pb.Teachers{Teachers: teachers}, nil
}

func (s *Server) UpdateTeachers(ctx context.Context, req *pb.Teachers) (*pb.Teachers, error) {
	updatedTeachers, err := repositories.UpdateTeachersDBHandler(ctx, req.Teachers)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.Teachers{Teachers: updatedTeachers}, nil
}
