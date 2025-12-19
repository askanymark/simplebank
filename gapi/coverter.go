package gapi

import (
	db "simplebank/db/sqlc"
	"simplebank/pb"

	"google.golang.org/protobuf/types/known/timestamppb"
)

// convertUser is used to turn db.User into *pb.User to convert Go data fields into database compatible values
func convertUser(user db.User) *pb.User {
	return &pb.User{
		Username:          user.Username,
		FullName:          user.FullName,
		Email:             user.Email,
		CreatedAt:         timestamppb.New(user.CreatedAt),
		PasswordChangedAt: timestamppb.New(user.PasswordChangedAt),
	}
}
