package db

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

type TxResult struct {
	Result    TransferTxResult
	Iteration int
}

func TestCreateTransferTransaction(t *testing.T) {
	store := NewStore(sqlDB)
	account1, account2 := createRandomAccount(t), createRandomAccount(t)
	n := 5
	amount := int64(10)

	errs := make(chan error)
	results := make(chan TransferTxResult)

	for i := 0; i < n; i++ {
		go func(it int) {
			tx, err := store.TransferTX(context.Background(), TransferTxParams{
				FromAccountID: account1.ID,
				ToAccountID:   account2.ID,
				Amount:        amount,
			})

			errs <- err
			results <- tx
		}(i)
	}

	existed := make(map[int]bool)
	for i := 0; i < n; i++ {
		err := <-errs

		require.NoError(t, err)

		result := <-results

		// validate transfer
		require.NotEmpty(t, result.Transfer)
		require.NotZero(t, result.Transfer.ID)
		require.NotZero(t, result.Transfer.CreatedAt)
		require.Equal(t, account1.ID, result.Transfer.FromAccountID)
		require.Equal(t, account2.ID, result.Transfer.ToAccountID)

		_, err = store.GetTransfer(context.Background(), result.Transfer.ID)
		require.NoError(t, err)

		// validate from entry
		require.NotEmpty(t, result.FromEntry)
		require.NotZero(t, result.FromEntry.ID)
		require.NotZero(t, result.FromEntry.CreatedAt)
		require.Equal(t, int64(-10), result.FromEntry.Amount)
		require.Equal(t, account1.ID, result.FromEntry.AccountID)

		_, err = store.GetEntry(context.Background(), result.FromEntry.ID)
		require.NoError(t, err)

		// validate to entry
		require.NotEmpty(t, result.ToEntry)
		require.NotZero(t, result.ToEntry.ID)
		require.NotZero(t, result.ToEntry.CreatedAt)
		require.Equal(t, int64(10), result.ToEntry.Amount)
		require.Equal(t, account2.ID, result.ToEntry.AccountID)

		_, err = store.GetEntry(context.Background(), result.ToEntry.ID)
		require.NoError(t, err)

		// validate accounts
		fromAccount := result.FromAccount
		require.NotEmpty(t, fromAccount)
		require.Equal(t, account1.ID, fromAccount.ID)

		toAccount := result.ToAccount
		require.NotEmpty(t, toAccount)
		require.Equal(t, account2.ID, toAccount.ID)

		diff1 := account1.Balance - fromAccount.Balance
		diff2 := toAccount.Balance - account2.Balance
		require.Equal(t, diff1, diff2)
		require.True(t, diff1 > 0)
		require.True(t, diff1%amount == 0) // 1 * amount, 2 * amount, 3 * amount, ..., n * amount

		k := int(diff1 / amount)
		require.True(t, k >= 1 && k <= n)
		require.NotContains(t, existed, k)
		existed[k] = true
	}

	fromAccount, err := testQueries.GetAccount(context.Background(), account1.ID)
	require.NoError(t, err)
	require.NotEmpty(t, fromAccount)
	require.Equal(t, account1.Balance-amount*int64(n), fromAccount.Balance)

	toAccount, err := testQueries.GetAccount(context.Background(), account2.ID)
	require.NoError(t, err)
	require.NotEmpty(t, toAccount)
	require.Equal(t, account2.Balance+amount*int64(n), toAccount.Balance)
}

func TestCreateTransferTransactionDeadlock(t *testing.T) {
	store := NewStore(sqlDB)
	account1, account2 := createRandomAccount(t), createRandomAccount(t)
	fmt.Printf(">> Account1 \n- ID: %d, Balance: %d\n", account1.ID, account1.Balance)
	fmt.Printf(">> Account1 \n- ID: %d, Balance: %d\n", account2.ID, account2.Balance)
	n := 10
	amount := int64(10)

	errs := make(chan error)

	var fromAccountId int64
	var toAccountId int64
	for i := 0; i < n; i++ {
		if i%2 == 1 {
			fromAccountId = account1.ID
			toAccountId = account2.ID
		} else {
			fromAccountId = account2.ID
			toAccountId = account1.ID
		}

		go func(it int, fromAccountId int64, toAccountId int64) {
			_, err := store.TransferTX(context.Background(), TransferTxParams{
				FromAccountID: fromAccountId,
				ToAccountID:   toAccountId,
				Amount:        amount,
			})

			errs <- err
		}(i, fromAccountId, toAccountId)
	}

	for i := 0; i < n; i++ {
		err := <-errs

		require.NoError(t, err)
	}

	fromAccount, err := testQueries.GetAccount(context.Background(), account1.ID)
	require.NoError(t, err)
	require.NotEmpty(t, fromAccount)
	require.Equal(t, account1.Balance, fromAccount.Balance)

	toAccount, err := testQueries.GetAccount(context.Background(), account2.ID)
	require.NoError(t, err)
	require.NotEmpty(t, toAccount)
	require.Equal(t, account2.Balance, toAccount.Balance)
}
