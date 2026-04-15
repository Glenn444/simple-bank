package api

import (
	"net/http"
	"net/http/httptest"
	"testing"

	mock_database "github.com/Glenn444/banking-app/internal/database/mock"
	"github.com/Glenn444/banking-app/internal/token"
)


func TestCreateTransferApi(t *testing.T){
	testCases := []struct{
		name string
		setupAuth func(t *testing.T,request *http.Request,tokenMaker token.Maker)
		buildStubs func(store *mock_database.MockStore)
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{}

	for i,_:= range testCases{
		tc := testCases[i]

		t.Run(tc.name,func(t *testing.T) {
			
		})
	}
}