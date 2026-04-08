package api

import (
	"database/sql"
	"fmt"
	"net/http"
	"strings"
	"time"

	db "github.com/Glenn444/banking-app/internal/database"
	"github.com/Glenn444/banking-app/util"
	"github.com/gin-gonic/gin"
	"github.com/lib/pq"
)

type CreateUserRequest struct {
	Username string `json:"username" binding:"required,alphanum"`
	FullName string `json:"full_name"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

type CreateUserResponse struct {
	Username          string    `json:"username"`
	FullName          string    `json:"full_name"`
	Email             string    `json:"email"`
	PasswordChangedAt time.Time `json:"password_changed_at"`
	CreatedAt         time.Time `json:"created_at"`
}

func (server *Server) createUser(ctx *gin.Context) {
	var req CreateUserRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	hashedPassword, err := util.HashPassword(req.Password)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	arg := db.CreateUsersParams{
		Username:       req.Username,
		FullName:       req.FullName,
		Email:          req.Email,
		HashedPassword: hashedPassword,
	}

	user, err := server.store.CreateUsers(ctx, arg)
	if err != nil {
		if pqError, ok := err.(*pq.Error); ok {
			switch pqError.Code.Name() {
			case "unique_violation":
				ctx.JSON(http.StatusForbidden, errorResponse(err))
				return
			}
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	resp := CreateUserResponse{
		Username:          user.Username,
		Email:             user.Email,
		FullName:          user.FullName,
		PasswordChangedAt: user.PasswordChangedAt,
		CreatedAt:         user.CreatedAt,
	}
	ctx.JSON(http.StatusOK, resp)

}

type SearchUserParams struct {
	Username string `form:"username" binding:"required,alphanum"`
}

func (server *Server) getUser(ctx *gin.Context) {
	var param SearchUserParams
	if err := ctx.ShouldBindQuery(&param); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	user, err := server.store.GetUser(ctx, param.Username)
	if err != nil {
		if pqError, ok := err.(*pq.Error); ok {
			switch pqError.Code.Name() {

			case "unique_violation":
				ctx.JSON(http.StatusForbidden, errorResponse(err))
			}
			fmt.Printf("databaseError: %s\n", pqError.Code.Name())
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	resp := CreateUserResponse{
		Username:          user.Username,
		Email:             user.Email,
		FullName:          user.FullName,
		PasswordChangedAt: user.PasswordChangedAt,
		CreatedAt:         user.CreatedAt,
	}
	ctx.JSON(http.StatusOK, resp)

}

type loginUserRequest struct {
	Username    string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type loginUserResponse struct {
	Username          string    `json:"username"`
	FullName          string    `json:"full_name"`
	Email             string    `json:"email"`
	PasswordChangedAt time.Time `json:"password_changed_at"`
	CreatedAt         time.Time `json:"created_at"`
	AccessToken       string    `json:"access_token"`
}

func (server *Server)loginUser(ctx *gin.Context){
	var req loginUserRequest
	err := ctx.ShouldBindJSON(&req)
	if err != nil{
		ctx.JSON(http.StatusBadRequest,errorResponse(err))
		return
	}

	user, err := server.store.GetUser(ctx,req.Username)
	if err != nil{
		// possible errors, 1. user not found
		if err == sql.ErrNoRows{
			ctx.JSON(http.StatusNotFound, errorResponse(err))
			return
		}
		//2. something happened to the database
		ctx.JSON(http.StatusInternalServerError,errorResponse(err))
	}

	//check user password against saved db password
	err = util.CheckPassword(user.HashedPassword,req.Password)
	if err != nil{
		ctx.JSON(http.StatusUnauthorized,errorMessage("Invalid username or password"))
		return
	}

	//create the token
	access_token,err := server.tokenMaker.CreateToken(req.Username,server.config.AcessTokenDuration)
	if err != nil{
		ctx.JSON(http.StatusInternalServerError,errorResponse(err))
	}

	resp := loginUserResponse{
		Username: req.Username,
		FullName: user.FullName,
		Email: user.Email,
		PasswordChangedAt: user.PasswordChangedAt,
		CreatedAt: user.CreatedAt,
		AccessToken: access_token,
	}

	ctx.JSON(http.StatusOK,resp)

}


type authHeader struct{
	Authorization string `header:"Authorization" binding:"required"`
}

type allUsersResponse struct {
	Username          string    `json:"username"`
	FullName          string    `json:"full_name"`
	Email             string    `json:"email"`
	PasswordChangedAt time.Time `json:"password_changed_at"`
	CreatedAt         time.Time `json:"created_at"`
}


//get all users in the app
func (server *Server)getAllUsers(ctx *gin.Context){

	var req authHeader

	err := ctx.ShouldBindHeader(&req)
	if err != nil{
		ctx.JSON(http.StatusUnauthorized,errorResponse(err))
		return
	}
	authParts := strings.Split(req.Authorization," ")
	
	if len(authParts) != 2 || authParts[0] != "Bearer"{
		ctx.JSON(http.StatusUnauthorized,errorMessage("invalid or missing authorization header"))
		return
	}
	authorizationBearerToken := authParts[1]
	//verify that the token is valid
	_, err = server.tokenMaker.VerifyToken(authorizationBearerToken)
	if err != nil{
		ctx.JSON(http.StatusUnauthorized,errorResponse(err))
		return
	}
	
	users,err := server.store.GetAllUsers(ctx)
	if err != nil{
		ctx.JSON(http.StatusInternalServerError,errorResponse(err))
		return
	}

	var allUsers []allUsersResponse
	for _,user:= range users{
		gotUser := allUsersResponse{
			Username: user.Username,
			FullName: user.FullName,
			Email: user.Email,
			PasswordChangedAt: user.PasswordChangedAt,
			CreatedAt: user.CreatedAt,
		}
		allUsers = append(allUsers, gotUser)
	}


	ctx.JSON(http.StatusOK,allUsers)
}