package script_utils

import (
	"github.com/LightningPeach/starlight-demo-impl/tools"
	"github.com/stellar/go/build"
	"github.com/stellar/go/hash"
	"github.com/stellar/go/strkey"
)

var (
	RPreImage = []byte{
		42, 42, 42, 42, 42, 42, 42, 42,
		42, 42, 42, 42, 42, 42, 42, 42,
		42, 42, 42, 42, 42, 42, 42, 42,
		42, 42, 42, 42, 42, 42, 42, 42,
	}
	RHash        = hash.Hash(RPreImage)
	RHashEncoded = strkey.MustEncode(strkey.VersionByteHashX, RHash[:])
)

func CreateRatchetTx(
	ratchetAddress,
	escrowAddress string,
	paymentTime uint64,
	roundSequenceNumber int,
) (
	*build.TransactionBuilder,
	error,
) {
	sequenceNumber, err := tools.LoadSequenceNumber(ratchetAddress)
	if err != nil {
		return nil, err
	}

	return build.Transaction(
		build.TestNetwork,
		build.SourceAccount{AddressOrSeed: ratchetAddress},
		build.Sequence{Sequence: uint64(sequenceNumber) + 1},
		build.Timebounds{
			MaxTime: paymentTime + tools.DefaultFinalityDelay + tools.DefaultMaxRoundDuration,
		},
		build.BumpSequence(
			build.SourceAccount{AddressOrSeed: escrowAddress},
			build.BumpTo(roundSequenceNumber+1),
		),
	)
}

func CreateSettleOnlyWithHostTx(
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
			MinTime: fundingTime + 2*tools.DefaultFinalityDelay + tools.DefaultMaxRoundDuration,
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

func CreateSettleWithGuestTx(
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
			MinTime: paymentTime + 2*tools.DefaultFinalityDelay + tools.DefaultMaxRoundDuration,
		},
		// Pay GuestAmount from EscrowAccount to GuestAccount
		build.Payment(
			build.Destination{AddressOrSeed: guestAddress},
			build.NativeAmount{Amount: guestAmount},
		),
	)
}

func CreateSettleWithHostTx(
	rsn,
	paymentTime uint64,
	escrowAddress,
	guestRatchetAddress,
	hostRatchetAddress,
	hostAddress string,
	htlcResolutionAddress string,
) (*build.TransactionBuilder, error) {

	mutations := []build.TransactionMutator{
		build.TestNetwork,
		build.SourceAccount{AddressOrSeed: escrowAddress},
		build.Sequence{Sequence: rsn + 3},
		build.Timebounds{
			MinTime: paymentTime + 2*tools.DefaultFinalityDelay + tools.DefaultMaxRoundDuration,
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
	}

	if tools.HtlcMode {
		mutations = append(mutations,
			build.AccountMerge(
				build.SourceAccount{AddressOrSeed: htlcResolutionAddress},
				build.Destination{AddressOrSeed: hostAddress},
			),
		)
	}

	tx, err := build.Transaction(mutations...)
	if err != nil {
		return nil, err
	}

	return tx, nil
}

func CreateSettleOnlyWithHostAndActiveHtlcTx(
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
	return build.Transaction(
		build.TestNetwork,
		build.SourceAccount{AddressOrSeed: escrowAddress},
		build.Sequence{Sequence: uint64(roundSequenceNumber) + 2},
		build.Timebounds{
			MinTime: fundingTime + 2*tools.DefaultFinalityDelay + tools.DefaultMaxRoundDuration,
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
		build.SetOptions(
			build.SourceAccount{AddressOrSeed: htlcResolutionAddress},
			build.Signer{Address: RHashEncoded, Weight: 1},
		),
	)
}

func CreateHtlcTimeoutTx(htlcResolutionAddress, hostAddress string, paymentTime uint64) (*build.TransactionBuilder, error) {
	sequence, err := tools.LoadSequenceNumber(htlcResolutionAddress)
	if err != nil {
		return nil, err
	}

	tx, err := build.Transaction(
		build.TestNetwork,
		build.SourceAccount{AddressOrSeed: htlcResolutionAddress},
		build.Sequence{Sequence: uint64(sequence) + 1},
		build.Timebounds{
			MinTime: paymentTime + 3*tools.DefaultFinalityDelay + tools.DefaultMaxRoundDuration,
		},
		build.Payment(
			build.Destination{AddressOrSeed: hostAddress},
			build.NativeAmount{Amount: tools.DefaultPaymentAmount},
		),
		build.AccountMerge(
			build.SourceAccount{AddressOrSeed: htlcResolutionAddress},
			build.Destination{AddressOrSeed: hostAddress},
		),
	)
	return tx, err
}

func CreateHtlcSuccessTx(htlcResolutionAddress, guestAddress string) (*build.TransactionBuilder, error) {
	sequence, err := tools.LoadSequenceNumber(htlcResolutionAddress)
	if err != nil {
		return nil, err
	}

	tx, err := build.Transaction(
		build.TestNetwork,
		build.SourceAccount{AddressOrSeed: htlcResolutionAddress},
		build.Sequence{Sequence: uint64(sequence) + 1},
		build.Payment(
			build.Destination{AddressOrSeed: guestAddress},
			build.NativeAmount{Amount: "100"},
		),
	)
	return tx, err
}
