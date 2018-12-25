package main

import (
	"github.com/stellar/go/build"
	"github.com/stellar/go/keypair"
	"log"
)

type guestAccount struct {
	keyPair *keypair.Full
}

func newGuestAccount() (*guestAccount, error) {
	keyPair, err := keypair.Random()
	if err != nil {
		return nil, err
	}
	return &guestAccount{
		keyPair: keyPair,
	}, nil
}

func (guest *guestAccount) createRatchetTx(
		hostRatchetAddress string,
		paymentTime uint64,
		roundSequenceNumber int,
	) (
		*build.TransactionEnvelopeBuilder,
		error,
	) {

	tx, err := build.Transaction(
		build.TestNetwork,
		build.SourceAccount{AddressOrSeed: hostRatchetAddress},
		build.Sequence{Sequence: uint64(guest.loadSequenceNumber()) + 1},
		build.Timebounds{
			MaxTime: paymentTime + defaultFinalityDelay + defaultMaxRoundDuration,
		},
		build.BumpSequence(build.BumpTo(roundSequenceNumber + 1)),
	)
	if err != nil {
		return nil, err
	}

	txe, err := tx.Sign(guest.keyPair.Seed())
	if err != nil {
		return nil, err
	}

	return &txe, nil
}

func (guest *guestAccount) receiveChannelProposeMsg(msg *ChannelProposeMsg) (*ChannelAcceptMsg, error) {
	baseSequenceNumber, err := loadSequenceNumber(msg.ChannelID)
	if err != nil {
		return nil, err
	}
	rsn := roundSequenceNumber(baseSequenceNumber, 1)

	ratchetTx, err := guest.createRatchetTx(msg.HostRatchetAccount, msg.FundingTime, rsn)
	if err != nil {
		return nil, err
	}

	return &ChannelAcceptMsg{
		ChannelID: msg.ChannelID,
		GuestRatchetRound1Sig: ratchetTx,
		// GuestSettleOnlyWithHostSig
	}, nil
}

func (guest *guestAccount) loadSequenceNumber() int {
	sequenceNumber, err := loadSequenceNumber(guest.keyPair.Address())
	if err != nil {
		log.Fatal(err)
	}
	return sequenceNumber
}