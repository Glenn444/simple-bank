package token

import (
	"time"

	"github.com/google/uuid"
)

// Payload contains the payload data of the token
type Payload struct {
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

	payload :=  &Payload{
		ID: newUUID,
		Username: username,
		ExpiredAt: time.Now().Add(duration),
		CreatedAt: time.Now(),
	}

	return payload,nil
}
