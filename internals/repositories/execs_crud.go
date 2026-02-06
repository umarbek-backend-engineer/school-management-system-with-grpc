package repositories

import (
	"context"
	"school_project_grpc/internals/models"
	"school_project_grpc/internals/repositories/mongodb"
	"school_project_grpc/pkg/utils"
	pb "school_project_grpc/proto/gen"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

func AddExecsDBHandler(ctx context.Context, execsFromReq []*pb.Exec) ([]*pb.Exec, error) {
	client, err := mongodb.CreatMongoClient()
	if err != nil {
		utils.ErrorHandler(err, "internal error")
		return nil, err
	}
	defer client.Disconnect(ctx)

	newExecs := make([]*models.Exec, 0, len(execsFromReq))
	for _, pbExec := range execsFromReq {
		newExecs = append(newExecs, MapPBToModelExec(pbExec))
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
