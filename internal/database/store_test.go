package database

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"
)

type TransferResult struct {
	Result TransferTxResult
	Err    error
}

func TestTransferTx_Success(t *testing.T) {
	store := NewStore(testDB)
	account1 := createRandomAccount(t)
	account2 := createRandomAccount(t)

	//Store inial Balances
	initialBalance1 := account1.Balance
	initialBalance2 := account2.Balance
	amount := decimal.NewFromInt(10)

	result, err := store.TransferTx(context.Background(), TrasferTxParams{
		FromAccountID: account1.ID,
		ToAccountID:   account2.ID,
		Amount:        amount,
	})

	require.NoError(t, err)
	require.NotEmpty(t, result)

	//Verify transfer record
	transfer := result.Transfer
	require.NotEmpty(t, transfer)
	require.Equal(t, account1.ID, transfer.FromAccountID)
	require.Equal(t, account2.ID, transfer.ToAccountID)
	require.Equal(t, amount.Round(2), transfer.Amount.Round(2))
	require.NotZero(t, transfer.ID)
	require.NotZero(t, transfer.CreatedAt)

	//Verify entry records
	require.NotEmpty(t, result.FromEntry)
	require.NotEmpty(t, result.ToEntry)
	require.Equal(t, account1.ID, result.FromEntry.AccountID)
	require.Equal(t, account2.ID, result.ToEntry.AccountID)
	require.Equal(t, amount.Neg().Round(2), result.FromEntry.Amount.Round(2))
	require.Equal(t, amount.Round(2), result.ToEntry.Amount.Round(2))

	//Verify account balances were updated correctly
	require.Equal(t, initialBalance1.Sub(amount), result.FromAccount.Balance)
	require.Equal(t, initialBalance2.Add(amount), result.ToAccount.Balance)

	//Verify balances in database
	updatedAccount1, err := store.GetAccount(context.Background(), account1.ID)
	require.NoError(t, err)
	require.Equal(t, initialBalance1.Sub(amount), updatedAccount1.Balance)

	updatedAccount2, err := store.GetAccount(context.Background(), account1.ID)
	require.NoError(t, err)
	require.Equal(t, initialBalance2.Sub(amount), updatedAccount2.Balance)
}

func TestTransferTx_Concurrent(t *testing.T) {
	store := NewStore(testDB)
	account1 := createRandomAccount(t)
	account2 := createRandomAccount(t)

	// Store initial balances
	initialBalance1 := account1.Balance
	initialBalance2 := account2.Balance
	amount := decimal.NewFromInt(10)
	n := 5

	// Channel to collect results
	results := make(chan TransferResult, n)

	// Run concurrent transfers
	for range n {
		go func() {
			result, err := store.TransferTx(context.Background(), TrasferTxParams{
				FromAccountID: account1.ID,
				ToAccountID:   account2.ID,
				Amount:        amount,
			})
			results <- TransferResult{Result: result, Err: err}
		}()
	}

	// Collect and verify all results
	for range n {
		result := <-results
		require.NoError(t, result.Err)
		require.NotEmpty(t, result.Result)

		// Verify transfer details
		transfer := result.Result.Transfer
		require.Equal(t, account1.ID, transfer.FromAccountID)
		require.Equal(t, account2.ID, transfer.ToAccountID)
		require.Equal(t, amount.Round(2), transfer.Amount.Round(2))
	}

	// Verify final balances are correct
	finalAccount1, err := store.GetAccount(context.Background(), account1.ID)
	require.NoError(t, err)

	finalAccount2, err := store.GetAccount(context.Background(), account2.ID)
	require.NoError(t, err)

	// Calculate expected balances
	totalTransferred := amount.Mul(decimal.NewFromInt(int64(n)))
	expectedBalance1 := initialBalance1.Sub(totalTransferred)
	expectedBalance2 := initialBalance2.Add(totalTransferred)

	require.Equal(t, expectedBalance1.Round(2), finalAccount1.Balance.Round(2))
	require.Equal(t, expectedBalance2, finalAccount2.Balance)

	// Verify total money in system is conserved
	totalBefore := initialBalance1.Add(initialBalance2)
	totalAfter := finalAccount1.Balance.Add(finalAccount2.Balance)
	require.Equal(t, totalBefore, totalAfter)
}

func TestTransferTx_InsufficientFunds(t *testing.T) {
	store := NewStore(testDB)
	account1 := createRandomAccount(t)
	account2 := createRandomAccount(t)

	// Try to transfer more than account1's balance
	largeAmount := account1.Balance.Add(decimal.NewFromInt(100))

	result, err := store.TransferTx(context.Background(), TrasferTxParams{
		FromAccountID: account1.ID,
		ToAccountID:   account2.ID,
		Amount:        largeAmount,
	})

	require.Error(t, err)
	require.Empty(t, result.Transfer)
	require.Contains(t, err.Error(), "Insufficient balance") // Compare error is same as Insufficient balance

	// Verify balances remain unchanged
	finalAccount1, err := store.GetAccount(context.Background(), account1.ID)
	require.NoError(t, err)
	require.Equal(t, account1.Balance, finalAccount1.Balance)

	finalAccount2, err := store.GetAccount(context.Background(), account2.ID)
	require.NoError(t, err)
	require.Equal(t, account2.Balance, finalAccount2.Balance)
}

func TestTransferTx_SameAccount(t *testing.T) {
	store := NewStore(testDB)
	account := createRandomAccount(t)
	amount := decimal.NewFromInt(10)

	result, err := store.TransferTx(context.Background(), TrasferTxParams{
		FromAccountID: account.ID,
		ToAccountID:   account.ID,
		Amount:        amount,
	})

	require.Error(t, err)
	require.Empty(t, result.Transfer)
	require.Contains(t, err.Error(), "different_accounts")

	// Verify balance remains unchanged
	finalAccount, err := store.GetAccount(context.Background(), account.ID)
	require.NoError(t, err)
	require.Equal(t, account.Balance, finalAccount.Balance)
}

func TestTransferTx_InvalidAccounts(t *testing.T) {
	store := NewStore(testDB)
	account1 := createRandomAccount(t)
	amount := decimal.NewFromInt(10)

	testCases := []struct {
		name          string
		fromAccountID uuid.UUID
		toAccountID   uuid.UUID
		expectedError string
	}{
		{
			name:          "Non-existent from account",
			fromAccountID: uuid.MustParse("45c8db32-a05f-4b22-b3db-7b06a71c20f6"),
			toAccountID:   account1.ID,
			expectedError: "violates foreign key constraint",
		},
		{
			name:          "Non-existent to account",
			fromAccountID: account1.ID,
			toAccountID:   uuid.MustParse("45c8db32-a05f-4b22-b3db-7b06a71c20f6"),
			expectedError: "violates foreign key constraint",
		},
		{
			name:          "Both accounts non-existent",
			fromAccountID: uuid.MustParse("45c8db32-a05f-4b22-b3db-7b06a71c20f6"),
			toAccountID:   uuid.MustParse("64de9240-9ae0-4e1c-a2f2-dc5a6a8de5ec"),
			expectedError: "violates foreign key constraint",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := store.TransferTx(context.Background(), TrasferTxParams{
				FromAccountID: tc.fromAccountID,
				ToAccountID:   tc.toAccountID,
				Amount:        amount,
			})

			
			require.Error(t, err)
			require.Empty(t, result.Transfer)
			require.Contains(t, err.Error(), tc.expectedError)
		})
	}
}

func TestTransferTx_InvalidAmounts(t *testing.T) {
	store := NewStore(testDB)
	account1 := createRandomAccount(t)
	account2 := createRandomAccount(t)

	testCases := []struct {
		name          string
		amount        decimal.Decimal
		expectedError string
	}{
		{
			name:          "Zero amount",
			amount:        decimal.Zero,
			expectedError: "violates check constraint",
		},
		{
			name:          "Negative amount",
			amount:        decimal.NewFromInt(-10),
			expectedError: "violates check constraint",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := store.TransferTx(context.Background(), TrasferTxParams{
				FromAccountID: account1.ID,
				ToAccountID:   account2.ID,
				Amount:        tc.amount,
			})

			fmt.Printf("invalid_amount %v\n",err)
			require.Error(t, err)
			require.Empty(t, result.Transfer)
			require.Contains(t, err.Error(), tc.expectedError)
		})
	}
}

func TestTransferTx_LargeAmount(t *testing.T) {
	store := NewStore(testDB)

	// Create accounts with sufficient balance for large transfer
	account1 := createAccountWithBalance(t, store, decimal.NewFromInt(2000000))
	account2 := createRandomAccount(t)

	largeAmount := decimal.NewFromInt(1000000)
	initialBalance1 := account1.Balance
	initialBalance2 := account2.Balance

	result, err := store.TransferTx(context.Background(), TrasferTxParams{
		FromAccountID: account1.ID,
		ToAccountID:   account2.ID,
		Amount:        largeAmount,
	})

	require.NoError(t, err)
	require.NotEmpty(t, result)

	// Verify precision is maintained
	require.Equal(t, largeAmount.Round(2), result.Transfer.Amount.Round(2))
	require.Equal(t, initialBalance1.Sub(largeAmount).Round(2), result.FromAccount.Balance.Round(2))
	require.Equal(t, initialBalance2.Add(largeAmount).Round(2), result.ToAccount.Balance.Round(2))

	// Verify in database
	updatedAccount1, err := store.GetAccount(context.Background(), account1.ID)
	require.NoError(t, err)
	require.Equal(t, initialBalance1.Sub(largeAmount), updatedAccount1.Balance)

	updatedAccount2, err := store.GetAccount(context.Background(), account2.ID)
	require.NoError(t, err)
	require.Equal(t, initialBalance2.Add(largeAmount), updatedAccount2.Balance)
}

func TestTransferTx_ConcurrentBidirectional(t *testing.T) {
	store := NewStore(testDB)
	account1 := createAccountWithBalance(t, store, decimal.NewFromInt(1000))
	account2 := createAccountWithBalance(t, store, decimal.NewFromInt(1000))

	initialBalance1 := account1.Balance
	initialBalance2 := account2.Balance
	amount := decimal.NewFromInt(10)
	n := 5

	var wg sync.WaitGroup
	errs := make(chan error, n*2)

	// Run concurrent transfers in both directions
	for range n {
		wg.Add(2)

		// Transfer from account1 to account2
		go func() {
			defer wg.Done()
			_, err := store.TransferTx(context.Background(), TrasferTxParams{
				FromAccountID: account1.ID,
				ToAccountID:   account2.ID,
				Amount:        amount,
			})
			errs <- err
		}()

		// Transfer from account2 to account1
		go func() {
			defer wg.Done()
			_, err := store.TransferTx(context.Background(), TrasferTxParams{
				FromAccountID: account2.ID,
				ToAccountID:   account1.ID,
				Amount:        amount,
			})
			errs <- err
		}()
	}

	wg.Wait()
	close(errs)

	// Check all operations succeeded
	for err := range errs {
		require.NoError(t, err)
	}

	// Verify final balances (should be the same as initial since equal transfers both ways)
	finalAccount1, err := store.GetAccount(context.Background(), account1.ID)
	require.NoError(t, err)

	finalAccount2, err := store.GetAccount(context.Background(), account2.ID)
	require.NoError(t, err)

	require.Equal(t, initialBalance1, finalAccount1.Balance)
	require.Equal(t, initialBalance2, finalAccount2.Balance)
}

func TestTransferTx_DeadlockPrevention(t *testing.T) {
	store := NewStore(testDB)
	account1 := createAccountWithBalance(t, store, decimal.NewFromInt(1000))
	account2 := createAccountWithBalance(t, store, decimal.NewFromInt(1000))

	amount := decimal.NewFromInt(10)
	n := 10

	var wg sync.WaitGroup
	errs := make(chan error, n*2)

	// Run many concurrent transfers in both directions to test deadlock prevention
	for range n {
		wg.Add(2)

		go func() {
			defer wg.Done()
			_, err := store.TransferTx(context.Background(), TrasferTxParams{
				FromAccountID: account1.ID,
				ToAccountID:   account2.ID,
				Amount:        amount,
			})
			errs <- err
		}()

		go func() {
			defer wg.Done()
			_, err := store.TransferTx(context.Background(), TrasferTxParams{
				FromAccountID: account2.ID,
				ToAccountID:   account1.ID,
				Amount:        amount,
			})
			errs <- err
		}()
	}

	// Add timeout to detect deadlocks
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		// All operations completed successfully
	case <-time.After(30 * time.Second):
		t.Fatal("Test timed out - possible deadlock detected")
	}

	close(errs)

	// Verify no errors occurred
	for err := range errs {
		require.NoError(t, err)
	}
}

func TestTransferTx_ContextCancellation(t *testing.T) {
	store := NewStore(testDB)
	account1 := createRandomAccount(t)
	account2 := createRandomAccount(t)
	amount := decimal.NewFromInt(10)

	// Create a context that's already cancelled
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	result, err := store.TransferTx(ctx, TrasferTxParams{
		FromAccountID: account1.ID,
		ToAccountID:   account2.ID,
		Amount:        amount,
	})

	require.Error(t, err)
	require.Empty(t, result.Transfer)
	require.Contains(t, err.Error(), "context canceled")

	// Verify balances remain unchanged
	finalAccount1, err := store.GetAccount(context.Background(), account1.ID)
	require.NoError(t, err)
	require.Equal(t, account1.Balance, finalAccount1.Balance)

	finalAccount2, err := store.GetAccount(context.Background(), account2.ID)
	require.NoError(t, err)
	require.Equal(t, account2.Balance, finalAccount2.Balance)
}

// Helper function to create an account with specific balance
func createAccountWithBalance(t *testing.T, store *Store, balance decimal.Decimal) Account {
	t.Helper()

	account := createRandomAccount(t)

	// Update the account balance (you'll need to implement this based on your store methods)
	err := store.UpdateAccount(context.Background(), UpdateAccountParams{
		ID:      account.ID,
		Balance: balance,
	})
	require.NoError(t, err)

	return account
}

// Benchmark tests
func BenchmarkTransferTx(b *testing.B) {
	store := NewStore(testDB)
	account1 := createRandomAccount(b)
	account2 := createRandomAccount(b)
	amount := decimal.NewFromInt(10)

	
	for b.Loop() {
		_, err := store.TransferTx(context.Background(), TrasferTxParams{
			FromAccountID: account1.ID,
			ToAccountID:   account2.ID,
			Amount:        amount,
		})
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkTransferTx_Concurrent(b *testing.B) {
	store := NewStore(testDB)
	account1 := createRandomAccount(b)
	account2 := createRandomAccount(b)
	amount := decimal.NewFromInt(1)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, err := store.TransferTx(context.Background(), TrasferTxParams{
				FromAccountID: account1.ID,
				ToAccountID:   account2.ID,
				Amount:        amount,
			})
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}
