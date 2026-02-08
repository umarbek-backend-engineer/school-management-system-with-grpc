package handlers

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"school_project_grpc/internals/models"
	"school_project_grpc/internals/repositories"
	"school_project_grpc/pkg/utils"
	pb "school_project_grpc/proto/gen"
	"strconv"
	"strings"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// Add execs
func (s *Server) AddExecs(ctx context.Context, req *pb.Execs) (*pb.Execs, error) {

	// authorization
	err := utils.Authorization(ctx, "admin", "manager")
	if err != nil {
		return nil, status.Error(codes.Unavailable, "user is not authorized for this function")
	}

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

	// authorization
	err := utils.Authorization(ctx, "admin", "manager")
	if err != nil {
		return nil, status.Error(codes.Unavailable, "user is not authorized for this function")
	}

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

	// authorization
	err := utils.Authorization(ctx, "admin", "manager")
	if err != nil {
		return nil, status.Error(codes.Unavailable, "user is not authorized for this function")
	}

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

	// authorization
	err := utils.Authorization(ctx, "admin", "manager")
	if err != nil {
		return nil, status.Error(codes.Unavailable, "user is not authorized for this function")
	}

	res, err := repositories.DeactivateUserDBHandler(ctx, req.GetExecIds())
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.Confirmation{
		Confirmation: res.ModifiedCount > 0,
	}, nil
}

func (s *Server) ReactivateUser(ctx context.Context, req *pb.ExecIds) (*pb.Confirmation, error) {

	// authorization
	err := utils.Authorization(ctx, "admin", "manager")
	if err != nil {
		return nil, status.Error(codes.Unavailable, "user is not authorized for this function")
	}

	res, err := repositories.ReactivateUserDBHandler(ctx, req.GetExecIds())
	if err != nil {
		return nil, err
	}

	return &pb.Confirmation{
		Confirmation: res.ModifiedCount > 0,
	}, nil
}

// forgot passwor handler, sends token to the user's email throught which user can reset password
func (s *Server) ForgotPassword(ctx context.Context, req *pb.ForgotPasswordRequst) (*pb.ForgotPasswordResponse, error) {
	email := req.GetEmail()

	// database operations
	err := repositories.ForgotPasswordDBHandler(ctx, email)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.ForgotPasswordResponse{
		Confirmation: true,
		Message:      fmt.Sprintf("Password Reset link was sent to %s", email),
	}, nil
}

func (s *Server) ResetPassword(ctx context.Context, req *pb.ResetPasswordRequst) (*pb.Confirmation, error) {
	token := req.GetResetCode()

	if req.NewPassword != req.ConfirmPassword {
		return nil, status.Error(codes.InvalidArgument, "passwords do not match")
	}

	// decoding the tokne to check it with db token
	bytes, err := hex.DecodeString(token)
	if err != nil {
		return nil, utils.ErrorHandler(err, "Internal error")
	}

	// encoding the byte token in db token encription style to compare them
	hashedToken := sha256.Sum256(bytes)
	hashedTokenString := hex.EncodeToString(hashedToken[:])

	err = repositories.ResetPasswordDBHandler(ctx, hashedTokenString, req.GetNewPassword())
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.Confirmation{
		Confirmation: true,
	}, nil

}

func (s *Server) Logout(ctx context.Context, req *pb.EmptyRequest) (*pb.ExecLogoutResponse, error) {
	metadata, ok := metadata.FromIncomingContext(ctx) // checking metadata from request
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "Missing metadata")
	}

	// retriving token form ctx
	val, ok := metadata["authorization"]
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "Missing token")
	}

	token := strings.TrimPrefix(val[0], "Bearer ")
	if token == "" {
		return nil, status.Error(codes.Unauthenticated, "Missing token")
	}

	// extracting expiration time
	expTimeStamp := ctx.Value("exp")
	expTimeString := fmt.Sprintf("%v", expTimeStamp)

	expTimeInt, err := strconv.ParseInt(expTimeString, 10, 64)
	if err != nil {
		utils.ErrorHandler(err, "")
		return nil, status.Error(codes.Internal, "Internal Error")
	}

	exptime := time.Unix(expTimeInt, 0)

	utils.JwtStore.AddToken(token, exptime)

	return &pb.ExecLogoutResponse{
		LoggedOut: true,
	}, nil

}
