package repositories

import (
	"context"
	"school_project_grpc/internals/models"
	"school_project_grpc/internals/repositories/mongodb"
	"school_project_grpc/pkg/utils"
	pb "school_project_grpc/proto/gen"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

func AddStudentsDBHandler(ctx context.Context, studentsFromReq []*pb.Student) ([]*pb.Student, error) {
	client, err := mongodb.CreatMongoClient()
	if err != nil {
		utils.ErrorHandler(err, "internal error")
		return nil, err
	}
	defer client.Disconnect(ctx)

	newStudents := make([]*models.Student, 0, len(studentsFromReq))
	for _, pbStudent := range studentsFromReq {
		newStudents = append(newStudents, MapPBToModelStudent(pbStudent))
	}

	var addedStudent []*pb.Student

	for _, student := range newStudents {

		if student == nil {
			continue
		}

		result, err := client.Database("school").Collection("students").InsertOne(ctx, student)
		if err != nil {
			return nil, utils.ErrorHandler(err, "Error adding value into database")
		}

		objectID, ok := result.InsertedID.(primitive.ObjectID)
		if ok {
			student.Id = objectID.Hex()
		}

		pbStudent := MapModelToPbStudent(student)

		addedStudent = append(addedStudent, pbStudent)
	}
	return addedStudent, nil
}
