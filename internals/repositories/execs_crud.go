package repositories

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"school_project_grpc/internals/models"
	"school_project_grpc/internals/repositories/mongodb"
	"school_project_grpc/pkg/utils"
	pb "school_project_grpc/proto/gen"
	"strconv"
	"time"

	"github.com/go-mail/mail"
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

func LoginDBHandler(ctx context.Context, username string) (models.Exec, error) {
	client, err := mongodb.CreatMongoClient()
	if err != nil {
		return models.Exec{}, utils.ErrorHandler(err, "Internal error")
	}
	defer client.Disconnect(ctx)

	// makeing filer for db to know which columt to change
	filter := bson.M{"username": username}
	var exec models.Exec
	err = client.Database("school").Collection("execs").FindOne(ctx, filter).Decode(&exec) // inserting the data recieved of the same id into exec
	if err != nil {
		if err == mongo.ErrNoDocuments { // if there is not user with that username
			return models.Exec{}, utils.ErrorHandler(err, "User not found. Incorrect password/username")
		}
		return models.Exec{}, utils.ErrorHandler(err, "Internal error")
	}
	return exec, nil
}

func UpdatePasswordDBHandler(ctx context.Context, req *pb.UpdatePasswordRequest) (models.Exec, error) {
	client, err := mongodb.CreatMongoClient()
	if err != nil {
		return models.Exec{}, utils.ErrorHandler(err, "Internal error")
	}
	defer client.Disconnect(ctx)

	objectID, err := primitive.ObjectIDFromHex(req.GetId())
	if err != nil {
		return models.Exec{}, utils.ErrorHandler(err, "Invalid ID")
	}

	// retriving the user (exec) from data base
	var user models.Exec
	err = client.Database("school").Collection("execs").FindOne(ctx, bson.M{"_id": objectID}).Decode(&user)
	if err != nil {
		return models.Exec{}, utils.ErrorHandler(err, "Internal error")
	}

	err = utils.VerifyPassword(req.CurrentPassword, user.Password)
	if err != nil {
		return models.Exec{}, utils.ErrorHandler(err, "Incorrect password/username")
	}

	// hashing the password
	hashedPassword, err := utils.HashPassword(req.NewPassword)
	if err != nil {
		return models.Exec{}, utils.ErrorHandler(err, "Internal error")
	}

	// filter of what to update in db
	update := bson.M{
		"$set": bson.M{
			"password":            hashedPassword,
			"password_changed_at": time.Now().Format(time.RFC3339),
		},
	}

	// updating in db
	_, err = client.Database("school").Collection("execs").UpdateOne(ctx, bson.M{"_id": objectID}, update)
	if err != nil {
		return models.Exec{}, utils.ErrorHandler(err, "Internal error")
	}
	return user, nil
}

func ReactivateUserDBHandler(ctx context.Context, ids []string) (*mongo.UpdateResult, error) {
	client, err := mongodb.CreatMongoClient()
	if err != nil {
		return nil, utils.ErrorHandler(err, "Internal error")
	}
	defer client.Disconnect(ctx)

	var objectIDs []primitive.ObjectID // id to store in db format
	for _, id := range ids {
		objectID, err := primitive.ObjectIDFromHex(id) // making id in db format
		if err != nil {
			return nil, utils.ErrorHandler(err, "Invalid ID")
		}
		objectIDs = append(objectIDs, objectID) // store them in list var
	}

	filter := bson.M{"_id": bson.M{"$in": objectIDs}}          // create file to find the spacified row
	update := bson.M{"$set": bson.M{"inactive_status": false}} // stating what to change

	res, err := client.Database("school").Collection("execs").UpdateMany(ctx, filter, update) // change the row
	if err != nil {
		return nil, utils.ErrorHandler(err, "Failed to deactivate users")
	}
	return res, nil
}

func DeactivateUserDBHandler(ctx context.Context, ids []string) (*mongo.UpdateResult, error) {
	client, err := mongodb.CreatMongoClient()
	if err != nil {
		return nil, utils.ErrorHandler(err, "Internal error")
	}
	defer client.Disconnect(ctx)

	var objectIDs []primitive.ObjectID // id to store in db format
	for _, id := range ids {
		objectID, err := primitive.ObjectIDFromHex(id) // making id in db format
		if err != nil {
			return nil, utils.ErrorHandler(err, "Invalid ID")
		}
		objectIDs = append(objectIDs, objectID) // store them in list var
	}

	filter := bson.M{"_id": bson.M{"$in": objectIDs}}         // create file to find the spacified row
	update := bson.M{"$set": bson.M{"inactive_status": true}} // stating what to change

	res, err := client.Database("school").Collection("execs").UpdateMany(ctx, filter, update) // change the row
	if err != nil {
		return nil, utils.ErrorHandler(err, "Failed to deactivate users")
	}
	return res, nil
}

func ForgotPasswordDBHandler(ctx context.Context, email string) error {
	client, err := mongodb.CreatMongoClient()
	if err != nil {
		return utils.ErrorHandler(err, "Internal error")
	}
	defer client.Disconnect(ctx)

	var exec models.Exec
	err = client.Database("school").Collection("execs").FindOne(ctx, bson.M{"email": email}).Decode(&exec) // getting the full user info and storing in in a var
	if err != nil {
		if err == mongo.ErrNoDocuments { // if there is not user with that username
			return utils.ErrorHandler(err, "User not found. Incorrect password/username")
		}
		return utils.ErrorHandler(err, "Internal error")
	}

	tokenbyte := make([]byte, 32) // generate tokne to send to the user
	_, err = rand.Read(tokenbyte)
	if err != nil {
		return utils.ErrorHandler(err, "Failed to generate token")
	}

	token := hex.EncodeToString(tokenbyte) // token that will be sent to the user
	hashedToken := sha256.Sum256(tokenbyte)
	hashedTokenString := hex.EncodeToString(hashedToken[:]) // token that will be stored in db

	duration, err := strconv.Atoi(os.Getenv("RESET_TOKEN_EXP_DURATION"))
	if err != nil {
		return utils.ErrorHandler(err, "Failed to get token exp duration")
	}

	mins := time.Duration(duration)
	expiry := time.Now().Add(mins * time.Minute).Format(time.RFC3339) // setting up expiry data for the token

	update := bson.M{
		"$set": bson.M{
			"password_reset_token": hashedTokenString,
			"password_token_exp":   expiry,
		},
	}
	_, err = client.Database("school").Collection("execs").UpdateOne(ctx, bson.M{"email": email}, update) // setting token and token exp data into the exec that is requesting
	if err != nil {
		return utils.ErrorHandler(err, "Internal error")
	}

	resetURL := fmt.Sprintf("https://localhost:50051/execs/resetpassword/reset/%s", token) // this will be send to the exec

	message := fmt.Sprintf(`
		Forgot your password? Reset your password using the following link: 
		%s
		Please use the reset code: %s along with your request to change password. 
		if you didn't request a password reset, please ignore this email. 

		This link is only valid for %v minutes.`, resetURL, token, mins)

	subject := "Your password reset link"

	m := mail.NewMessage()
	m.SetHeader("From", "schooladmin@gmail.com") // replay with your own email
	m.SetHeader("To", email)
	m.SetHeader("Subject", subject)
	m.SetBody("text/plain", message)

	d := mail.NewDialer("localhost", 1025, "", "")
	err = d.DialAndSend(m)
	if err != nil {
		cleanup := bson.M{
			"$set": bson.M{
				"password_reset_token": nil,
				"password_token_exp":   nil,
			},
		}
		_, _ = client.Database("school").Collection("execs").UpdateOne(ctx, bson.M{"email": email}, cleanup) // resetting the token and token exp columns in case of an error
		return utils.ErrorHandler(err, "Failed to send password reset link.")
	}
	return nil
}

func ResetPasswordDBHandler(ctx context.Context, hashedTokenString string, password string) error {
	client, err := mongodb.CreatMongoClient()
	if err != nil {
		return utils.ErrorHandler(err, "Internal Error")
	}
	defer client.Disconnect(ctx)

	filter := bson.M{"password_reset_token": hashedTokenString, "password_token_exp": bson.M{"$gt": time.Now().Format(time.RFC3339)}} // building filters and checking if the token is expired or not comparing to time.Now()

	var exec models.Exec
	err = client.Database("school").Collection("execs").FindOne(ctx, filter).Decode(&exec) // store the resulting value in a variable
	if err != nil {
		return utils.ErrorHandler(err, "Invalid or expired token")
	}

	newPassword, err := utils.HashPassword(password)
	if err != nil {
		return utils.ErrorHandler(err, "Internal error")
	}

	update := bson.M{
		"$set": bson.M{
			"password":             newPassword,
			"password_reset_token": nil,
			"password_token_exp":   nil,
			"password_changed_at":  time.Now().Format(time.RFC3339),
		},
	}
	_, err = client.Database("school").Collection("execs").UpdateOne(ctx, filter, update) // setting token and token exp data into the exec that is requesting
	if err != nil {
		return utils.ErrorHandler(err, "Internal error")
	}
	return nil
}
