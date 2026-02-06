package handlers

import (
	"reflect"
	"school_project_grpc/pkg/utils"
	"strings"

	pb "school_project_grpc/proto/gen"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// basicly these functions are used to control the out put of the monogdb request

// Build MongoDB filter from request object
func buildfilter(Obj interface{}, model interface{}) (bson.M, error) {
	filter := bson.M{}

	// If request object is nil, return empty filter
	if Obj == nil || reflect.ValueOf(Obj).IsNil() {
		return filter, nil
	}

	// Prepare reflection values
	mval := reflect.ValueOf(model).Elem()
	mtype := mval.Type()

	rval := reflect.ValueOf(Obj).Elem()
	rtype := rval.Type()

	// Copy non-zero fields from request -> model
	for i := 0; i < rval.NumField(); i++ {
		fieldval := rval.Field(i)
		fieldname := rtype.Field(i).Name

		if fieldval.IsValid() && !fieldval.IsZero() {
			mfield := mval.FieldByName(fieldname)
			if mfield.IsValid() && mfield.CanSet() {
				mfield.Set(fieldval)
			}
		}
	}

	// Build Mongo filter from model fields
	for i := 0; i < mval.NumField(); i++ {
		fieldval := mval.Field(i)

		if fieldval.IsValid() && !fieldval.IsZero() {
			bsonTag := mtype.Field(i).Tag.Get("bson")
			bsonTag = strings.TrimSuffix(bsonTag, ",omitempty")

			// Special case for Mongo _id
			if bsonTag == "_id" {
				objid, err := primitive.ObjectIDFromHex(fieldval.String())
				if err != nil {
					return nil, utils.ErrorHandler(err, "Invalid ID")
				}
				filter["_id"] = objid
			} else {
				filter[bsonTag] = fieldval.Interface().(string)
			}
		}
	}

	return filter, nil
}

// Build Mongo sort options from gRPC request
func buildSortOptions(sortFields []*pb.SortField) bson.D {
	var sortOptions bson.D

	for _, field := range sortFields {
		order := 1
		if field.GetOrder() == pb.Order_DESC {
			order = -1
		}
		sortOptions = append(sortOptions, bson.E{Key: field.Field, Value: order})
	}

	return sortOptions
}
