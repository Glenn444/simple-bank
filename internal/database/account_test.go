package database

import (
	"context"
	"testing"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"
)

func TestCreateAccount(t *testing.T){
	arg := CreateAccountParams{
		Owner: "tom",
		Balance: decimal.NewFromFloat(100.00), //NewFromFloat(100.00)
		Currency: "USD",
	}
	
	
	account, err := testQueries.CreateAccount(context.Background(),arg)
	require.NoError(t,err)
	require.NotEmpty(t,account)

	expectedBalance := arg.Balance.Round(2)
	actualBalance := arg.Balance.Round(2)

	require.Equal(t,arg.Owner,account.Owner)
	require.Equal(t,expectedBalance,actualBalance)
	require.Equal(t,arg.Currency,account.Currency)

	require.NotZero(t,account.ID)
	require.NotZero(t,account.CreatedAt)
}