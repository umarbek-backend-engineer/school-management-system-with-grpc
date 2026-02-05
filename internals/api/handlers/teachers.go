package handlers

import (
	"context"

	"school_project_grpc/internals/models"
	"school_project_grpc/internals/repositories"
	"school_project_grpc/internals/repositories/mongodb"
	"school_project_grpc/pkg/utils"
	pb "school_project_grpc/proto/gen"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
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

	teachers, err := GetTeachersDBhandler(ctx, sortOption, filter)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	return &pb.Teachers{Teachers: teachers}, nil
}

func GetTeachersDBhandler(ctx context.Context, sortOption bson.D, filter bson.M) ([]*pb.Teacher, error) {
	client, err := mongodb.CreatMongoClient()
	if err != nil {
		return nil, utils.ErrorHandler(err, "Internal error")
	}
	defer client.Disconnect(ctx)

	coll := client.Database("school").Collection("teachers")

	var cursor *mongo.Cursor
	if len(sortOption) < 1 {
		cursor, err = coll.Find(ctx, filter)
	} else {
		cursor, err = coll.Find(ctx, filter, options.Find().SetSort(sortOption))
	}

	if err != nil {
		return nil, utils.ErrorHandler(err, "Internal error")
	}
	defer cursor.Close(ctx)

	var teachers []*pb.Teacher
	for cursor.Next(ctx) {
		var teacher models.Teacher
		err = cursor.Decode(&teacher)
		if err != nil {
			return nil, utils.ErrorHandler(err, "Internal error")
		}
		teachers = append(teachers, &pb.Teacher{
			Id:        teacher.Id,
			FirstName: teacher.FirstName,
			LastName:  teacher.LastName,
			Email:     teacher.Email,
			Class:     teacher.Class,
			Subject:   teacher.Subject,
		})
	}
	return teachers, nil
}
