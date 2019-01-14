package host

import (
	"fmt"

	"github.com/LightningPeach/starlight-demo-impl/script_utils"
	"github.com/LightningPeach/starlight-demo-impl/tools"
	"github.com/stellar/go/build"
	"github.com/stellar/go/xdr"
)

func (host *Account) PublishRatchetTx(sig *xdr.DecoratedSignature, state tools.State) error {
	fmt.Println("publish ratchet tx")
	var (
		paymentTime uint64
		roundNumber int
	)
	switch state {
	case tools.PaymentState:
		paymentTime = host.cache.paymentProposeMsg.PaymentTime
		roundNumber = host.cache.paymentProposeMsg.RoundNumber
	case tools.HtlcPaymentState:
		paymentTime = host.cache.htlcPaymentProposeMsg.PaymentTime
		roundNumber = host.cache.htlcPaymentProposeMsg.RoundNumber
	}
	tx, err := host.CreateAndSignRatchetTxForSelf(
		sig,
		paymentTime,
		tools.RoundSequenceNumber(host.BaseSequenceNumber, roundNumber),
	)
	if err != nil {
		return err
	}
	if err := host.PublishTx(tx); err != nil {
		tools.ShowDetailError(err)
		return err
	}
	return nil
}

func (host *Account) CreateAndSignRatchetTxForSelf(
	sig *xdr.DecoratedSignature,
	paymentTime uint64,
	roundSequenceNumber int,
) (
	*build.TransactionEnvelopeBuilder,
	error,
) {

	tx, err := host.CreateRatchetTxForSelf(paymentTime, roundSequenceNumber)
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

func (host *Account) CreateRatchetTxForSelf(
	paymentTime uint64,
	roundSequenceNumber int,
) (
	*build.TransactionBuilder,
	error,
) {
	return script_utils.CreateRatchetTx(
		host.hostRatchetAccount.KeyPair.Address(),
		host.escrowKeyPair.Address(),
		paymentTime,
		roundSequenceNumber,
	)
}