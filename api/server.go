package api

import (
	"fmt"
	"net/http"

	db "github.com/Glenn444/banking-app/internal/database"
	"github.com/Glenn444/banking-app/internal/token"
	"github.com/Glenn444/banking-app/util"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
)

// Server serves HTTP requests for our banking service
type Server struct{
	config util.Config
	tokenMaker token.Maker
	store db.Store
	router *gin.Engine
}

func NewServer(config util.Config,store db.Store) (*Server,error){
	jwtTokenMaker,err := token.NewJWTMaker(config.TokenSymmetricKey)
	if err != nil{
		return nil,fmt.Errorf("cannot create token maker %w\n",err)
	}
	server := &Server{
		tokenMaker: jwtTokenMaker,
		store: store,
		config: config,
	}

	 // Force log's color
    gin.ForceConsoleColor()
	router := gin.Default()

	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
    v.RegisterValidation("currency", validCurrency)
  }
  
	//add routes to router
	router.GET("/",server.welcome)

	router.POST("/accounts", server.createAccount)
	router.GET("/accounts/:id",server.getAccountById)
	router.GET("/accounts", server.listAllAccounts)

	router.POST("/transfers", server.createTransfer)

	router.POST("/user", server.createUser)
	router.GET("/user", server.getUser)
	router.GET("/users",server.getAllUsers)

	router.POST("/login", server.loginUser)

	server.router = router
	return server,nil
}

//start runs the HTTP server on a specific address
func (server *Server) Start(address string) error{
	return server.router.Run(address)
}


func errorResponse(err error)gin.H{
	return  gin.H{"error":err.Error()}
}

func errorMessage(message string)gin.H{
	return gin.H{"error:":message}
}

func (server *Server) welcome(ctx *gin.Context){
	ctx.JSON(http.StatusOK,"Welcome to the Server")

}