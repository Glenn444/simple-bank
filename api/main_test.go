package api

import (
	"os"
	"testing"
	"time"

	"github.com/Glenn444/banking-app/internal/database"
	"github.com/Glenn444/banking-app/util"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func newTestServer(t *testing.T,store database.Store) *Server{
	config := util.Config{
		TokenSymmetricKey: util.RandomString(32),
		AcessTokenDuration: time.Minute,
	}

	server,err := NewServer(config,store)
	require.NoError(t,err)

	return server
}

func TestMain(m *testing.M)  {

	gin.SetMode(gin.TestMode)
	
	os.Exit(m.Run())
}