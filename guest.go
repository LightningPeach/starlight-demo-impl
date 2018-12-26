package main

import (
	"fmt"
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
	hostRatchetAddress,
	escrowAddress string,
	paymentTime uint64,
	roundSequenceNumber int,
) (
	*build.TransactionEnvelopeBuilder,
	error,
) {
	sequenceNumber, err := loadSequenceNumber(hostRatchetAddress)
	if err != nil {
		return nil, err
	}

	tx, err := build.Transaction(
		build.TestNetwork,
		build.SourceAccount{AddressOrSeed: hostRatchetAddress},
		build.Sequence{Sequence: uint64(sequenceNumber) + 1},
		build.Timebounds{
			MaxTime: paymentTime + defaultFinalityDelay + defaultMaxRoundDuration,
		},
		build.BumpSequence(
			build.SourceAccount{AddressOrSeed: escrowAddress},
			build.BumpTo(roundSequenceNumber+1),
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
	fmt.Println("createSettleOnlyWithHostTx.MinTime", fundingTime + 2*defaultFinalityDelay + defaultMaxRoundDuration)

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

	fmt.Println("createRatchetTx for host")
	ratchetTx, err := guest.createRatchetTx(msg.HostRatchetAccount, msg.ChannelID, msg.FundingTime, rsn)
	if err != nil {
		return nil, err
	}

	fmt.Println("createSettleOnlyWithHostTx")
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

func (guest *guestAccount) createRatchetTxForOffChainPayment(
	escrowAddress,
	hostRatchetAddress string,
	paymentTime uint64,
	rsn int64,
) (
	*build.TransactionEnvelopeBuilder,
	error,
) {

	sequence, err := loadSequenceNumber(hostRatchetAddress)
	if err != nil {
		return nil, err
	}

	tx, err := build.Transaction(
		build.TestNetwork,
		build.SourceAccount{AddressOrSeed: hostRatchetAddress},
		// RatchetAccount.SequenceNumber + 1
		build.Sequence{Sequence: uint64(sequence) + 1},
		build.Timebounds{
			MaxTime: paymentTime + defaultFinalityDelay + defaultMaxRoundDuration,
		},
		// Bump sequence of EscrowAccount to RoundSequenceNumber + 1
		build.BumpSequence(
			build.SourceAccount{AddressOrSeed: escrowAddress},
			build.BumpTo(rsn + 1),
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

func (guest *guestAccount) receivePaymentProposeMsg(msg *PaymentProposeMsg) *PaymentAcceptMsg {
	// guest.createRatchetTxForOffChainPayment()

	return &PaymentAcceptMsg{
		ChannelID: msg.ChannelID,
		RoundNumber: msg.RoundNumber,
		//RecipientRatchetSig         *build.TransactionEnvelopeBuilder
		//RecipientSettleWithGuestSig *build.TransactionEnvelopeBuilder
		//RecipientSettleWithHostSig  *build.TransactionEnvelopeBuilder
	}
}

func (guest *guestAccount) loadSequenceNumber() int {
	sequenceNumber, err := loadSequenceNumber(guest.keyPair.Address())
	if err != nil {
		log.Fatal(err)
	}
	return sequenceNumber
}
