package db

import (
	"context"
	"github.com/stretchr/testify/require"
	"simplebank/util"
	"testing"
	"time"
)

func createRandomEntry(t *testing.T, accountId int64) Entry {
	arg := CreateEntryParams{
		AccountID: accountId,
		Amount:    util.RandomMoney(),
	}
	entry, err := testStore.CreateEntry(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, entry)

	require.NotZero(t, entry.ID)
	require.Equal(t, arg.AccountID, entry.AccountID)
	require.Equal(t, arg.Amount, entry.Amount)
	require.NotZero(t, entry.CreatedAt)

	return entry
}

func TestCreateEntry(t *testing.T) {
	account := createRandomAccount(t)
	createRandomEntry(t, account.ID)
}

func TestGetEntry(t *testing.T) {
	account := createRandomAccount(t)
	entry := createRandomEntry(t, account.ID)

	dbRecord, err := testStore.GetEntry(context.Background(), entry.ID)
	require.NoError(t, err)

	require.Equal(t, entry.ID, dbRecord.ID)
	require.Equal(t, entry.AccountID, dbRecord.AccountID)
	require.Equal(t, entry.Amount, dbRecord.Amount)
	require.WithinDuration(t, dbRecord.CreatedAt, entry.CreatedAt, time.Second)
}

func TestListEntries(t *testing.T) {
	account := createRandomAccount(t)

	for i := 0; i < 10; i++ {
		createRandomEntry(t, account.ID)
	}

	arg := ListEntriesParams{
		AccountID: account.ID,
		Limit:     5,
		Offset:    5,
	}

	dbRecords, err := testStore.ListEntries(context.Background(), arg)
	require.NoError(t, err)
	require.Len(t, dbRecords, 5)

	for _, dbRecord := range dbRecords {
		require.NotEmpty(t, dbRecord)
		require.Equal(t, dbRecord.AccountID, account.ID)
	}
}
