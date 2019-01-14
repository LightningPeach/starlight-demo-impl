package host

import (
	"fmt"

	"github.com/LightningPeach/starlight-demo-impl/script_utils"
	"github.com/LightningPeach/starlight-demo-impl/tools"
	"github.com/stellar/go/build"
	"github.com/stellar/go/xdr"
)

func (host *Account) PublishSettleWithHostTx(rsn, paymentTime uint64, sig *xdr.DecoratedSignature) error {
	fmt.Println("publish settle with host tx")

	tx, err := host.createAndSignSettleWithHostTx(rsn, paymentTime)
	if err != nil {
		return err
	}
	tx.E.Signatures = append(tx.E.Signatures, *sig)

	if err := host.PublishTx(tx); err != nil {
		tools.ShowDetailError(err)
		return err
	}
	return nil
}

func (host *Account) createAndSignSettleWithHostTx(rsn, paymentTime uint64) (*build.TransactionEnvelopeBuilder, error) {
	tx, err := script_utils.CreateSettleWithHostTx(
		rsn,
		paymentTime,
		host.escrowKeyPair.Address(),
		host.guestRatchetAccount.KeyPair.Address(),
		host.hostRatchetAccount.KeyPair.Address(),
		host.selfKeyPair.Address(),
		host.HtlcResolutionAccount.KeyPair.Address(),
	)
	if err != nil {
		return nil, err
	}

	signers := []string{host.escrowKeyPair.Seed()}
	if tools.HtlcMode {
		signers = append(signers, host.HtlcResolutionAccount.KeyPair.Seed())
	}

	txe, err := tx.Sign(signers...)
	if err != nil {
		return nil, err
	}
	return &txe, nil
}