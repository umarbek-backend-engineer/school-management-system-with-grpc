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

// Add execs
func (s *Server) AddExecs(ctx context.Context, req *pb.Execs) (*pb.Execs, error) {

	// Validate: ID must be empty on create
	for _, exec := range req.Execs {
		if exec.Id != "" {
			return nil, status.Error(codes.InvalidArgument,
				"request is incorrect format: non-empty ID fields are not allowed.")
		}
	}

	addedExec, err := repositories.AddExecsDBHandler(ctx, req.GetExecs())
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.Execs{Execs: addedExec}, nil
}

func (s *Server) GetExecs(ctx context.Context, req *pb.GetExecRequset) (*pb.Execs, error) {
	// build mongo filter from request

	filter, err := buildfilter(req.Exec, &models.Exec{})
	if err != nil {
		return nil, utils.ErrorHandler(err, "internal err")
	}
	// build sort options from the request
	sortOption := buildSortOptions(req.GetSortBy())
	// Fetch from db

	execs, err := repositories.GetExecsDBHandler(ctx, sortOption, filter)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.Execs{Execs: execs}, nil
}
