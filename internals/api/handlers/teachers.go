package handlers

import (
	"context"
	"fmt"

	"school_project_grpc/internals/models"
	"school_project_grpc/internals/repositories"
	"school_project_grpc/internals/repositories/mongodb"
	"school_project_grpc/pkg/utils"
	pb "school_project_grpc/proto/gen"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Add teachers
func (s *Server) AddTeachers(ctx context.Context, req *pb.Teachers) (*pb.Teachers, error) {

	// Validate: ID must be empty on create
	for _, teacher := range req.Teachers {
		if teacher.Id != "" {
			return nil, status.Error(codes.InvalidArgument,
				"request is incorrect format: non-empty ID fields are not allowed.")
		}
	}

	addedTeacher, err := repositories.AddTeachersDBHandler(ctx, req.GetTeachers())
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.Teachers{Teachers: addedTeacher}, nil
}

// Get teachers with filter + sort
func (s *Server) GetTeachers(ctx context.Context, req *pb.GetTeacherRequset) (*pb.Teachers, error) {

	// Build Mongo filter from request
	filter, err := buildfilter(req.Teacher, &models.Teacher{})
	if err != nil {
		return nil, utils.ErrorHandler(err, "internal err")
	}

	// Build sort options from request
	sortOption := buildSortOptions(req.GetSortBy())

	pageNumber := req.GetPageNum()
	pageSize := req.GetPageSize()

	if pageNumber < 1 {
		pageNumber = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}

	// Fetch from database
	teachers, err := repositories.GetTeachersDBhandler(ctx, sortOption, filter, pageSize, pageNumber)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	return &pb.Teachers{Teachers: teachers}, nil
}

// Update teachers
func (s *Server) UpdateTeachers(ctx context.Context, req *pb.Teachers) (*pb.Teachers, error) {
	updatedTeachers, err := repositories.UpdateTeachersDBHandler(ctx, req.Teachers)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.Teachers{Teachers: updatedTeachers}, nil
}

// Delete teachers by IDs
func (s *Server) DeleteTeacher(ctx context.Context, req *pb.TeacherIds) (*pb.DeleteTeacherConfirm, error) {

	ids := req.TeacherIds
	var teacherIDsTODelete []string

	// Collect string IDs
	for _, v := range ids {
		teacherIDsTODelete = append(teacherIDsTODelete, v.Id)
	}

	client, err := mongodb.CreatMongoClient()
	if err != nil {
		return nil, utils.ErrorHandler(err, "Internal error")
	}
	defer client.Disconnect(ctx)

	// Convert to Mongo ObjectIDs
	objectIds := make([]primitive.ObjectID, 0, len(teacherIDsTODelete))
	for _, id := range teacherIDsTODelete {
		objectId, err := primitive.ObjectIDFromHex(id)
		if err != nil {
			return nil, utils.ErrorHandler(err, fmt.Sprintf("Invalid id: %v", id))
		}
		objectIds = append(objectIds, objectId)
	}

	// Delete many by IDs
	filter := bson.M{"_id": bson.M{"$in": objectIds}}

	res, err := client.Database("school").Collection("teachers").DeleteMany(ctx, filter)
	if err != nil {
		return nil, utils.ErrorHandler(err, "Internal error")
	}

	if res.DeletedCount == 0 {
		return nil, utils.ErrorHandler(err, "No teachers were deleted")
	}

	// Return deleted IDs
	deletedIds := make([]string, 0, len(objectIds))
	for _, v := range objectIds {
		deletedIds = append(deletedIds, v.Hex())
	}

	return &pb.DeleteTeacherConfirm{
		Status:     "Teacher successfully deleted",
		DeletedIds: deletedIds,
	}, nil
}
