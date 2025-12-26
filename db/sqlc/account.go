package db

import (
	"simplebank/pb"

	"google.golang.org/protobuf/types/known/timestamppb"
)

func (a Account) ToResponse() *pb.Account {
	return &pb.Account{
		Id:        a.ID,
		Owner:     a.Owner,
		Balance:   a.Balance,
		Currency:  pb.Currency(pb.Currency_value[a.Currency]),
		CreatedAt: timestamppb.New(a.CreatedAt),
	}
}
