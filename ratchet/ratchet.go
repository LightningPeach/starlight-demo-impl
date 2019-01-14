package ratchet

import "github.com/stellar/go/keypair"

type Account struct {
	KeyPair *keypair.Full
}

func New() (*Account, error) {
	keyPair, err := keypair.Random()
	if err != nil {
		return nil, err
	}
	return &Account{
		KeyPair: keyPair,
	}, nil
}