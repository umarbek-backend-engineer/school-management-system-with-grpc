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
