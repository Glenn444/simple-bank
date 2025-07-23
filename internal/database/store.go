package database

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type Store struct {
	*Queries
	db *sql.DB
}

// NewStore creates a new store
func NewStore(db *sql.DB) *Store {
	return &Store{
		db:      db,
		Queries: New(db),
	}
}

// execTx executes a function within a database transaction
func (store *Store) execTx(ctx context.Context, fn func(*Queries) error) error {
	tx, err := store.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	//generate a queries object tied to a specific transaction tx
	q := New(tx)

	err = fn(q)
	if err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			return fmt.Errorf("tx err: %v, rb err: %v", err, rbErr)
		}
		return err // Return the original error, don't continue to commit
	}

	return tx.Commit()
}

type TrasferTxParams struct {
	FromAccountID uuid.UUID       `json:"from_account_id"`
	ToAccountID   uuid.UUID       `json:"to_account_id"`
	Amount        decimal.Decimal `json:"amount"`
}

// TransferTxResult is the result of the transfer
type TransferTxResult struct {
	Transfer    Transfer `json:"transfer"`
	FromAccount Account  `json:"from_account"`
	ToAccount   Account  `json:"to_account"`
	FromEntry   Entry    `json:"from_entry"`
	ToEntry     Entry    `json:"to_entry"`
}

type txKeyType struct{}

var txKey = txKeyType{}
// TransferTx performs a money Transfer from one account to another
func (store *Store) TransferTx(ctx context.Context, arg TrasferTxParams) (TransferTxResult, error) {
	var result TransferTxResult

	err := store.execTx(ctx, func(q *Queries) error {
		var err error
		result.Transfer, err = q.CreateTransfer(ctx, CreateTransferParams(arg))

		if err != nil {
			return err
		}

		txName := ctx.Value(txKey)

		fmt.Println(txName,"create entry 1")
		result.FromEntry, err = q.CreateEntries(ctx, CreateEntriesParams{
			AccountID: arg.FromAccountID,
			Amount:    arg.Amount.Neg(),
		})
		if err != nil {
			return err
		}

		fmt.Println(txName,"create entry 2")
		result.ToEntry, err = q.CreateEntries(ctx, CreateEntriesParams{
			AccountID: arg.ToAccountID,
			Amount:    arg.Amount,
		})
		if err != nil {
			return err
		}

		//TODO: update accounts' balance

		fmt.Println(txName,"Get Account 1")
		account1, err := q.GetAccountByIdForUpdate(ctx, arg.FromAccountID)
		if err != nil {
			return err
		}

		fmt.Println(txName,"Get Account 2")
		account2, err := q.GetAccountByIdForUpdate(ctx, arg.ToAccountID)
		if err != nil {
			return err
		}

		//1. Check if senders balance is able to transfer amount
		if account1.Balance.LessThan(arg.Amount) {
			return fmt.Errorf("insufficient funds: account %v has balance %v, transfer amount %v",
				arg.FromAccountID, account1.Balance, arg.Amount)
		}

		updateAccount1Balance := account1.Balance.Sub(arg.Amount)
		updateAccount2Balance := account2.Balance.Add(arg.Amount)

		fmt.Println(txName,"update Account 1")
		err = q.UpdateAccount(ctx, UpdateAccountParams{
			ID:      arg.FromAccountID,
			Balance: updateAccount1Balance,
		})
		if err != nil {
			return err
		}

		fmt.Println(txName,"update Account 2")
		err = q.UpdateAccount(ctx, UpdateAccountParams{
			ID:      arg.ToAccountID,
			Balance: updateAccount2Balance,
		})
		if err != nil {
			return err
		}

		fmt.Println(txName,"Get Account 1")
		result.FromAccount, err = q.GetAccount(ctx, arg.FromAccountID)
		if err != nil {
			return err
		}

		fmt.Println(txName,"Get Account 2")
		result.ToAccount, err = q.GetAccount(ctx, arg.ToAccountID)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return TransferTxResult{}, err
	}
	return result, nil
}
