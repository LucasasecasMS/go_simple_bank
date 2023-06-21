package db

import (
	"context"
	"database/sql"
	"fmt"
)

type Store struct {
	*Queries
	db *sql.DB
}

func NewStore(db *sql.DB) *Store {
	return &Store{
		db:      db,
		Queries: New(db),
	}
}

func (store *Store) execTx(ctx context.Context, fn func(*Queries) error) error {

	tx, err := store.db.BeginTx(ctx, nil)

	if err != nil {
		return err
	}

	q := New(tx)
	err = fn(q)

	if err != nil {
		if errTx := tx.Rollback(); errTx != nil {
			fmt.Errorf("tx error: %v, rb error: %v", errTx, err)
		}
		return err
	}

	return tx.Commit()
}

type TransferTxParams struct {
	FromAccountID int64 `json:"from_account_id"`
	ToAccountID   int64 `json:"to_account_id"`
	Amount        int64 `json:"amount"`
}

type TransferTxResult struct {
	FromEntry   Entry    `json:"from_entry"`
	ToEntry     Entry    `json:"to_entry"`
	FromAccount Account  `json:"from_account"`
	ToAccount   Account  `json:"to_account"`
	Transfer    Transfer `json:"transfer"`
}

func (store *Store) TransferTX(ctx context.Context, params TransferTxParams) (TransferTxResult, error) {
	var result TransferTxResult
	var err error

	store.execTx(ctx, func(q *Queries) error {
		result.Transfer, err = q.CreateTransfer(ctx, CreateTransferParams{
			FromAccountID: params.FromAccountID,
			ToAccountID:   params.ToAccountID,
			Amount:        params.Amount,
		})

		if err != nil {
			return err
		}

		result.FromEntry, err = q.CreateEntry(ctx, CreateEntryParams{
			AccountID: params.FromAccountID,
			Amount:    -params.Amount,
		})

		if err != nil {
			return err
		}

		result.ToEntry, err = q.CreateEntry(ctx, CreateEntryParams{
			AccountID: params.ToAccountID,
			Amount:    params.Amount,
		})
		if err != nil {
			return err
		}

		if params.FromAccountID < params.ToAccountID {
			result.FromAccount, result.ToAccount, err = updateAccountBalance(ctx, q, params.FromAccountID, -params.Amount, params.ToAccountID, params.Amount)
		} else {
			result.FromAccount, result.ToAccount, err = updateAccountBalance(ctx, q, params.ToAccountID, params.Amount, params.FromAccountID, -params.Amount)
		}

		if err != nil {
			return err
		}

		return err
	})

	return result, err
}

func updateAccountBalance(ctx context.Context,
	q *Queries,
	account1Id int64,
	amount1 int64,
	account2Id int64,
	amount2 int64,
) (account1 Account, account2 Account, err error) {
	account1, err = q.AddAccountBalance(ctx, AddAccountBalanceParams{
		ID:     account1Id,
		Amount: amount1,
	})
	if err != nil {
		return
	}

	account2, err = q.AddAccountBalance(ctx, AddAccountBalanceParams{
		ID:     account2Id,
		Amount: amount2,
	})
	return
}
