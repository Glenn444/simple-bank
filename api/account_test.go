package api

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	db "github.com/Glenn444/banking-app/internal/database"
	mock_database "github.com/Glenn444/banking-app/internal/database/mock"
	"github.com/Glenn444/banking-app/util"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)


func TestGetAccountApi(t *testing.T)  {
	account := randomAccount()

	ctrl := gomock.NewController(t)
	//defer ctrl.Finish()
	
	store := mock_database.NewMockStore(ctrl)

	//build stubs
	store.EXPECT().
		GetAccount(gomock.Any(),gomock.Eq(account.ID)).
		Times(1).
		Return(account,nil)

	//start test server and send request
	server := NewServer(store)
	recorder := httptest.NewRecorder()

	url := fmt.Sprintf("/accounts/%s",account.ID)
	request,err := http.NewRequest(http.MethodGet,url,nil)
	require.NoError(t,err)

	server.router.ServeHTTP(recorder,request)

	//check response
	require.Equal(t,http.StatusOK,recorder.Code)
}

func randomAccount() db.Account{
	return db.Account{
		ID: uuid.New(),
		Owner: util.RandomOwner(),
		Balance: util.RandomMoney(),
		Currency: util.RandomCurrency(),
	}
}