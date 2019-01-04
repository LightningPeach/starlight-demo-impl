package main

import "github.com/stellar/go/keypair"

type htlcResolutionAccount struct {
	keyPair *keypair.Full
}

func newHTLCResolutionAccount() (*htlcResolutionAccount, error) {
	keyPair, err := keypair.Random()
	if err != nil {
		return nil, err
	}
	return &htlcResolutionAccount{
		keyPair: keyPair,
	}, nil
}