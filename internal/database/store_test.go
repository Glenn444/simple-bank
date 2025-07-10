package database

import (
	"context"
	"testing"

	"github.com/Glenn444/banking-app/util"
	"github.com/stretchr/testify/require"
)

func TestTransferTx(t *testing.T) {
	store := NewStore(testDB)
	account1 := createRandomAccount(t)
	account2 := createRandomAccount(t)

	type TransferResult struct {
		Result TransferTxResult
		Err    error
	}
	n := 5
	ch := make(chan TransferResult, n)

	for range n {
		go func() {
			result, err := store.TransferTx(context.Background(), TrasferTxParams{
				FromAccountID: account1.ID,
				ToAccountID:   account2.ID,
				Amount:        util.RandomMoney(),
			})
			ch <- TransferResult{
				Result: result,
				Err:    err,
			}
		}()
	}

	//check results
	for range n {
		results := <-ch

		require.NoError(t, results.Err)

		require.NotEmpty(t, results.Result)
	}
}
