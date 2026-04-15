package api

import (
	"database/sql"
	"fmt"
	"net/http"

	db "github.com/Glenn444/banking-app/internal/database"
	"github.com/Glenn444/banking-app/internal/token"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type transferMoneyRequest struct {
	FromAccountID uuid.UUID       `json:"from_account_id" binding:"required"`
	ToAccountID   uuid.UUID       `json:"to_account_id" binding:"required"`
	Amount        decimal.Decimal `json:"amount" bindings:"required,gt=0"`
	Currency      string          `json:"currency" binding:"required,currency"`
}

func (server *Server) createTransfer(ctx *gin.Context) {
	var req transferMoneyRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	fromAccount,valid := server.validAccount(ctx,req.FromAccountID,req.Currency)
	if !valid{
		return
	}

	authPayload := ctx.MustGet(authorizationPayloadKey).(*token.Payload)
	//check if the from_account is valid
	if fromAccount.Owner != authPayload.Username {
		ctx.JSON(http.StatusForbidden,errorMessage("from account doesn't belong to you"))
		return
	}

	//check if the to_account is valid
	_,valid = server.validAccount(ctx,req.ToAccountID,req.Currency)
	if !valid{
		return
	}

	arg := db.TransferTxParams{
		FromAccountID: req.FromAccountID,
		ToAccountID:   req.ToAccountID,
		Amount:        req.Amount,
	}

	result, err := server.store.TransferTx(ctx,arg)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	ctx.JSON(http.StatusOK, result)
}

//check if to_account and from_account have matching currency type
func (server *Server) validAccount(ctx *gin.Context, accounID uuid.UUID, currency string) (db.Account,bool) {
	account, err := server.store.GetAccount(ctx, accounID)
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, errorResponse(err))
			return account,false
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return account,false
	}

	if account.Currency != currency {
		err := fmt.Errorf("account [%v] currency mismatch: %s vs %s", account.ID, account.Currency, currency)
		ctx.JSON(http.StatusBadRequest,errorResponse(err))
		return account,false
	}
	return account,true
}
