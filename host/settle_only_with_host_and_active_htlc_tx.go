package host

import (
	"github.com/LightningPeach/starlight-demo-impl/script_utils"
	"github.com/LightningPeach/starlight-demo-impl/tools"
	"github.com/stellar/go/build"
	"github.com/stellar/go/xdr"
)

func (host *Account) CreateAndSignSettleOnlyWithHostAndActiveHtlcTx(
	rsn,
	paymentTime uint64,
	sig *xdr.DecoratedSignature,
) (*build.TransactionEnvelopeBuilder, error) {

	tx, err := script_utils.CreateSettleOnlyWithHostAndActiveHtlcTx(
		host.selfKeyPair.Address(),
		host.escrowKeyPair.Address(),
		host.guestRatchetAccount.KeyPair.Address(),
		host.hostRatchetAccount.KeyPair.Address(),
		paymentTime,
		int(rsn),
		host.HtlcResolutionAccount.KeyPair.Address(),
		tools.DefaultPaymentAmount,
	)
	if err != nil {
		return nil, err
	}

	txe, err := tx.Sign(host.escrowKeyPair.Seed())
	if err != nil {
		return nil, err
	}
	txe.E.Signatures = append(txe.E.Signatures, *sig)

	return &txe, nil
}
