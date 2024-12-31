package gapi

import (
	"google.golang.org/protobuf/types/known/timestamppb"
	db "simplebank/db/sqlc"
	"simplebank/pb"
)

func convertUser(user db.User) *pb.User {
	return &pb.User{
		Username:          user.Username,
		FullName:          user.FullName,
		Email:             user.Email,
		CreatedAt:         timestamppb.New(user.CreatedAt),
		PasswordChangedAt: timestamppb.New(user.PasswordChangedAt),
	}
}
