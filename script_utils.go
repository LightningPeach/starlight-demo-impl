package main

import "github.com/stellar/go/build"

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