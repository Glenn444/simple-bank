package database

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/Glenn444/banking-app/util"
	"github.com/stretchr/testify/require"
)
type TestingT interface {
	Helper()
	Fatal(args ...interface{})
	Fatalf(format string, args ...interface{})
	Error(args ...interface{})
	Errorf(format string, args ...interface{})
	FailNow()
	Log(args ...interface{})
	Logf(format string, args ...interface{})
}
func TestCreateAccount(t *testing.T){
	arg := CreateAccountParams{
		Owner: util.RandomOwner(),
		Balance:util.RandomMoney(), //NewFromFloat(100.00)
		Currency: util.RandomCurrency(),
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

//helper func to createRandom Accounts in the db
func createRandomAccount(t TestingT)Account{
	t.Helper()
	arg := CreateAccountParams{
		Owner: util.RandomOwner(),
		Balance: util.RandomMoney(),
		Currency: util.RandomCurrency(),
	}
	account, err := testQueries.CreateAccount(context.Background(),arg)
	require.NoError(t,err)
	require.NotEmpty(t,account)

	return account
}
func TestListAccounts(t *testing.T){
	for range 5{
		createRandomAccount(t)
	}
	arg := ListAccountsParams{
		Offset: 0,
		Limit: 5,
	}
	accounts,err := testQueries.ListAccounts(context.Background(),arg)
	require.NoError(t,err)
	require.Len(t,accounts,5)

	for _, account := range accounts{
		require.NotEmpty(t,account)
	}
}

func TestGetAccount(t *testing.T){
	t.Parallel()

	account1 := createRandomAccount(t)

	account2,err := testQueries.GetAccount(context.Background(),account1.ID)
	require.NoError(t,err)
	require.NotEmpty(t,account2)

	//Assert retrieved values match
	require.Equal(t,account1.ID,account2.ID)
	require.Equal(t,account1.Owner,account2.Owner)
	require.Equal(t,account1.Balance.Round(2),account2.Balance.Round(2))
	require.Equal(t,account1.Currency,account2.Currency)
	require.WithinDuration(t,account1.CreatedAt,account2.CreatedAt,1e9)
}


func TestUpdateAccount(t *testing.T){
	t.Parallel()

	account1 := createRandomAccount(t)
	newBalance := util.RandomMoney()

	arg := UpdateAccountParams{
		ID: account1.ID,
		Balance:newBalance,
	}
	err := testQueries.UpdateAccount(context.Background(),arg)
	require.NoError(t,err)

	account2,err := testQueries.GetAccount(context.Background(),account1.ID)
	require.NoError(t,err)

	
	require.Equal(t,account1.ID,account2.ID)
	require.NotEqual(t,account1.Balance.Round(2),account2.Balance.Round(2))
	require.Equal(t,newBalance.Round(2),account2.Balance.Round(2))

	//Confirm unchanged fields
	require.Equal(t,account1.Owner,account2.Owner)
	require.Equal(t,account1.Currency,account2.Currency)
	require.WithinDuration(t,account1.CreatedAt,account2.CreatedAt,time.Second)
}

func TestAccountDeletion(t *testing.T){
	t.Parallel()

	account1 := createRandomAccount(t)

	err := testQueries.DeleteAccount(context.Background(),account1.ID)
	require.NoError(t,err)

	account2,err := testQueries.GetAccount(context.Background(),account1.ID)
	//require.NoError(t,err)

	//Expect an SQL error
	require.Error(t,err)
	require.EqualError(t,err,sql.ErrNoRows.Error())

	//returned account2 should be empty
	require.Empty(t,account2)
}