package token

import (
	"fmt"
	"time"

	"github.com/aead/chacha20poly1305"
	"github.com/o1egl/paseto"
)

type Paseto struct {
	paseto *paseto.V2
	key []byte
}

func NewPaseto(key string) (*Paseto, error) {
	if len(key) != chacha20poly1305.KeySize {
		return nil, fmt.Errorf("the length of key must be %v", chacha20poly1305.KeySize)
	}

	paseto := &Paseto{
		paseto: paseto.NewV2(),
		key: []byte(key),
	}

	return paseto, nil
} 

func (p *Paseto) CreateToken(username string, duration time.Duration) (string, error) {
	payload, err := NewPayload(username, duration)
	if err != nil {
		return "", err
	}
	return p.paseto.Encrypt(p.key, payload, nil)
}

func (p *Paseto) VerifyToken(token string) (*Payload, error) {
	payload := &Payload{}
	err := p.paseto.Decrypt(token, p.key, payload, nil)
	if err != nil {
		return nil, err
	}

	err = payload.isValid()
	if err != nil {
		return nil, err
	}

	return payload, nil
} 