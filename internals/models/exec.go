package models

type Exec struct {
	Id                 string `protobuf:"id,omitempty" bson:"_id,omitempty"`
	FirstName          string `protobuf:"first_name,omitempty" bson:"first_name,omitempty"`
	LastName           string `protobuf:"last_name,omitempty" bson:"last_name,omitempty"`
	Email              string `protobuf:"email,omitempty" bson:"email,omitempty"`
	Username           string `protobuf:"username,omitmepty" bson:"username,omitempty"`
	Password           string `protobuf:"password,omitmepty" bson:"password,omitempty"`
	Role               string `protobuf:"role,omitmepty" bson:"role,omitempty"`
	PasswordChangedAt  string `protobuf:"password_changed_at,omitmepty" bson:"password_changed_at,omitempty"`
	UserCreatedAt      string `protobuf:"user_created_at,omitmepty" bson:"user_created_at,omitempty"`
	PasswordResetToken string `protobuf:"password_reset_token,omitmepty" bson:"password_reset_token,omitempty"`
	PasswordTokenExp   string `protobuf:"password_token_exp,omitmepty" bson:"password_token_exp,omitempty"`
	InactiveStatus     bool `protobuf:"inactive_status" bson:"inactive_status"`
}


