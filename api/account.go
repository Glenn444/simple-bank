package api

import (
	"database/sql"
	"net/http"

	db "github.com/Glenn444/banking-app/internal/database"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type createAccountRequest struct{
	Owner string `json:"owner" binding:"required"`
	Currency string `json:"currency" binding:"required,currency"`
}


func (server *Server) createAccount(ctx *gin.Context){
	var req createAccountRequest
	if err := ctx.ShouldBindJSON(&req);err != nil{
		ctx.JSON(http.StatusBadRequest,errorResponse(err))
		return
	}

	arg := db.CreateAccountParams{
		Owner: req.Owner,
		Currency: req.Currency,
		Balance: decimal.Zero,
	}

	acc,err := server.store.CreateAccount(ctx,arg)
	if err != nil{
		ctx.JSON(http.StatusInternalServerError,errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK,acc)

}

type AccountId struct{
	ID string `uri:"id" binding:"required,uuid"`
}

func (server *Server)getAccountById(ctx *gin.Context){

	var accountId AccountId
	
	if err := ctx.ShouldBindUri(&accountId); err != nil{
		ctx.JSON(http.StatusBadRequest,errorResponse(err))
		return
	}

	argId,_ := uuid.Parse(accountId.ID)
	acc,err := server.store.GetAccount(ctx,argId)
	if err != nil{
		if err == sql.ErrNoRows{
			ctx.JSON(http.StatusNotFound,errorResponse(err))
			return
		}
		ctx.JSON(http.StatusInternalServerError,errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK,acc)
}


type listAllAccountsParams struct{
	PageNum int32 `form:"page_num" binding:"required,min=1"` //offset
	PageSize int32 `form:"page_size" binding:"required,min=5,max=10"` //limit
}

func (server *Server)listAllAccounts(ctx *gin.Context){
	var req listAllAccountsParams
	
	if err := ctx.ShouldBindQuery(&req); err != nil{
		ctx.JSON(http.StatusBadRequest,errorResponse(err))
		return
	}

	arg := db.ListAccountsParams{
		Limit: req.PageSize,
		Offset: (req.PageNum - 1) * req.PageSize,
	}
	
	accs,err := server.store.ListAccounts(ctx,arg)
	if err != nil{
		ctx.JSON(http.StatusInternalServerError,errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK,accs)
}