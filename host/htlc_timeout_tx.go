package host

import (
	"github.com/LightningPeach/starlight-demo-impl/script_utils"
	"github.com/stellar/go/build"
	"github.com/stellar/go/xdr"
)

func (host *Account) CreateAndSignHtlcTimeoutTx(guestSig *xdr.DecoratedSignature, paymentTime uint64) (*build.TransactionEnvelopeBuilder, error) {
	tx, err := script_utils.CreateHtlcTimeoutTx(host.HtlcResolutionAccount.KeyPair.Address(), host.selfKeyPair.Address(), paymentTime)
	if err != nil {
		return nil, err
	}

	txe, err := tx.Sign(host.escrowKeyPair.Seed())
	if err != nil {
		return nil, err
	}
	txe.E.Signatures = append(txe.E.Signatures, *guestSig)

	return &txe, nil
}
