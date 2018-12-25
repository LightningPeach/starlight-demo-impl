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
	keyPair, err := createAccount()
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
		build.BumpSequence(build.BumpTo(roundSequenceNumber+1)),
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

func (guest *guestAccount) createSettleOnlyWithHostTx(
	hostAddress,
	escrowAddress,
	guestRatchetAccount,
	hostRatchetAccount string,
	fundingTime uint64,
	roundSequenceNumber int,
) (
	*build.TransactionEnvelopeBuilder,
	error,
) {

	tx, err := build.Transaction(
		build.TestNetwork,
		build.SourceAccount{AddressOrSeed: escrowAddress},
		build.Sequence{Sequence: uint64(roundSequenceNumber) + 2},
		build.Timebounds{
			MinTime: fundingTime + 2*defaultFinalityDelay + defaultMaxRoundDuration,
		},
		build.AccountMerge(
			build.SourceAccount{AddressOrSeed: escrowAddress},
			build.Destination{AddressOrSeed: hostAddress},
		),
		build.AccountMerge(
			build.SourceAccount{AddressOrSeed: guestRatchetAccount},
			build.Destination{AddressOrSeed: hostAddress},
		),
		build.AccountMerge(
			build.SourceAccount{AddressOrSeed: hostRatchetAccount},
			build.Destination{AddressOrSeed: hostAddress},
		),
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

	settleOnlyWithHostTx, err := guest.createSettleOnlyWithHostTx(
		msg.HostAccount,
		msg.ChannelID,
		msg.GuestRatchetAccount,
		msg.HostRatchetAccount,
		msg.FundingTime,
		rsn,
	)

	return &ChannelAcceptMsg{
		ChannelID:                  msg.ChannelID,
		GuestRatchetRound1Sig:      ratchetTx,
		GuestSettleOnlyWithHostSig: settleOnlyWithHostTx,
	}, nil
}

func (guest *guestAccount) loadSequenceNumber() int {
	sequenceNumber, err := loadSequenceNumber(guest.keyPair.Address())
	if err != nil {
		log.Fatal(err)
	}
	return sequenceNumber
}
