package db

import (
	"context"
	"github.com/stretchr/testify/require"
	"simplebank/util"
	"testing"
	"time"
)

func createRandomTransfer(t *testing.T, from, to int64) Transfer {
	arg := CreateTransferParams{
		FromAccountID: from,
		ToAccountID:   to,
		Amount:        util.RandomMoney(),
	}

	transfer, err := testQueries.CreateTransfer(context.Background(), arg)

	require.NoError(t, err)
	require.NotEmpty(t, transfer)
	require.Equal(t, arg.FromAccountID, transfer.FromAccountID)
	require.Equal(t, arg.ToAccountID, transfer.ToAccountID)
	require.Equal(t, arg.Amount, transfer.Amount)

	return transfer
}

func TestCreateTransfer(t *testing.T) {
	fromAccount := createRandomAccount(t)
	toAccount := createRandomAccount(t)
	createRandomTransfer(t, fromAccount.ID, toAccount.ID)
}

func TestGetTransfer(t *testing.T) {
	fromAccount := createRandomAccount(t)
	toAccount := createRandomAccount(t)
	transfer := createRandomTransfer(t, fromAccount.ID, toAccount.ID)

	dbRecord, err := testQueries.GetTransfer(context.Background(), transfer.ID)
	require.NoError(t, err)
	require.NotEmpty(t, dbRecord)

	require.Equal(t, fromAccount.ID, dbRecord.FromAccountID)
	require.Equal(t, toAccount.ID, dbRecord.ToAccountID)
	require.Equal(t, transfer.Amount, dbRecord.Amount)
	require.WithinDuration(t, transfer.CreatedAt, dbRecord.CreatedAt, time.Second)
}

func TestListTransfers(t *testing.T) {
	fromAccount := createRandomAccount(t)
	toAccount := createRandomAccount(t)

	for i := 0; i < 10; i++ {
		createRandomTransfer(t, fromAccount.ID, toAccount.ID)
	}

	arg := ListTransfersParams{
		FromAccountID: fromAccount.ID,
		ToAccountID:   toAccount.ID,
		Limit:         5,
		Offset:        5,
	}

	dbRecords, err := testQueries.ListTransfers(context.Background(), arg)
	require.NoError(t, err)
	require.Len(t, dbRecords, 5)

	for _, dbRecord := range dbRecords {
		require.NotEmpty(t, dbRecord)
	}
}
