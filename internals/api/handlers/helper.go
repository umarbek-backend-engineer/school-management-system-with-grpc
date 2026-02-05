package handlers

import (
	"reflect"
	"school_project_grpc/pkg/utils"
	"strings"

	pb "school_project_grpc/proto/gen"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func buildfilter(Obj interface{}, model interface{}) (bson.M, error) {
	filter := bson.M{}

	if Obj == nil || reflect.ValueOf(Obj).IsNil() {
		return filter, nil
	}

	mval := reflect.ValueOf(model).Elem()
	mtype := mval.Type()

	rval := reflect.ValueOf(Obj).Elem()
	rtype := rval.Type()

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

	// now I interate  model val to build filter using bson.M

	for i := 0; i < mval.NumField(); i++ {
		fieldval := mval.Field(i)

		if fieldval.IsValid() && !fieldval.IsZero() {
			bsonTag := mtype.Field(i).Tag.Get("bson")
			bsonTag = strings.TrimSuffix(bsonTag, ",omitempty")
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
