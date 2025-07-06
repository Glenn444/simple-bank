package database

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/Glenn444/banking-app/util"
	"github.com/stretchr/testify/require"
)

func createRandomTransfer(t *testing.T, fromAccount, toAccount Account) Transfer {
	arg := CreateTransferParams{
		FromAccountID: fromAccount.ID,
		ToAccountID:   toAccount.ID,
		Amount:        util.RandomMoney(),
	}

	transfer, err := testQueries.CreateTransfer(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, transfer)

	require.Equal(t, arg.FromAccountID, transfer.FromAccountID)
	require.Equal(t, arg.ToAccountID, transfer.ToAccountID)
	require.Equal(t, arg.Amount.Round(2), transfer.Amount.Round(2))

	require.NotZero(t, transfer.ID)
	require.NotZero(t, transfer.CreatedAt)
	require.NotZero(t, transfer.UpdatedAt)

	return transfer
}

func TestCreateTransfer(t *testing.T) {
	fromAccount := createRandomAccount(t)
	toAccount := createRandomAccount(t)

	createRandomTransfer(t, fromAccount, toAccount)
}

func TestGetTransfer(t *testing.T) {
	fromAccount := createRandomAccount(t)
	toAccount := createRandomAccount(t)

	transfer1 := createRandomTransfer(t, fromAccount, toAccount)

	transfer2, err := testQueries.GetTransfer(context.Background(), transfer1.ID)
	require.NoError(t, err)
	require.NotEmpty(t, transfer2)

	require.Equal(t, transfer1.ID, transfer2.ID)
	require.Equal(t, transfer1.FromAccountID, transfer2.FromAccountID)
	require.Equal(t, transfer1.ToAccountID, transfer2.ToAccountID)
	require.Equal(t, transfer1.Amount.Round(2), transfer2.Amount.Round(2))
	require.WithinDuration(t, transfer1.CreatedAt, transfer2.CreatedAt, time.Second)
	require.WithinDuration(t, transfer1.UpdatedAt, transfer2.UpdatedAt, time.Second)
}

func TestListTransfers(t *testing.T) {
	fromAccount := createRandomAccount(t)
	toAccount := createRandomAccount(t)

	for range 10 {
		createRandomTransfer(t, fromAccount, toAccount)
	}

	arg := ListTransfersParams{
		Limit:  5,
		Offset: 0,
	}

	transfers, err := testQueries.ListTransfers(context.Background(), arg)
	require.NoError(t, err)
	require.Len(t, transfers, 5)

	for _, transfer := range transfers {
		require.NotEmpty(t, transfer)
	}
}

func TestUpdateTransfer(t *testing.T) {
	fromAccount := createRandomAccount(t)
	toAccount := createRandomAccount(t)

	transfer1 := createRandomTransfer(t, fromAccount, toAccount)
	newAmount := util.RandomMoney()

	arg := UpdateTransferParams{
		ID:     transfer1.ID,
		Amount: newAmount,
	}

	err := testQueries.UpdateTransfer(context.Background(), arg)
	require.NoError(t, err)

	transfer2, err := testQueries.GetTransfer(context.Background(), transfer1.ID)
	require.NoError(t, err)
	require.Equal(t, newAmount.Round(2), transfer2.Amount.Round(2))
}

func TestDeleteTransfer(t *testing.T) {
	fromAccount := createRandomAccount(t)
	toAccount := createRandomAccount(t)

	transfer := createRandomTransfer(t, fromAccount, toAccount)

	err := testQueries.DeleteTransfer(context.Background(), transfer.ID)
	require.NoError(t, err)

	deletedTransfer, err := testQueries.GetTransfer(context.Background(), transfer.ID)
	require.Error(t, err)
	require.EqualError(t, err, sql.ErrNoRows.Error())
	require.Empty(t, deletedTransfer)
}
