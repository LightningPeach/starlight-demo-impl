package main

import (
	"fmt"
	"github.com/stellar/go/build"
	"github.com/stellar/go/clients/horizon"
	"github.com/stellar/go/keypair"
	"log"
	"time"
)

type hostAccount struct {
	selfKeyPair *keypair.Full

	escrowKeyPair       *keypair.Full
	hostRatchetAccount  *ratchetAccount
	guestRatchetAccount *ratchetAccount
}

func newHostAccount() (*hostAccount, error) {
	selfKeyPair, err := createAccount()
	if err != nil {
		return nil, err
	}

	escrowKeyPair, err := keypair.Random()
	if err != nil {
		return nil, err
	}

	hostRatchetAccount, err := newRatchetAccount()
	if err != nil {
		return nil, err
	}

	guestRatchetAccount, err := newRatchetAccount()
	if err != nil {
		return nil, err
	}

	return &hostAccount{
		selfKeyPair:         selfKeyPair,
		escrowKeyPair:       escrowKeyPair,
		hostRatchetAccount:  hostRatchetAccount,
		guestRatchetAccount: guestRatchetAccount,
	}, nil
}

func (host *hostAccount) setupAccountTx() error {
	tx, err := build.Transaction(
		build.TestNetwork,
		build.SourceAccount{host.selfKeyPair.Address()},
		// build.AutoSequence{horizon.DefaultTestNetClient},
		build.Sequence{Sequence: uint64(host.loadSequenceNumber()) + 1},
		build.CreateAccount(
			build.Destination{host.escrowKeyPair.Address()},
			build.NativeAmount{"1"},
		),
		build.CreateAccount(
			build.Destination{host.hostRatchetAccount.keyPair.Address()},
			build.NativeAmount{"1"},
		),
		build.CreateAccount(
			build.Destination{host.guestRatchetAccount.keyPair.Address()},
			build.NativeAmount{"1"},
		),
	)
	if err != nil {
		return err
	}

	// Sign the transaction to prove you are actually the person sending it.
	txe, err := tx.Sign(host.selfKeyPair.Seed())
	if err != nil {
		return err
	}

	return host.publishTx(&txe)
}

// TODO(evg): minTime/maxTime
func (host *hostAccount) fundingTx(guestEscrowPubKey string) error {
	// TODO(evg): use blockchain timestamp instead of system time
	fundingTime := time.Now().Unix()
	// TODO(evg): adjust constants
	const (
		maxRoundDuration = 3600
		finalityDelay    = 3600
	)

	tx, err := build.Transaction(
		build.TestNetwork,
		build.SourceAccount{AddressOrSeed: host.selfKeyPair.Address()},
		build.Sequence{Sequence: uint64(host.loadSequenceNumber()) + 1},
		build.Timebounds{
			MaxTime: uint64(fundingTime + maxRoundDuration + finalityDelay),
		},
		// Escrow Account
		build.Payment(
			build.Destination{AddressOrSeed: host.escrowKeyPair.Address()},
			build.NativeAmount{Amount: "0.5"},
		),
		build.SetOptions(
			build.SourceAccount{AddressOrSeed: host.escrowKeyPair.Address()},
			build.SetThresholds(2, 2, 2),
			build.Signer{
				Address: guestEscrowPubKey,
				Weight:  1,
			},
		),
		// GuestRatchetAccount
		build.Payment(
			build.Destination{AddressOrSeed: host.guestRatchetAccount.keyPair.Address()},
			build.NativeAmount{Amount: "1"},
		),
		build.SetOptions(
			build.SourceAccount{AddressOrSeed: host.guestRatchetAccount.keyPair.Address()},
			build.MasterWeight(0),
			build.SetThresholds(2, 2, 2),
			build.Signer{
				Address: guestEscrowPubKey,
				Weight:  1,
			},
		),
		build.SetOptions(
			build.SourceAccount{AddressOrSeed: host.guestRatchetAccount.keyPair.Address()},
			build.Signer{
				Address: host.escrowKeyPair.Address(),
				Weight:  1,
			},
		),
		// HostRatchetAccount
		build.Payment(
			build.Destination{AddressOrSeed: host.hostRatchetAccount.keyPair.Address()},
			build.NativeAmount{Amount: "0.5"},
		),
		build.SetOptions(
			build.MasterWeight(0),
			build.Signer{
				Address: host.escrowKeyPair.Address(),
				Weight:  1,
			},
		),
	)
	if err != nil {
		return err
	}

	// Sign the transaction to prove you are actually the person sending it.
	txe, err := tx.Sign(host.selfKeyPair.Seed(), host.escrowKeyPair.Seed(), host.guestRatchetAccount.keyPair.Seed())
	if err != nil {
		return err
	}

	return host.publishTx(&txe)
}

func (host *hostAccount) cleanupTx() {}

func (host *hostAccount) publishTx(txe *build.TransactionEnvelopeBuilder) error {
	txeB64, err := txe.Base64()
	if err != nil {
		return err
	}

	// And finally, send it off to Stellar!
	resp, err := horizon.DefaultTestNetClient.SubmitTransaction(txeB64)
	if err != nil {
		return err
	}

	fmt.Println("Successful Transaction:")
	fmt.Println("Ledger:", resp.Ledger)
	fmt.Println("Hash:", resp.Hash)

	return nil
}

func (host *hostAccount) loadSequenceNumber() int {
	sequenceNumber, err := loadSequenceNumber(host.selfKeyPair.Address())
	if err != nil {
		log.Fatal(err)
	}
	return sequenceNumber
}