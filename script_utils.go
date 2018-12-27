package main

import (
	"github.com/stellar/go/build"
)

func createRatchetTx(
	ratchetAddress,
	escrowAddress string,
	paymentTime uint64,
	roundSequenceNumber int,
) (
	*build.TransactionBuilder,
	error,
) {
	sequenceNumber, err := loadSequenceNumber(ratchetAddress)
	if err != nil {
		return nil, err
	}

	return build.Transaction(
		build.TestNetwork,
		build.SourceAccount{AddressOrSeed: ratchetAddress},
		build.Sequence{Sequence: uint64(sequenceNumber) + 1},
		build.Timebounds{
			MaxTime: paymentTime + defaultFinalityDelay + defaultMaxRoundDuration,
		},
		build.BumpSequence(
			build.SourceAccount{AddressOrSeed: escrowAddress},
			build.BumpTo(roundSequenceNumber+1),
		),
	)
}

func createSettleOnlyWithHostTx(
	hostAddress,
	escrowAddress,
	guestRatchetAddress,
	hostRatchetAddress string,
	fundingTime uint64,
	roundSequenceNumber int,
) (
	*build.TransactionBuilder,
	error,
) {

	return build.Transaction(
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
			build.SourceAccount{AddressOrSeed: guestRatchetAddress},
			build.Destination{AddressOrSeed: hostAddress},
		),
		build.AccountMerge(
			build.SourceAccount{AddressOrSeed: hostRatchetAddress},
			build.Destination{AddressOrSeed: hostAddress},
		),
	)
}

func createSettleWithGuestTx(
	rsn,
	paymentTime uint64,
	guestAddress,
	guestAmount,
	escrowAddress string,
) (
	*build.TransactionBuilder,
	error,
) {

	return build.Transaction(
		build.TestNetwork,
		build.SourceAccount{AddressOrSeed: escrowAddress},
		build.Sequence{Sequence: rsn + 2},
		build.Timebounds{
			MinTime: paymentTime + 2*defaultFinalityDelay + defaultMaxRoundDuration,
		},
		// Pay GuestAmount from EscrowAccount to GuestAccount
		build.Payment(
			build.Destination{AddressOrSeed: guestAddress},
			build.NativeAmount{Amount: guestAmount},
		),
	)
	//if err != nil {
	//	return nil, err
	//}

	//txe, err := tx.Sign(host.escrowKeyPair.Seed())
	//if err != nil {
	//	return nil, err
	//}

	//return &txe, nil
}

func createSettleWithHostTx(
	rsn,
	paymentTime uint64,
	escrowAddress,
	guestRatchetAddress,
	hostRatchetAddress,
	hostAddress string,
) (*build.TransactionBuilder, error) {

	tx, err := build.Transaction(
		build.TestNetwork,
		build.SourceAccount{AddressOrSeed: escrowAddress},
		build.Sequence{Sequence: rsn + 3},
		build.Timebounds{
			MinTime: paymentTime + 2*defaultFinalityDelay + defaultMaxRoundDuration,
		},
		//Merge account EscrowAccount to HostAccount
		//Merge account GuestRatchetAccount to HostAccount
		//Merge account HostRatchetAccount to HostAccount
		build.AccountMerge(
			build.SourceAccount{AddressOrSeed: escrowAddress},
			build.Destination{AddressOrSeed: hostAddress},
		),
		build.AccountMerge(
			build.SourceAccount{AddressOrSeed: guestRatchetAddress},
			build.Destination{AddressOrSeed: hostAddress},
		),
		build.AccountMerge(
			build.SourceAccount{AddressOrSeed: hostRatchetAddress},
			build.Destination{AddressOrSeed: hostAddress},
		),
	)
	if err != nil {
		return nil, err
	}

	//txe, err := tx.Sign(host.escrowKeyPair.Seed())
	//if err != nil {
	//	return nil, err
	//}
	return tx, nil
}