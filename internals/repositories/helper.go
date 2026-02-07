package repositories

import (
	"context"
	"reflect"
	"school_project_grpc/internals/models"
	"school_project_grpc/pkg/utils"
	pb "school_project_grpc/proto/gen"

	"go.mongodb.org/mongo-driver/mongo"
)

/*
DecodedEntities converts a Mongo cursor result (models) into a slice of protobuf entities.

How it works:
- It iterates over the cursor.
- For each document, it decodes into a model struct (*M).
- It creates a protobuf entity (*T).
- It copies fields from the model to the protobuf entity by matching field names (reflection).

Requirements / assumptions:
- Model fields and protobuf fields must have the SAME names and compatible types.
- newmodel returns a pointer to a struct (e.g. &models.Teacher{}).
- newentity returns a pointer to a protobuf struct (e.g. &pb.Teacher{}).

Why generics:
- So you can reuse this logic for Teacher, Exec, and any other entity without duplicating code.
*/
func DecodedEntities[T any, M any](ctx context.Context, cursor *mongo.Cursor, newmodel func() *M, newentity func() *T) ([]*T, error) {

	var entities []*T

	// Loop through cursor results
	for cursor.Next(ctx) {

		// Create a new model instance and decode Mongo document into it
		model := newmodel()
		if err := cursor.Decode(&model); err != nil {
			return nil, err
		}

		// Create a new protobuf entity
		entity := newentity()

		// Use reflection to copy fields from model -> protobuf entity
		// model is a pointer, so we need Elem() to access the struct value
		modelVal := reflect.ValueOf(model).Elem()
		// entity is also a pointer, so Elem() gives the struct inside
		pbVal := reflect.ValueOf(entity).Elem()

		// Go through each field in model and try to set same field in protobuf entity
		for i := 0; i < modelVal.NumField(); i++ {
			modelField := modelVal.Field(i)
			modelFieldName := modelVal.Type().Field(i).Name

			// Find the field with same name inside protobuf struct
			pbField := pbVal.FieldByName(modelFieldName)

			// If field exists and can be set, set it
			if pbField.IsValid() && pbField.CanSet() {
				pbField.Set(modelField)
			}
		}

		entities = append(entities, entity)
	}

	// Check if cursor had errors during iteration
	if err := cursor.Err(); err != nil {
		return nil, utils.ErrorHandler(err, "Internal error")
	}

	return entities, nil
}

/*
mapModelToPb converts a single model (*M) into a protobuf entity (*T)
by copying fields with matching names using reflection.

Requirements:
- model must be a pointer to struct (e.g. *models.Teacher)
- entity (pb struct) must have matching field names and compatible types
*/
func mapModelToPb[T any, M any](model *M, newentity func() *T) *T {
	// Create protobuf entity instance
	entity := newentity()

	// NOTE:
	// model is a pointer. If you want fields you typically do Elem().
	// But here your original code uses reflect.ValueOf(model) directly.
	// That will NOT allow NumField() if it remains a pointer.
	// So usually you'd want: reflect.ValueOf(model).Elem()
	//
	// I'll keep your structure but fix it to be correct and safe.
	modelVal := reflect.ValueOf(model).Elem()
	modelType := modelVal.Type()

	// entity is a pointer, Elem() gives struct
	pbVal := reflect.ValueOf(entity).Elem()

	// Copy each model field -> pb field with same name
	for i := 0; i < modelVal.NumField(); i++ {
		modelField := modelVal.Field(i)
		fieldName := modelType.Field(i).Name

		pbField := pbVal.FieldByName(fieldName)
		if pbField.IsValid() && pbField.CanSet() {
			pbField.Set(modelField)
		}
	}

	return entity
}

// MapModelToPbTeacher maps internal Teacher model -> protobuf Teacher entity.
func MapModelToPbTeacher(teacher *models.Teacher) *pb.Teacher {
	return mapModelToPb(teacher, func() *pb.Teacher { return &pb.Teacher{} })
}

// MapModelToPbExec maps internal Exec model -> protobuf Exec entity.
func MapModelToPbExec(exec *models.Exec) *pb.Exec {
	return mapModelToPb(exec, func() *pb.Exec { return &pb.Exec{} })
}

// MapModelToPbStudent maps internal Student model -> protobuf Student entity.
func MapModelToPbStudent(Student *models.Student) *pb.Student {
	return mapModelToPb(Student, func() *pb.Student { return &pb.Student{} })
}

/*
mapPBToModel converts a protobuf entity (*T) into an internal model (*M)
by copying fields with matching names.

Requirements:
- entity must be a pointer to struct (protobuf generated struct)
- model must be a pointer to struct (internal models struct)
- Field names and types should match or be assignable.
*/
func mapPBToModel[T any, M any](entity *T, newModel func() *M) *M {
	// Create new internal model
	model := newModel()

	// entity is pointer -> Elem() gives struct
	pbVal := reflect.ValueOf(entity).Elem()
	// model is pointer -> Elem() gives struct
	modelVal := reflect.ValueOf(model).Elem()

	// Copy each protobuf field -> model field with same name
	for i := 0; i < pbVal.NumField(); i++ {
		field := pbVal.Field(i)
		fieldName := pbVal.Type().Field(i).Name

		modelField := modelVal.FieldByName(fieldName)
		if modelField.IsValid() && modelField.CanSet() {
			modelField.Set(field)
		}
	}

	return model
}

// MapPBToModelTeacher maps protobuf Teacher -> internal Teacher model.
func MapPBToModelTeacher(pbTeacher *pb.Teacher) *models.Teacher {
	return mapPBToModel(pbTeacher, func() *models.Teacher { return &models.Teacher{} })
}

// MapPBToModelExec maps protobuf Exec -> internal Exec model.
func MapPBToModelExec(pbExec *pb.Exec) *models.Exec {
	return mapPBToModel(pbExec, func() *models.Exec { return &models.Exec{} })
}

// MapPBToModelStrudent maps protobuf Exec -> internal Student model.
func MapPBToModelStudent(pbStudent *pb.Student) *models.Student {
	return mapPBToModel(pbStudent, func() *models.Student { return &models.Student{} })
}
