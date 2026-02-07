package repositories

import (
	"context"
	"errors"
	"fmt"
	"school_project_grpc/internals/models"
	"school_project_grpc/internals/repositories/mongodb"
	"school_project_grpc/pkg/utils"
	pb "school_project_grpc/proto/gen"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
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

// Update students in MongoDB
func UpdateStudentsDBHandler(ctx context.Context, pbStudents []*pb.Student) ([]*pb.Student, error) {
	client, err := mongodb.CreatMongoClient()
	if err != nil {
		return nil, utils.ErrorHandler(err, "failed to created monogdb client")
	}
	defer client.Disconnect(ctx)

	var updatedStudents []*pb.Student

	for _, student := range pbStudents {

		// Validate ID
		if student.Id == "" {
			return nil, utils.ErrorHandler(errors.New("Missing id: invalid request"), "ID cannot be blank")
		}

		// Convert pb -> model
		modelStudent := MapPBToModelStudent(student)

		// Convert string ID -> Mongo ObjectID
		obj, err := primitive.ObjectIDFromHex(modelStudent.Id)
		if err != nil {
			return nil, utils.ErrorHandler(err, "invalid id")
		}

		// Convert model -> bson
		mstudent, err := bson.Marshal(modelStudent)
		if err != nil {
			return nil, utils.ErrorHandler(err, "Internal error")
		}

		var updateDoc bson.M
		err = bson.Unmarshal(mstudent, &updateDoc)
		if err != nil {
			return nil, utils.ErrorHandler(err, "Internal error")
		}

		// Remove _id from update
		delete(updateDoc, "_id")

		// Update in MongoDB
		_, err = client.Database("school").Collection("students").
			UpdateOne(ctx, bson.M{"_id": obj}, bson.M{"$set": updateDoc})
		if err != nil {
			return nil, utils.ErrorHandler(err, fmt.Sprintf("error updating student id: %s", student.Id))
		}

		// Convert model -> pb for response
		updatedStudent := MapModelToPbStudent(modelStudent)
		updatedStudents = append(updatedStudents, updatedStudent)
	}

	return updatedStudents, nil
}

// delete Student in mongoDB by user id
func DeleteStudentsDBHandler(ctx context.Context, idstodelete []string) ([]string, error) {

	client, err := mongodb.CreatMongoClient()
	if err != nil {
		return nil, utils.ErrorHandler(err, "Internal error")
	}
	defer client.Disconnect(ctx)

	// Convert to Mongo ObjectIDs
	objectIds := make([]primitive.ObjectID, 0, len(idstodelete))
	for _, id := range idstodelete {
		objectId, err := primitive.ObjectIDFromHex(id)
		if err != nil {
			return nil, utils.ErrorHandler(err, fmt.Sprintf("Invalid id: %v", id))
		}
		objectIds = append(objectIds, objectId)
	}

	// Delete many by IDs
	filter := bson.M{"_id": bson.M{"$in": objectIds}}

	res, err := client.Database("school").Collection("students").DeleteMany(ctx, filter)
	if err != nil {
		return nil, utils.ErrorHandler(err, "Internal error")
	}

	if res.DeletedCount == 0 {
		return nil, utils.ErrorHandler(err, "No Students were deleted")
	}

	// Return deleted IDs
	deletedIds := make([]string, 0, len(objectIds))
	for _, v := range objectIds {
		deletedIds = append(deletedIds, v.Hex())
	}
	return deletedIds, nil
}

func GetStudentByTeacherIDDBhandler(ctx context.Context, id string) ([]*pb.Student, error) {
	// connecting to db and created client
	client, err := mongodb.CreatMongoClient()
	if err != nil {
		return nil, utils.ErrorHandler(err, "Failed to create mongo client")
	}
	defer client.Disconnect(ctx) // disconnecting the client

	// makeing the id in a way so that is the same as in database "fcayt32erf7atyeg76d2" = ObjectId("fcayt32erf7atyeg76d2")
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, utils.ErrorHandler(err, "Failed to get primitive object id")
	}

	// retriving the Teacher from data base
	var teacher models.Teacher
	err = client.Database("school").Collection("teachers").FindOne(ctx, bson.M{"_id": objectID}).Decode(&teacher)
	if err != nil {
		if err == mongo.ErrNoDocuments { // if teacher is not found return invalid id message
			return nil, utils.ErrorHandler(err, "Invalid ID")
		}
		return nil, utils.ErrorHandler(err, "Failed to retrive teacher")
	}

	cursor, err := client.Database("school").Collection("students").Find(ctx, bson.M{"calss": teacher.Class})
	if err != nil {
		return nil, utils.ErrorHandler(err, "Internal Error")
	}
	defer cursor.Close(ctx)

	students, err := DecodedEntities(ctx, cursor, func() *models.Student { return &models.Student{} }, func() *pb.Student { return &pb.Student{} })
	if err != nil {
		return nil, utils.ErrorHandler(err, "Internal Error")
	}

	err = cursor.Err()
	if err != nil {
		return nil, utils.ErrorHandler(err, "Internal Error")
	}

	return students, nil
}
