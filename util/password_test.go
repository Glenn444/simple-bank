package util

import (
	"testing"

	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)


func TestPass(t *testing.T){
	password := RandomString(6)
	hashedPassword,hasherr := HashPassword(password)

	require.NoError(t,hasherr)
	require.NotEmpty(t,hashedPassword)

	compareError := CheckPassword(hashedPassword,password)
	require.NoError(t,compareError)
	

	password2 := RandomString(6)
	compareError1 := CheckPassword(hashedPassword,password2)
	require.EqualError(t,compareError1,bcrypt.ErrMismatchedHashAndPassword.Error())
}