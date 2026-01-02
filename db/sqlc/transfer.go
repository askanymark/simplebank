package db

import (
	"simplebank/pb"

	"google.golang.org/protobuf/types/known/timestamppb"
)

func (t Transfer) ToResponse() *pb.Transfer {
	// TODO actual numbers
	return &pb.Transfer{
		Id:      t.ID,
		Date:    timestamppb.New(t.CreatedAt),
		Credit:  t.Amount,
		Debit:   t.Amount,
		Balance: 0,
		Description: func() *string {
			if t.Description.Valid {
				return &t.Description.String
			}
			return nil
		}(),
	}
}
