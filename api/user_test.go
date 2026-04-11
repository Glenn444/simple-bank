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
	"github.com/Glenn444/banking-app/util"
	"github.com/lib/pq"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestCreateUser(t *testing.T) {

	userParams := CreateUserRequest{
		Username: util.RandomOwner(),
		Password: util.RandomString(8),
		FullName: util.RandomOwner(),
		Email:    util.RandomEmail(),
	}

	user := randomUser()
	testCases := []struct {
		name          string
		buildStubs    func(store *mock_database.MockStore)
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			buildStubs: func(store *mock_database.MockStore) {
				store.EXPECT().CreateUsers(gomock.Any(), gomock.Any()).
					Times(1).
					Return(user, nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
			},
		},
		{
			name: "InternalServerError",
			buildStubs: func(store *mock_database.MockStore) {
				store.
					EXPECT().
					CreateUsers(gomock.Any(), gomock.Any()).
					Times(1).
					Return(database.User{}, sql.ErrConnDone) //simulate DB failure
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name: "DuplicateUsername",
			buildStubs: func(store *mock_database.MockStore) {
				store.
					EXPECT().
					CreateUsers(gomock.Any(), gomock.Any()).
					Times(1).
					Return(database.User{}, &pq.Error{Code: "23505"}) //postgres unique violation
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusForbidden, recorder.Code)
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

			server := newTestServer(t, store)
			recorder := httptest.NewRecorder()

			urlPath := "/user"
			body, err := json.Marshal(userParams)
			require.NoError(t, err)

			request, err := http.NewRequest(http.MethodPost, urlPath, bytes.NewReader(body))

			request.Header.Set("Content-Type", "application/json")

			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(t, recorder)
		})
	}
}

// func hashedPassword(t *testing.T,pass string) string {
// 	//password := util.RandomString(8)
// 	hash, err := util.HashPassword(pass)
// 	require.NoError(t, err)

// 	return hash
// }

func TestGetUser(t *testing.T) {
	user := randomUser()

	testCases := []struct {
		name          string
		urlPath       string
		buildStubs    func(store *mock_database.MockStore)
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			urlPath: fmt.Sprintf("/user?username=%s", user.Username),
			buildStubs: func(store *mock_database.MockStore) {
				store.
					EXPECT().GetUser(gomock.Any(), gomock.Eq(user.Username)).
					Times(1).
					Return(user, nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
			},
		},
		{
			name: "NotFound",
			urlPath: fmt.Sprintf("/user?username=%s", user.Username),
			buildStubs: func(store *mock_database.MockStore) {
				store.
					EXPECT().
					GetUser(gomock.Any(), gomock.Eq(user.Username)).
					Times(1).
					Return(db.User{}, sql.ErrNoRows)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusNotFound, recorder.Code)
			},
		},
		{
			name: "BadRequest",
			urlPath: "/user",
			buildStubs: func(store *mock_database.MockStore) {
				store.
					EXPECT().
					GetUser(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "InternalServerError",
			urlPath: fmt.Sprintf("/user?username=%s",user.Username),
			buildStubs: func(store *mock_database.MockStore) {
				store.
				EXPECT().
				GetUser(gomock.Any(),gomock.Eq(user.Username)).
				Times(1).
				Return(db.User{},sql.ErrConnDone)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t,http.StatusInternalServerError,recorder.Code)
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

			server := newTestServer(t, store)

			
			request, err := http.NewRequest(http.MethodGet, tc.urlPath, nil)
			require.NoError(t, err)

			recorder := httptest.NewRecorder()

			server.router.ServeHTTP(recorder, request)

			tc.checkResponse(t, recorder)
		})

	}
}

func randomUser() db.User {
	return database.User{
		Username:          util.RandomOwner(),
		FullName:          util.RandomOwner(),
		Email:             util.RandomEmail(),
		PasswordChangedAt: time.Now(),
		CreatedAt:         time.Now(),
	}
}
