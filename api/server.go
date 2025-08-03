package api

import (
	"net/http"

	db "github.com/Glenn444/banking-app/internal/database"
	"github.com/gin-gonic/gin"
)

// Server serves HTTP requests for our banking service
type Server struct{
	store *db.Store
	router *gin.Engine
}

func NewServer(store *db.Store) *Server{
	server := &Server{store: store}
	router := gin.Default()

	//add routes to router
	router.GET("/",server.welcome)

	router.POST("/accounts", server.createAccount)
	router.GET("/accounts/:id",server.getAccountById)
	router.GET("/accounts", server.listAllAccounts)


	server.router = router
	return server
}

//start runs the HTTP server on a specific address
func (server *Server) Start(address string) error{
	return server.router.Run(address)
}


func errorResponse(err error)gin.H{
	return  gin.H{"error":err.Error()}
}

func (server *Server) welcome(ctx *gin.Context){
	ctx.JSON(http.StatusOK,"Welcome to the Server")

}