package database

import (
	"context"
	"database/sql"
	"fmt"
	"testing"
	"time"

	"github.com/Glenn444/banking-app/util"
	"github.com/stretchr/testify/require"
)



func createRandomEntry(t *testing.T,account Account)Entry{
	arg := CreateEntriesParams{
		AccountID: account.ID,
		Amount: util.RandomMoney(),
	}

	entry,err := testQueries.CreateEntries(context.Background(),arg)
	require.NoError(t,err)
	require.NotEmpty(t,entry)

	require.Equal(t,arg.AccountID,entry.AccountID)
	require.Equal(t, arg.Amount.Round(2), entry.Amount.Round(2))

	require.NotZero(t, entry.ID)
	require.NotZero(t, entry.CreatedAt)
	require.NotZero(t, entry.UpdatedAt)

	return entry
}

func TestCreateEntry(t *testing.T){
	account := createRandomAccount(t)
	createRandomEntry(t,account)
}

func TestGetEntry(t *testing.T){
	account := createRandomAccount(t)
	entry1 := createRandomEntry(t,account)
	fmt.Print(entry1)
	entry2,err := testQueries.GetEntry(context.Background(),entry1.ID)
	require.NoError(t, err)
	require.NotEmpty(t, entry2)

	require.Equal(t, entry1.ID, entry2.ID)
	require.Equal(t, entry1.AccountID, entry2.AccountID)
	require.Equal(t, entry1.Amount.Round(2), entry2.Amount.Round(2))
	require.WithinDuration(t, entry1.CreatedAt.Time, entry2.CreatedAt.Time, time.Second)
}

func TestListEntries(t *testing.T) {
	account := createRandomAccount(t)
	for range 10 {
		createRandomEntry(t, account)
	}

	arg := ListEntriesParams{
		Limit:  5,
		Offset: 0,
	}

	entries, err := testQueries.ListEntries(context.Background(), arg)
	require.NoError(t, err)
	require.Len(t, entries, 5)

	for _, entry := range entries {
		require.NotEmpty(t, entry)
	}
}
func TestUpdateEntry(t *testing.T) {
	account := createRandomAccount(t)
	entry1 := createRandomEntry(t, account)

	newAmount := util.RandomMoney()
	arg := UpdateEntryParams{
		ID:     entry1.ID,
		Amount: newAmount,
	}

	err := testQueries.UpdateEntry(context.Background(), arg)
	require.NoError(t, err)

	entry2, err := testQueries.GetEntry(context.Background(), entry1.ID)
	require.NoError(t, err)
	require.Equal(t, entry1.ID, entry2.ID)
	require.Equal(t, newAmount.Round(2), entry2.Amount.Round(2))
}
func TestDeleteEntry(t *testing.T) {
	account := createRandomAccount(t)
	entry := createRandomEntry(t, account)

	err := testQueries.DeleteEntry(context.Background(), entry.ID)
	require.NoError(t, err)

	deletedEntry, err := testQueries.GetEntry(context.Background(), entry.ID)
	require.Error(t, err)
	require.EqualError(t, err, sql.ErrNoRows.Error())
	require.Empty(t, deletedEntry)
}
