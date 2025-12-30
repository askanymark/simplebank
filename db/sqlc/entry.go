package db

import (
	"simplebank/pb"

	"google.golang.org/protobuf/types/known/timestamppb"
)

func (e Entry) ToResponse() *pb.Entry {
	return &pb.Entry{
		Id:        e.ID,
		AccountId: e.AccountID,
		Amount:    e.Amount,
		CreatedAt: timestamppb.New(e.CreatedAt),
	}
}
