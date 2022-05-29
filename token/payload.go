package token

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

type Payload struct {
	ID uuid.UUID `json:"id"`
	Username string `json:"username"`
	Issued_At time.Time `json:"issued_at"`
	Expired_At time.Time `json:"expired_at"`
}

var (
	ErrInvalidToken = errors.New("token is invalid")
	ErrExpiredToken = errors.New("token is expired")
)

func NewPayload(username string, duration time.Duration) (*Payload, error) {
	tokenId, err := uuid.NewRandom()
	if err != nil {
		return nil, err
	}

	payload := &Payload{
		ID: tokenId,
		Username: username,
		Issued_At: time.Now(),
		Expired_At: time.Now().Add(duration),
	}
	return payload, nil
}

func (p *Payload) isValid() error {
	if time.Now().After(p.Expired_At) {
		return ErrExpiredToken
	}
	return nil
}