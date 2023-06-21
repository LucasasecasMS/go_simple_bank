package db

import (
	"context"
	"database/sql"
	"testing"

	"github.com/LucasasecasMS/go_simple_bank/util"
	"github.com/stretchr/testify/require"
)

func createRandomAccount(t *testing.T) Account {
	arg := CreateAccountParams{
		Owner:    util.RandomOwner(),
		Balance:  util.RandomMoney(),
		Currency: util.RandomCurrency(),
	}

	account, err := testQueries.CreateAccount(context.Background(), arg)

	require.NoError(t, err)
	require.NotEmpty(t, account)

	require.Equal(t, arg.Owner, account.Owner)
	require.Equal(t, arg.Balance, account.Balance)
	require.Equal(t, arg.Currency, account.Currency)

	require.NotZero(t, account.ID)
	require.NotZero(t, account.CreatedAt)

	return account
}

func TestCreateAccount(t *testing.T) {
	createRandomAccount(t)
}

func TestUpdateAccount(t *testing.T) {
	account := createRandomAccount(t)

	params := UpdateAccountParams{
		Balance: int64(float64(account.Balance) * 0.5),
		ID:      account.ID,
	}

	updatedAcc, err := testQueries.UpdateAccount(context.Background(), params)

	require.NoError(t, err)
	require.Equal(t, account.ID, updatedAcc.ID)
	require.Equal(t, account.Owner, updatedAcc.Owner)
	require.Equal(t, account.CreatedAt, updatedAcc.CreatedAt)
	require.Equal(t, account.Currency, updatedAcc.Currency)
	require.Equal(t, params.Balance, updatedAcc.Balance)
	require.Equal(t, params.ID, updatedAcc.ID)
}

func TestGetAccount(t *testing.T) {
	account := createRandomAccount(t)

	retrivedAccount, err := testQueries.GetAccount(context.Background(), account.ID)

	require.NoError(t, err)
	require.Equal(t, account.ID, retrivedAccount.ID)
	require.Equal(t, account.Owner, retrivedAccount.Owner)
	require.Equal(t, account.CreatedAt, retrivedAccount.CreatedAt)
	require.Equal(t, account.Currency, retrivedAccount.Currency)
	require.Equal(t, account.Balance, retrivedAccount.Balance)
	require.Equal(t, account.ID, retrivedAccount.ID)
}

func TestDeleteAccount(t *testing.T) {
	account := createRandomAccount(t)

	err := testQueries.DeleteAccount(context.Background(), account.ID)

	deletedAccount, err2 := testQueries.GetAccount(context.Background(), account.ID)

	require.NoError(t, err)
	require.Error(t, err2)
	require.EqualError(t, err2, sql.ErrNoRows.Error())
	require.Empty(t, deletedAccount)
}

func TestListAccounts(t *testing.T) {
	for i := 0; i < 10; i++ {
		createRandomAccount(t)
	}

	arg := ListAccountsParams{
		Limit:  5,
		Offset: 0,
	}

	accounts, err := testQueries.ListAccounts(context.Background(), arg)
	require.NoError(t, err)
	require.Len(t, accounts, 5)

	for _, account := range accounts {
		require.NotEmpty(t, account)
	}
}
