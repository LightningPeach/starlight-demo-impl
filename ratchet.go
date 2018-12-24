package main

import "github.com/stellar/go/keypair"

type ratchetAccount struct {
	keyPair *keypair.Full
}

func newRatchetAccount() (*ratchetAccount, error) {
	keyPair, err := keypair.Random()
	if err != nil {
		return nil, err
	}
	return &ratchetAccount{
		keyPair: keyPair,
	}, nil
}