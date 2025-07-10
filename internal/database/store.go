package database

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type Store struct{
	*Queries
	db *sql.DB
}


//NewStore creates a new store
func NewStore(db *sql.DB)*Store{
	return &Store{
		db: db,
		Queries: New(db),
	}
}

//execTx executes a function within a database transaction
func (store *Store) execTx(ctx context.Context, fn func(*Queries)error)error{
	tx,err := store.db.BeginTx(ctx,nil)
	if err != nil{
		return err
	}

	//generate a queries object tied to a specific transaction tx
	q := New(tx) 
	
	err = fn(q)
	if err != nil{
		if rbErr := tx.Rollback(); rbErr != nil{
			return fmt.Errorf("tx err: %v, rb err: %v",err,rbErr)
		}
	}

	return tx.Commit()
}

type TrasferTxParams struct{
	FromAccountID uuid.UUID `json:"from_account_id"`
	ToAccountID uuid.UUID `json:"to_account_id"`
	Amount decimal.Decimal `json:"amount"`
}

//TransferTxResult is the result of the transfer
type TransferTxResult struct{
	Transfer Transfer `json:"transfer"`
	FromAccount Account `json:"from_account"`
	ToAccount Account `json:"to_account"`
	FromEntry Entry `json:"from_entry"`
	ToEntry Entry `json:"to_entry"`
}

//TransferTx performs a money Transfer from one account to another
func (store *Store) TransferTx(ctx context.Context,arg TrasferTxParams)(TransferTxResult,error){
	var result TransferTxResult

	err := store.execTx(ctx,func(q *Queries)error{
		var err error
		result.Transfer,err = q.CreateTransfer(ctx,CreateTransferParams(arg))

		if err != nil{
			return err
		}

		result.FromEntry,err = q.CreateEntries(ctx,CreateEntriesParams{
			AccountID: arg.FromAccountID,
			Amount: arg.Amount.Neg(),
		})
		if err != nil{
			return  err
		}
		result.ToEntry,err = q.CreateEntries(ctx,CreateEntriesParams{
			AccountID: arg.ToAccountID,
			Amount: arg.Amount,
		})

		//TODO: update accounts' balance
		if err != nil{
			
		}
		return  nil
	})
	if err != nil{
		return  TransferTxResult{},err
	}
	return  result,nil
}