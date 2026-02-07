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

func (s *Server) UpdateExecs(ctx context.Context, req *pb.Execs) (*pb.Execs, error) {
	execs, err := repositories.UpdateExecsDBHandler(ctx, req.Execs)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.Execs{Execs: execs}, nil
}

func (s *Server) DeleteExecs(ctx context.Context, req *pb.ExecIds) (*pb.DeleteExecsConfirm, error) {

	deletedIds, err := repositories.DeleteExecsDBHandler(ctx, req.GetExecIds())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	return &pb.DeleteExecsConfirm{
		Status:     "Execs successfully deleted",
		DeletedIds: deletedIds,
	}, nil
}

// login function
func (s *Server) Login(ctx context.Context, req *pb.ExecLogInRequest) (*pb.ExecLogInResponse, error) {

	// data base handler
	exec, err := repositories.LoginDBHandler(ctx, req.GetUsername())
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	// checking if the user is active or not if not active user will not be acceptable
	if exec.InactiveStatus {
		return nil, status.Error(codes.Unauthenticated, "Account is Inactive")
	}

	// verify password
	err = utils.VerifyPassword(req.GetPassword(), exec.Password)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "Incorrect Password")
	}

	// signing jwt
	token, err := utils.SingingJWT(exec.Id, exec.Username, exec.Role)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "Failed to created jwt token")
	}

	return &pb.ExecLogInResponse{
		Status: true,
		Token:  token,
	}, nil
}

// function to update the user password
func (s *Server) UpdatePassword(ctx context.Context, req *pb.UpdatePasswordRequest) (*pb.UpdatePasswordResponse, error) {

	// update password db operations
	user, err := repositories.UpdatePasswordDBHandler(ctx, req)
	if err != nil {
		return nil, err
	}

	// signing token
	token, err := utils.SingingJWT(user.Id, user.Username, user.Role)
	if err != nil {
		return nil, utils.ErrorHandler(err, "Internal error")
	}

	// giving response to the user
	return &pb.UpdatePasswordResponse{
		PasswordUpdated: true,
		Token:           token,
	}, nil
}

func (s *Server) DeactivateUser(ctx context.Context, req *pb.ExecIds) (*pb.Confirmation, error) {
	res, err := repositories.DeactivateUserDBHandler(ctx, req.GetExecIds())
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.Confirmation{
		Confirmation: res.ModifiedCount > 0,
	}, nil
}

func (s *Server) ReactivateUser(ctx context.Context, req *pb.ExecIds) (*pb.Confirmation, error) {
	res, err := repositories.ReactivateUserDBHandler(ctx, req.GetExecIds())
	if err != nil {
		return nil, err
	}

	return &pb.Confirmation{
		Confirmation: res.ModifiedCount > 0,
	}, nil
}
