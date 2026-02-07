package models

type Student struct {
	Id        string `protobuf:"id,omitempty" bson:"_id,omitempty"`
	FirstName string `protobuf:"first_name,omitempty" bson:"first_name,omitempty"`
	LastName  string `protobuf:"last_name,omitempty" bson:"last_name,omitempty"`
	Email     string `protobuf:"email,omitempty" bson:"email,omitempty"`
	Class     string `protobuf:"class,omitempty" bson:"class,omitempty"`
}
