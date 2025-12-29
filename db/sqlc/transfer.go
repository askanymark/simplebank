package db

import (
	"simplebank/pb"

	"google.golang.org/protobuf/types/known/timestamppb"
)

func (t Transfer) ToResponse() *pb.Transaction {
	// TODO actual numbers
	return &pb.Transaction{
		Id:      t.ID,
		Date:    timestamppb.New(t.CreatedAt),
		Credit:  t.Amount,
		Debit:   t.Amount,
		Balance: 0,
	}
}
