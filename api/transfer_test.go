package api

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Glenn444/banking-app/internal/database"
	db "github.com/Glenn444/banking-app/internal/database"
	mock_database "github.com/Glenn444/banking-app/internal/database/mock"
	"github.com/Glenn444/banking-app/internal/token"
	"github.com/Glenn444/banking-app/util"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestCreateTransferApi(t *testing.T) {
	//Create two accounts with matching currency
	fromAccount := randomAccountWithCurrency("USD")
	toAccount := randomAccountWithCurrency("USD")
	wrongCurrencyAccount := randomAccountWithCurrency("EUR")

	amount := decimal.NewFromFloat(100)

	//A user whose token we'll forge for auth
	user := randomUser()

	//Make fromAcccount owned by the auth user
	fromAccount.Owner = user.Username

	testCases := []struct {
		name          string
		body          gin.H
		setupAuth     func(t *testing.T, request *http.Request, tokenMaker token.Maker)
		buildStubs    func(store *mock_database.MockStore)
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			body: gin.H{
				"from_account_id": fromAccount.ID,
				"to_account_id":   toAccount.ID,
				"amount":          amount,
				"currency":        "USD",
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.Username, time.Minute)
			},
			buildStubs: func(store *mock_database.MockStore) {
				//validate that the from Account is a valid account
				store.EXPECT().
					GetAccount(gomock.Any(),gomock.Eq(fromAccount.ID)).
					Times(1).
					Return(fromAccount,nil)

				//validate that the toAccount is valid
				store.EXPECT().
					GetAccount(gomock.Any(),gomock.Eq(toAccount.ID)).
					Times(1).
					Return(toAccount,nil)

				//if both are valid, perfom the transfer
				expectedArgs := db.TransferTxParams{
					FromAccountID: fromAccount.ID,
					ToAccountID: toAccount.ID,
					Amount: amount,
				}
				store.EXPECT().
					TransferTx(gomock.Any(),EqTransferParams(expectedArgs)).
					Times(1).
					Return(db.TransferTxResult{},nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
			},
		},
		{
			name: "UnauthorizedUser",
			body: gin.H{
				"from_account_id":fromAccount.ID,
				"to_account_id":toAccount.ID,
				"amount":amount,
				"currency":"USD",
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				//pass in a different user not owner of the fromAccount
				addAuthorization(t,request,tokenMaker,authorizationTypeBearer,"other_user",time.Minute)
			},
			buildStubs: func(store *mock_database.MockStore) {
				store.EXPECT().
					GetAccount(gomock.Any(),gomock.Eq(fromAccount.ID)).
					Times(1).
					Return(fromAccount,nil)

				// GetAccount for toAccount and TransferTx should never be called
				store.EXPECT().GetAccount(gomock.Any(),gomock.Eq(toAccount.ID)).Times(0)
				store.EXPECT().TransferTx(gomock.Any(),gomock.Any()).Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t,http.StatusForbidden,recorder.Code)
			},
		},
		{
			name: "NoAuthorization",
			body: gin.H{
				"from_account_id": fromAccount.ID,
				"to_account_id":   toAccount.ID,
				"amount":          amount,
				"currency":        "USD",
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				// don't add any auth header
			},
			buildStubs: func(store *mock_database.MockStore) {
				store.EXPECT().GetAccount(gomock.Any(), gomock.Any()).Times(0)
				store.EXPECT().TransferTx(gomock.Any(), gomock.Any()).Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		},
		{
			name: "FromAccountNotFound",
			body: gin.H{
				"from_account_id": fromAccount.ID,
				"to_account_id":   toAccount.ID,
				"amount":          amount,
				"currency":        "USD",
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.Username, time.Minute)
			},
			buildStubs: func(store *mock_database.MockStore) {
				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Eq(fromAccount.ID)).
					Times(1).
					Return(db.Account{}, sql.ErrNoRows)

				store.EXPECT().GetAccount(gomock.Any(), gomock.Eq(toAccount.ID)).Times(0)
				store.EXPECT().TransferTx(gomock.Any(), gomock.Any()).Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusNotFound, recorder.Code)
			},
		},
		{
			name: "ToAccountNotFound",
			body: gin.H{
				"from_account_id": fromAccount.ID,
				"to_account_id":   toAccount.ID,
				"amount":          amount,
				"currency":        "USD",
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.Username, time.Minute)
			},
			buildStubs: func(store *mock_database.MockStore) {
				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Eq(fromAccount.ID)).
					Times(1).
					Return(fromAccount, nil)

				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Eq(toAccount.ID)).
					Times(1).
					Return(db.Account{}, sql.ErrNoRows)

				store.EXPECT().TransferTx(gomock.Any(), gomock.Any()).Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusNotFound, recorder.Code)
			},
		},
		{
			name: "FromAccountCurrencyMismatch",
			body: gin.H{
				"from_account_id": wrongCurrencyAccount.ID,
				"to_account_id":   toAccount.ID,
				"amount":          amount,
				"currency":        "USD",
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.Username, time.Minute)
			},
			buildStubs: func(store *mock_database.MockStore) {
				// wrongCurrencyAccount is EUR, but request says USD → mismatch
				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Eq(wrongCurrencyAccount.ID)).
					Times(1).
					Return(wrongCurrencyAccount, nil)

				store.EXPECT().GetAccount(gomock.Any(), gomock.Eq(toAccount.ID)).Times(0)
				store.EXPECT().TransferTx(gomock.Any(), gomock.Any()).Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "ToAccountCurrencyMismatch",
			body: gin.H{
				"from_account_id": fromAccount.ID,
				"to_account_id":   wrongCurrencyAccount.ID,
				"amount":          amount,
				"currency":        "USD",
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.Username, time.Minute)
			},
			buildStubs: func(store *mock_database.MockStore) {
				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Eq(fromAccount.ID)).
					Times(1).
					Return(fromAccount, nil)

				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Eq(wrongCurrencyAccount.ID)).
					Times(1).
					Return(wrongCurrencyAccount, nil)

				store.EXPECT().TransferTx(gomock.Any(), gomock.Any()).Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "InvalidBody",
			body: gin.H{
				// missing required fields
				"currency": "USD",
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.Username, time.Minute)
			},
			buildStubs: func(store *mock_database.MockStore) {
				store.EXPECT().GetAccount(gomock.Any(), gomock.Any()).Times(0)
				store.EXPECT().TransferTx(gomock.Any(), gomock.Any()).Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "TransferTxInternalError",
			body: gin.H{
				"from_account_id": fromAccount.ID,
				"to_account_id":   toAccount.ID,
				"amount":          amount,
				"currency":        "USD",
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.Username, time.Minute)
			},
			buildStubs: func(store *mock_database.MockStore) {
				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Eq(fromAccount.ID)).
					Times(1).
					Return(fromAccount, nil)

				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Eq(toAccount.ID)).
					Times(1).
					Return(toAccount, nil)

				store.EXPECT().
					TransferTx(gomock.Any(), gomock.Any()).
					Times(1).
					Return(db.TransferTxResult{}, sql.ErrConnDone)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
	}

	for i := range testCases {
		tc := testCases[i]

		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			store := mock_database.NewMockStore(ctrl)
			tc.buildStubs(store)

			//start server and send request
			server := newTestServer(t, store)
			recorder := httptest.NewRecorder()

			boady,err := json.Marshal(tc.body)
			require.NoError(t,err)

			request,err := http.NewRequest(http.MethodPost,"/transfers",bytes.NewReader(boady))
			require.NoError(t,err)
			request.Header.Set("Content-Type","application/json")

			tc.setupAuth(t,request,server.tokenMaker)

			server.router.ServeHTTP(recorder,request)
			tc.checkResponse(t,recorder)

		})
	}
}

// helper - to build a random account with a given currency
func randomAccountWithCurrency(currency string) database.Account {
	return database.Account{
		ID:       uuid.New(),
		Owner:    util.RandomOwner(),
		Balance:  decimal.NewFromFloat(1000),
		Currency: currency,
	}
}


type eqTransferParamsMatcher struct {
    arg db.TransferTxParams
}

func (e eqTransferParamsMatcher) Matches(x interface{}) bool {
    arg, ok := x.(db.TransferTxParams)
    if !ok {
        return false
    }

    return arg.FromAccountID == e.arg.FromAccountID &&
        arg.ToAccountID == e.arg.ToAccountID &&
        arg.Amount.Equal(e.arg.Amount) // ← use decimal's own Equal method
}

func (e eqTransferParamsMatcher) String() string {
    return fmt.Sprintf("matches TransferTxParams %v", e.arg)
}

func EqTransferParams(arg db.TransferTxParams) gomock.Matcher {
    return eqTransferParamsMatcher{arg}
}