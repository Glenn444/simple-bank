package token

import (
	"testing"
	"time"

	"github.com/Glenn444/banking-app/util"
	"github.com/stretchr/testify/require"
)


func TestJWTMaker(t *testing.T){
	m, err := NewJWTMaker(util.RandomString(32))
	require.NoError(t,err)

	username := util.RandomOwner()
	duration := time.Minute

	issuedAt := time.Now()
	expireAt := time.Now().Add(duration)

	token,errToken := m.CreateToken(username,duration)
	require.NoError(t,errToken)
	require.NotEmpty(t,token)

	payload,errPayload := m.VerifyToken(token)
	require.NoError(t,errPayload)
	require.NotEmpty(t,payload)

	require.NotZero(t,payload.ID)
	require.Equal(t,username,payload.Username)
	require.WithinDuration(t,issuedAt,payload.IssuedAt.Time,time.Second)
	require.WithinDuration(t,expireAt,payload.ExpiresAt.Time,time.Second)

}