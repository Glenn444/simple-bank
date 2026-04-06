package token

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

var (
	ErrInvalidToken = errors.New("invalid token")
	ErrExpiredToken = errors.New("token is expired")
)


// Payload contains the payload data of the token
type Payload struct {
	jwt.RegisteredClaims
	ID        uuid.UUID
	Username  string    `json:"username"`
	ExpiredAt time.Time `json:"expired_at"`
	CreatedAt time.Time `json:"created_at"`
}


// NewPayload creates a new token payload with a specific username and duration
func NewPayload(username string,duration time.Duration)(*Payload,error){
	newUUID,err := uuid.NewRandom()
	if err != nil{
		return nil,err
	}
fmt.Printf("generated uuid: %v\n",newUUID.String())
	payload :=  &Payload{
		RegisteredClaims: jwt.RegisteredClaims{
			ID: newUUID.String(),
			Subject: username,
			IssuedAt: jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(duration)),
		},
		Username:username,
		ID: newUUID,
	}

	return payload,nil
}
