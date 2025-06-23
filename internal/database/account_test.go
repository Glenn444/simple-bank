package database

import "testing"

func TestCreateAccount(t *testing.T){
	arg := CreateAccountParams{
		Owner: "tom",
		Balance: 100,
	}
}