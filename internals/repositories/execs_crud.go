package repositories

import (
	"context"
	"errors"
	"fmt"
	"school_project_grpc/internals/models"
	"school_project_grpc/internals/repositories/mongodb"
	"school_project_grpc/pkg/utils"
	pb "school_project_grpc/proto/gen"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func AddExecsDBHandler(ctx context.Context, execsFromReq []*pb.Exec) ([]*pb.Exec, error) {

	// creating db client throught with I will be inserting data
	client, err := mongodb.CreatMongoClient()
	if err != nil {
		utils.ErrorHandler(err, "internal error")
		return nil, err
	}
	defer client.Disconnect(ctx) // alway must be closed after usage

	newExecs := make([]*models.Exec, 0, len(execsFromReq)) //  pb value  to model value
	for i, pbExec := range execsFromReq {
		newExecs = append(newExecs, MapPBToModelExec(pbExec))

		// encoding the password into hash (security)
		hashedPassword, err := utils.HashPassword(newExecs[i].Password)
		if err != nil {
			return nil, utils.ErrorHandler(err, "Internal error")
		}
		newExecs[i].Password = hashedPassword // ovevwriting password of the current newExecs

		// setting the curret time to the field UserCreatedAt
		currentTime := time.Now().Format(time.RFC3339)
		newExecs[i].UserCreatedAt = currentTime // overwrite the field with the current time
	}

	var addedExec []*pb.Exec

	for _, exec := range newExecs {

		if exec == nil {
			continue
		}

		result, err := client.Database("school").Collection("execs").InsertOne(ctx, exec)
		if err != nil {
			return nil, utils.ErrorHandler(err, "Error adding value into database")
		}

		objectID, ok := result.InsertedID.(primitive.ObjectID)
		if ok {
			exec.Id = objectID.Hex()
		}

		pbExec := MapModelToPbExec(exec)

		addedExec = append(addedExec, pbExec)
	}
	return addedExec, nil
}

func GetExecsDBHandler(ctx context.Context, sortOption bson.D, filter bson.M) ([]*pb.Exec, error) {
	client, err := mongodb.CreatMongoClient()
	if err != nil {
		return nil, err
	}
	defer client.Disconnect(ctx)

	// getting collection of the execs
	coll := client.Database("school").Collection("execs")

	var cursor *mongo.Cursor
	if len(sortOption) < 1 {
		cursor, err = coll.Find(ctx, filter)
	} else {
		cursor, err = coll.Find(ctx, filter, options.Find().SetSort(sortOption))
	}

	// cheking the error from above coll.find
	if err != nil {
		return nil, utils.ErrorHandler(err, "Failed to fetch data from db")
	}
	defer cursor.Close(ctx)

	// decode mongo documents -> pb execs
	execs, err := DecodedEntities(ctx, cursor, func() *models.Exec { return &models.Exec{} }, func() *pb.Exec { return &pb.Exec{} })
	if err != nil {
		return nil, utils.ErrorHandler(err, "Failed to fetch data from db")
	}

	return execs, nil
}

// Update Execs in MongoDB
func UpdateExecsDBHandler(ctx context.Context, pbExecs []*pb.Exec) ([]*pb.Exec, error) {
	client, err := mongodb.CreatMongoClient()
	if err != nil {
		return nil, utils.ErrorHandler(err, "failed to created monogdb client")
	}
	defer client.Disconnect(ctx)

	var updatedExecs []*pb.Exec

	for _, exec := range pbExecs {

		// Validate ID
		if exec.Id == "" {
			return nil, utils.ErrorHandler(errors.New("Missing id: invalid request"), "ID cannot be blank")
		}

		// Convert pb -> model
		modelExec := MapPBToModelExec(exec)

		// Convert string ID -> Mongo ObjectID
		obj, err := primitive.ObjectIDFromHex(modelExec.Id)
		if err != nil {
			return nil, utils.ErrorHandler(err, "invalid id")
		}

		// Convert model -> bson
		mExec, err := bson.Marshal(modelExec)
		if err != nil {
			return nil, utils.ErrorHandler(err, "Internal error")
		}

		var updateDoc bson.M
		err = bson.Unmarshal(mExec, &updateDoc)
		if err != nil {
			return nil, utils.ErrorHandler(err, "Internal error")
		}

		// Remove _id from update
		delete(updateDoc, "_id")

		// Update in MongoDB
		_, err = client.Database("school").Collection("Execs").
			UpdateOne(ctx, bson.M{"_id": obj}, bson.M{"$set": updateDoc})
		if err != nil {
			return nil, utils.ErrorHandler(err, fmt.Sprintf("error updating exec id: %s", exec.Id))
		}

		// Convert model -> pb for response
		updatedExec := MapModelToPbExec(modelExec)
		updatedExecs = append(updatedExecs, updatedExec)
	}

	return updatedExecs, nil
}

// delete Exec in mongoDB by user id
func DeleteExecsDBHandler(ctx context.Context, idstodelete []string) ([]string, error) {
	

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

	res, err := client.Database("school").Collection("execs").DeleteMany(ctx, filter)
	if err != nil {
		return nil, utils.ErrorHandler(err, "Internal error")
	}

	if res.DeletedCount == 0 {
		return nil, utils.ErrorHandler(err, "No Execs were deleted")
	}

	// Return deleted IDs
	deletedIds := make([]string, 0, len(objectIds))
	for _, v := range objectIds {
		deletedIds = append(deletedIds, v.Hex())
	}
	return deletedIds, nil
}
