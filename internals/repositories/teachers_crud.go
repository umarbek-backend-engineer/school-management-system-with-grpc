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

// Add teachers to MongoDB
func AddTeachersDBHandler(ctx context.Context, teacherFromReq []*pb.Teacher) ([]*pb.Teacher, error) {
	client, err := mongodb.CreatMongoClient()
	if err != nil {
		utils.ErrorHandler(err, "internal error")
		return nil, err
	}
	defer client.Disconnect(ctx)

	// Convert pb -> model
	newTeachers := make([]*models.Teacher, 0, len(teacherFromReq))
	for _, pbTeacher := range teacherFromReq {
		newTeachers = append(newTeachers, MapPBToModelTeacher(pbTeacher))
	}

	var addedTeacher []*pb.Teacher

	for _, teacher := range newTeachers {
		if teacher == nil {
			continue
		}

		// Insert into MongoDB
		result, err := client.Database("school").Collection("teachers").InsertOne(ctx, teacher)
		if err != nil {
			return nil, utils.ErrorHandler(err, "Error adding value into database")
		}

		// Save generated Mongo ID
		objectID, ok := result.InsertedID.(primitive.ObjectID)
		if ok {
			teacher.Id = objectID.Hex()
		}

		// Convert model -> pb for response
		pbTeacher := MapModelToPbTeacher(teacher)
		addedTeacher = append(addedTeacher, pbTeacher)
	}

	return addedTeacher, nil
}

// Get teachers from MongoDB with optional sorting
func GetTeachersDBhandler(ctx context.Context, sortOption bson.D, filter bson.M, pageSize, pageNumber uint32) ([]*pb.Teacher, error) {
	client, err := mongodb.CreatMongoClient()
	if err != nil {
		return nil, utils.ErrorHandler(err, "Internal error")
	}
	defer client.Disconnect(ctx)

	coll := client.Database("school").Collection("teachers")

	findOptions := options.Find()

	findOptions.SetSkip(int64((pageNumber - 1) * pageSize))
	findOptions.SetLimit(int64(pageSize))

	if len(sortOption) > 0 {
		findOptions.SetSort(sortOption)
	}
	cursor, err := coll.Find(ctx, filter, findOptions)

	if err != nil {
		return nil, utils.ErrorHandler(err, "Internal error")
	}
	defer cursor.Close(ctx)

	// Decode Mongo documents -> pb teachers
	teachers, err := DecodedEntities(
		ctx,
		cursor,
		func() *models.Teacher { return &models.Teacher{} },
		func() *pb.Teacher { return &pb.Teacher{} },
	)
	if err != nil {
		return nil, err
	}

	return teachers, nil
}

// Update teachers in MongoDB
func UpdateTeachersDBHandler(ctx context.Context, pbTeachers []*pb.Teacher) ([]*pb.Teacher, error) {
	client, err := mongodb.CreatMongoClient()
	if err != nil {
		return nil, utils.ErrorHandler(err, "failed to created monogdb client")
	}
	defer client.Disconnect(ctx)

	var updatedTeachers []*pb.Teacher

	for _, teacher := range pbTeachers {

		// Validate ID
		if teacher.Id == "" {
			return nil, utils.ErrorHandler(errors.New("Missing id: invalid request"), "ID cannot be blank")
		}

		// Convert pb -> model
		modelTeacher := MapPBToModelTeacher(teacher)

		// Convert string ID -> Mongo ObjectID
		obj, err := primitive.ObjectIDFromHex(modelTeacher.Id)
		if err != nil {
			return nil, utils.ErrorHandler(err, "invalid id")
		}

		// Convert model -> bson
		mteacher, err := bson.Marshal(modelTeacher)
		if err != nil {
			return nil, utils.ErrorHandler(err, "Internal error")
		}

		var updateDoc bson.M
		err = bson.Unmarshal(mteacher, &updateDoc)
		if err != nil {
			return nil, utils.ErrorHandler(err, "Internal error")
		}

		// Remove _id from update
		delete(updateDoc, "_id")

		// Update in MongoDB
		_, err = client.Database("school").Collection("teachers").
			UpdateOne(ctx, bson.M{"_id": obj}, bson.M{"$set": updateDoc})
		if err != nil {
			return nil, utils.ErrorHandler(err, fmt.Sprintf("error updating teacher id: %s", teacher.Id))
		}

		// Convert model -> pb for response
		updatedTeacher := MapModelToPbTeacher(modelTeacher)
		updatedTeachers = append(updatedTeachers, updatedTeacher)
	}

	return updatedTeachers, nil
}

// delete teacher in mongoDB by user id
func DeleteTeachersDBHandler(ctx context.Context, idsTodelete []string) ([]string, error) {

	client, err := mongodb.CreatMongoClient()
	if err != nil {
		return nil, utils.ErrorHandler(err, "Internal error")
	}
	defer client.Disconnect(ctx)

	// Convert to Mongo ObjectIDs
	objectIds := make([]primitive.ObjectID, 0, len(idsTodelete))
	for _, id := range idsTodelete {
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

func GetStudentCountByTeacherDBHandler(ctx context.Context, id string) (int64, error) {
	client, err := mongodb.CreatMongoClient()
	if err != nil {
		return 0, utils.ErrorHandler(err, "Internal error")
	}
	defer client.Disconnect(ctx)

	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return 0, utils.ErrorHandler(err, "Invalid ID")
	}

	var teacher models.Teacher
	err = client.Database("school").Collection("teachers").FindOne(ctx, bson.M{"_id": objectID}).Decode(&teacher)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return 0, utils.ErrorHandler(err, "Teacher not found")
		}
		return 0, utils.ErrorHandler(err, "Internal error")
	}

	count, err := client.Database("school").Collection("students").CountDocuments(ctx, bson.M{"class": teacher.Class})
	if err != nil {
		return 0, utils.ErrorHandler(err, "Internal Error")
	}
	return count, nil
}
