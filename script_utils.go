package main

import (
	"fmt"
	"github.com/stellar/go/build"
	"github.com/stellar/go/hash"
	"github.com/stellar/go/strkey"
)

var (
	rPreImage = []byte{
		42, 42, 42, 42, 42, 42, 42, 42,
		42, 42, 42, 42, 42, 42, 42, 42,
		42, 42, 42, 42, 42, 42, 42, 42,
		42, 42, 42, 42, 42, 42, 42, 42,
	}
	rHash = hash.Hash(rPreImage)
	rHashEncoded = strkey.MustEncode(strkey.VersionByteHashX, rHash[:])
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
	htlcResolutionAddress string,
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
		build.AccountMerge(
			build.SourceAccount{AddressOrSeed: htlcResolutionAddress},
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
}

func createSettleWithHostTx(
	rsn,
	paymentTime uint64,
	escrowAddress,
	guestRatchetAddress,
	hostRatchetAddress,
	hostAddress,
	htlcResolutionAddress string,
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
		build.AccountMerge(
			build.SourceAccount{AddressOrSeed: htlcResolutionAddress},
			build.Destination{AddressOrSeed: hostAddress},
		),
	)
	if err != nil {
		return nil, err
	}

	return tx, nil
}

func createSettleOnlyWithHostAndActiveHtlcTx(
	hostAddress,
	escrowAddress,
	guestRatchetAddress,
	hostRatchetAddress string,
	fundingTime uint64,
	roundSequenceNumber int,
	htlcResolutionAddress,
	htlcAmount string,
) (
	*build.TransactionBuilder,
	error,
) {
	fmt.Println("@\n\n ############################################################ @\n\n")
	tmpl := `
	hostAddress           %v
	escrowAddress         %v
	guestRatchetAddress   %v
	hostRatchetAddress    %v
	fundingTime           %v
	roundSequenceNumber   %v
	htlcResolutionAddress %v
	htlcAmount            %v
	`
	fmt.Printf(
		tmpl,
		hostAddress,
		escrowAddress,
		guestRatchetAddress,
		hostRatchetAddress,
		fundingTime,
		roundSequenceNumber,
		htlcResolutionAddress,
		htlcAmount,
	)
	fmt.Println("@\n\n ############################################################ @\n\n")

	return build.Transaction(
		build.TestNetwork,
		build.SourceAccount{AddressOrSeed: escrowAddress},
		build.Sequence{Sequence: uint64(roundSequenceNumber) + 2},
		build.Timebounds{
			MinTime: fundingTime + 2*defaultFinalityDelay + defaultMaxRoundDuration,
		},
		build.Payment(
			build.Destination{AddressOrSeed: htlcResolutionAddress},
			build.NativeAmount{Amount: "100.00002"}, // defaultPaymentAmount + fee
		),
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
		//build.AccountMerge(
		//	build.SourceAccount{AddressOrSeed: htlcResolutionAddress},
		//	build.Destination{AddressOrSeed: hostAddress},
		//),
		build.SetOptions(
			build.SourceAccount{AddressOrSeed: htlcResolutionAddress},
			build.Signer{Address: rHashEncoded, Weight: 1},
		),
	)
}

func createHtlcTimeoutTx(htlcResolutionAddress, hostAddress string, paymentTime uint64) (*build.TransactionBuilder, error) {
	sequence, err := loadSequenceNumber(htlcResolutionAddress)
	if err != nil {
		return nil, err
	}

	tx, err := build.Transaction(
		build.TestNetwork,
		build.SourceAccount{AddressOrSeed: htlcResolutionAddress},
		build.Sequence{Sequence: uint64(sequence) + 1},
		build.Timebounds{
			MinTime: paymentTime + 3*defaultFinalityDelay + defaultMaxRoundDuration,
		},
		build.Payment(
			build.Destination{AddressOrSeed: hostAddress},
			build.NativeAmount{Amount: defaultPaymentAmount},
		),
		build.AccountMerge(
			build.SourceAccount{AddressOrSeed: htlcResolutionAddress},
			build.Destination{AddressOrSeed: hostAddress},
		),
	)
	return tx, err
}

func createHtlcSuccessTx(htlcResolutionAddress, guestAddress string) (*build.TransactionBuilder, error) {
	sequence, err := loadSequenceNumber(htlcResolutionAddress)
	if err != nil {
		return nil, err
	}

	tx, err := build.Transaction(
		build.TestNetwork,
		build.SourceAccount{AddressOrSeed: htlcResolutionAddress},
		build.Sequence{Sequence: uint64(sequence) + 1},
		build.Payment(
			// build.SourceAccount{AddressOrSeed: htlcResolutionAddress},
			build.Destination{AddressOrSeed: guestAddress},
			build.NativeAmount{Amount: "100"},
		),
	)
	return tx, err
}