package database

import (
	"context"
	"testing"
	"time"

	"github.com/Glenn444/banking-app/util"
	"github.com/stretchr/testify/require"
)



//helper func to createRandom Users in the db
func createRandomUser(t TestingT)User{
	arg := CreateUsersParams{
		Username: util.RandomOwner(),
		HashedPassword:"randomstring",
		FullName: util.RandomOwner(),
		Email: util.RandomEmail(),
	}
	
	
	user, err := testQueries.CreateUsers(context.Background(),arg)
	require.NoError(t,err)
	require.NotEmpty(t,user)


	require.Equal(t,arg.Username,user.Username)
	require.Equal(t,arg.FullName,user.FullName)
	require.Equal(t,arg.Email,user.Email)
	require.Equal(t,arg.HashedPassword,user.HashedPassword)

	require.NotZero(t,user.PasswordChangedAt)
	require.NotZero(t,user.CreatedAt)

	return user
}

func TestCreateUser(t *testing.T){
	createRandomUser(t)
}
func TestGetUser(t *testing.T){
	user1 := createRandomUser(t)

	user2,err := testQueries.GetUser(context.Background(),user1.Username)
	require.NoError(t,err)
	require.NotEmpty(t,user2)

	require.Equal(t,user1.Username,user2.Username)
	require.Equal(t,user1.FullName,user2.FullName)
	require.Equal(t,user1.Email,user2.Email)
	require.WithinDuration(t,user1.PasswordChangedAt,user2.PasswordChangedAt,time.Second)
	require.WithinDuration(t,user1.CreatedAt,user2.CreatedAt,time.Second)
}

