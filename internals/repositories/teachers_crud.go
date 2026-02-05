package repositories

import (
	"context"
	"reflect"
	"school_project_grpc/internals/models"
	"school_project_grpc/internals/repositories/mongodb"
	"school_project_grpc/pkg/utils"
	pb "school_project_grpc/proto/gen"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

func AddTeachersDBHandler(ctx context.Context, teacherFromReq []*pb.Teacher) ([]*pb.Teacher, error) {
	db, err := mongodb.CreatMongoClient()
	if err != nil {
		utils.ErrorHandler(err, "internal error")
		return nil, err
	}
	defer db.Disconnect(ctx)

	newTeachers := make([]*models.Teacher, 0, len(teacherFromReq))
	for i, pbTeacher := range teacherFromReq {
		newTeachers[i] = mapPBTeacherToModelTeacher(pbTeacher)
	}

	var addedTeacher []*pb.Teacher

	for _, teacher := range newTeachers {

		if teacher == nil {
			continue
		}

		result, err := db.Database("school").Collection("teachers").InsertOne(ctx, teacher)
		if err != nil {
			return nil, utils.ErrorHandler(err, "Error adding value into database")
		}

		objectID, ok := result.InsertedID.(primitive.ObjectID)
		if ok {
			teacher.Id = objectID.Hex()
		}

		pbTeacher := mapModelTeacherToPb(teacher)

		addedTeacher = append(addedTeacher, pbTeacher)
	}
	return addedTeacher, nil
}

func mapModelTeacherToPb(teacher *models.Teacher) *pb.Teacher {
	pbTeacher := &pb.Teacher{}
	modelVal := reflect.ValueOf(*teacher)
	modelType := modelVal.Type()
	pbVal := reflect.ValueOf(pbTeacher).Elem()

	for i := 0; i < modelVal.NumField(); i++ {
		modelfield := modelVal.Field(i)
		fieldName := modelType.Field(i).Name

		pbfield := pbVal.FieldByName(fieldName)
		if pbfield.IsValid() && pbfield.CanSet() {
			pbfield.Set(modelfield)
		}
	}
	return pbTeacher
}

func mapPBTeacherToModelTeacher(pbTeacher *pb.Teacher) *models.Teacher {
	modelTeacher := models.Teacher{}

	pbVal := reflect.ValueOf(pbTeacher).Elem()
	modelVal := reflect.ValueOf(&modelTeacher).Elem()

	for i := 0; i < pbVal.NumField(); i++ {
		field := pbVal.Field(i)
		fieldName := pbVal.Type().Field(i).Name

		modelfield := modelVal.FieldByName(fieldName)
		if modelfield.IsValid() && modelfield.CanSet() {
			modelfield.Set(field)
		}
	}
	return &modelTeacher
}
