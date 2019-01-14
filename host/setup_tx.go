package host

import (
	"fmt"

	"github.com/LightningPeach/starlight-demo-impl/tools"
	"github.com/stellar/go/build"
)

func (host *Account) PublishSetupAccountTx(account tools.AccountType) error {
	fmt.Printf("creating: %v\n", account)
	var dest string
	switch account {
	case tools.HostRatchetAccount:
		dest = host.hostRatchetAccount.KeyPair.Address()
	case tools.GuestRatchetAccount:
		dest = host.guestRatchetAccount.KeyPair.Address()
	case tools.HtlcResolutionAccountType:
		dest = host.HtlcResolutionAccount.KeyPair.Address()
	case tools.EscrowAccount:
		dest = host.escrowKeyPair.Address()
	}

	tx, err := build.Transaction(
		build.TestNetwork,
		build.SourceAccount{AddressOrSeed: host.selfKeyPair.Address()},
		build.Sequence{Sequence: uint64(host.loadSequenceNumber()) + 1},
		build.CreateAccount(
			build.Destination{AddressOrSeed: dest},
			build.NativeAmount{Amount: "1"},
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

	if err := host.PublishTx(&txe); err != nil {
		return err
	}

	if account == tools.EscrowAccount {
		sequence, err := tools.LoadSequenceNumber(host.escrowKeyPair.Address())
		if err != nil {
			return err
		}
		host.BaseSequenceNumber = sequence
	}
	fmt.Printf("balance: %v\n\n", tools.LoadBalance(dest))
	return nil
}