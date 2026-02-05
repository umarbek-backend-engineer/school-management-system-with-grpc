package repositories

import (
	"context"
	"reflect"
	"school_project_grpc/internals/models"
	"school_project_grpc/pkg/utils"
	pb "school_project_grpc/proto/gen"

	"go.mongodb.org/mongo-driver/mongo"
)

func DecodedEntities[T any, M any](ctx context.Context, cursor *mongo.Cursor, newmodel func() *M, newentity func() *T) ([]*T, error) {
	var entities []*T
	for cursor.Next(ctx) {
		model := newmodel()
		err := cursor.Decode(&model)
		if err != nil {
			return nil, err
		}

		entity := newentity()
		modelVal := reflect.ValueOf(model).Elem()
		pbVal := reflect.ValueOf(entity).Elem()
		for i := 0; i < modelVal.NumField(); i++ {
			modelField := modelVal.Field(i)
			modelFieldName := modelVal.Type().Field(i).Name

			pbField := pbVal.FieldByName(modelFieldName)
			if pbField.IsValid() && pbField.CanSet() {
				pbField.Set(modelField)
			}
		}

		entities = append(entities, entity)
	}

	err := cursor.Err()
	if err != nil {
		return nil, utils.ErrorHandler(err, "Internal error")
	}
	return entities, nil
}

func MapModelToPbTeacher(teacher *models.Teacher) *pb.Teacher {
	return mapModelToPb(teacher, func() *pb.Teacher { return &pb.Teacher{} })
}
func mapModelToPb[T any, M any](model *M, newentity func() *T) *T {

	entity := newentity()

	modelVal := reflect.ValueOf(model)
	pbVal := reflect.ValueOf(entity).Elem()
	modelType := modelVal.Type()

	for i := 0; i < modelVal.NumField(); i++ {
		modelfield := modelVal.Field(i)
		fieldName := modelType.Field(i).Name

		pbfield := pbVal.FieldByName(fieldName)
		if pbfield.IsValid() && pbfield.CanSet() {
			pbfield.Set(modelfield)
		}
	}
	return entity
}

func MapPBToModelTeacher(pbTeacher *pb.Teacher) *models.Teacher {
	return mapPBToModel(pbTeacher, func() *models.Teacher { return &models.Teacher{} })
}

func mapPBToModel[T any, M any](entity *T, newModel func() *M) *M {
	model := newModel()
	pbVal := reflect.ValueOf(entity).Elem()
	modelVal := reflect.ValueOf(model).Elem()

	for i := 0; i < pbVal.NumField(); i++ {
		field := pbVal.Field(i)
		fieldName := pbVal.Type().Field(i).Name

		modelfield := modelVal.FieldByName(fieldName)
		if modelfield.IsValid() && modelfield.CanSet() {
			modelfield.Set(field)
		}
	}
	return model
}
