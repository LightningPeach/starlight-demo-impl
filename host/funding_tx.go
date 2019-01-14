package host

import (
	"fmt"

	"github.com/LightningPeach/starlight-demo-impl/tools"
	"github.com/stellar/go/build"
)

func (host *Account) PublishFundingTx(guestEscrowPubKey string) error {
	fmt.Println("publish funding tx")
	host.guestAddress = guestEscrowPubKey
	fundingTime := tools.GetBlockChainTime()

	mutations := []build.TransactionMutator{
		build.TestNetwork,
		build.SourceAccount{AddressOrSeed: host.selfKeyPair.Address()},
		build.Sequence{Sequence: uint64(host.loadSequenceNumber()) + 1},
		build.Timebounds{
			MaxTime: uint64(fundingTime + tools.DefaultMaxRoundDuration + tools.DefaultFinalityDelay),
		},
		// Escrow Account
		build.Payment(
			build.Destination{AddressOrSeed: host.escrowKeyPair.Address()},
			// TODO(evg): refactor it
			build.NativeAmount{Amount: "500.50008"}, // defaultHostAmount + 0.5 + 8 * feerate
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
			build.Destination{AddressOrSeed: host.guestRatchetAccount.KeyPair.Address()},
			build.NativeAmount{Amount: "1.00001"}, // 1 + 1 * feerate
		),
		build.SetOptions(
			build.SourceAccount{AddressOrSeed: host.guestRatchetAccount.KeyPair.Address()},
			build.MasterWeight(0),
			build.SetThresholds(2, 2, 2),
			build.Signer{
				Address: guestEscrowPubKey,
				Weight:  1,
			},
		),
		build.SetOptions(
			build.SourceAccount{AddressOrSeed: host.guestRatchetAccount.KeyPair.Address()},
			build.Signer{
				Address: host.escrowKeyPair.Address(),
				Weight:  1,
			},
		),
		// HostRatchetAccount
		build.Payment(
			build.Destination{AddressOrSeed: host.hostRatchetAccount.KeyPair.Address()},
			build.NativeAmount{Amount: "0.50001"}, // 0.5 + 1 * feerate
		),
		build.SetOptions(
			build.SourceAccount{AddressOrSeed: host.hostRatchetAccount.KeyPair.Address()},
			build.MasterWeight(0),
			build.Signer{
				Address: host.escrowKeyPair.Address(),
				Weight:  1,
			},
		),
	}

	if tools.HtlcMode {
		mutations = append(mutations,
			// htlcResolutionAccountType
			build.Payment(
				build.Destination{AddressOrSeed: host.HtlcResolutionAccount.KeyPair.Address()},
				build.NativeAmount{Amount: "1.5"},
			),
			build.SetOptions(
				build.SourceAccount{AddressOrSeed: host.HtlcResolutionAccount.KeyPair.Address()},
				build.MasterWeight(0),
				build.SetThresholds(3, 3, 3),
				build.Signer{
					Address: host.escrowKeyPair.Address(),
					Weight:  1,
				},
			),
			build.SetOptions(
				build.SourceAccount{AddressOrSeed: host.HtlcResolutionAccount.KeyPair.Address()},
				build.Signer{
					Address: guestEscrowPubKey,
					Weight:  2,
				},
			),
		)
	}

	tx, err := build.Transaction(mutations...)
	if err != nil {
		return err
	}

	// Sign the transaction to prove you are actually the person sending it.
	signers := []string{
		host.selfKeyPair.Seed(),
		host.escrowKeyPair.Seed(),
		host.hostRatchetAccount.KeyPair.Seed(),
		host.guestRatchetAccount.KeyPair.Seed(),
	}
	if tools.HtlcMode {
		signers = append(signers, host.HtlcResolutionAccount.KeyPair.Seed())
	}

	txe, err := tx.Sign(signers...)
	if err != nil {
		return err
	}

	return host.PublishTx(&txe)
}
