package repositories

import (
	"context"
	"school_project_grpc/internals/models"
	"school_project_grpc/internals/repositories/mongodb"
	"school_project_grpc/pkg/utils"
	pb "school_project_grpc/proto/gen"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
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

func GetStudentsDBHandler(ctx context.Context, sortOption bson.D, filter bson.M, pageSize, pageNumber uint32) ([]*pb.Student, error) {
	client, err := mongodb.CreatMongoClient()
	if err != nil {
		return nil, err
	}
	defer client.Disconnect(ctx)

	// getting collection of the execs
	coll := client.Database("school").Collection("students")

	findOptions := options.Find()

	findOptions.SetSkip(int64((pageNumber - 1) * pageSize))
	findOptions.SetLimit(int64(pageSize))

	if len(sortOption) > 0 {
		findOptions.SetSort(sortOption)
	}
	cursor, err := coll.Find(ctx, filter, findOptions)

	// cheking the error from above coll.find
	if err != nil {
		return nil, utils.ErrorHandler(err, "Failed to fetch data from db")
	}
	defer cursor.Close(ctx)

	// decode mongo documents to pb.students
	students, err := DecodedEntities(ctx, cursor, func() *models.Student { return &models.Student{} }, func() *pb.Student { return &pb.Student{} })
	if err != nil {
		return nil, utils.ErrorHandler(err, "Failed to fetch data from db")
	}

	return students, nil
}
