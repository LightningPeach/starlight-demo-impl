package host

import (
	"fmt"

	"github.com/LightningPeach/starlight-demo-impl/script_utils"
	"github.com/LightningPeach/starlight-demo-impl/tools"
	"github.com/stellar/go/build"
	"github.com/stellar/go/xdr"
)

func (host *Account) PublishSettleWithGuestTx(
	rsn,
	paymentTime uint64,
	guestAmount string,
	sig *xdr.DecoratedSignature,
) error {
	fmt.Println("publish settle with guest tx")

	tx, err := host.CreateAndSignSettleWithGuestTx(rsn, paymentTime, guestAmount)
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

func (host *Account) CreateAndSignSettleWithGuestTx(
	rsn uint64,
	paymentTime uint64,
	guestAmount string,
) (
	*build.TransactionEnvelopeBuilder,
	error,
) {

	tx, err := script_utils.CreateSettleWithGuestTx(
		rsn,
		paymentTime,
		host.guestAddress,
		guestAmount,
		host.escrowKeyPair.Address(),
	)
	if err != nil {
		return nil, err
	}

	txe, err := tx.Sign(host.escrowKeyPair.Seed())
	if err != nil {
		return nil, err
	}

	return &txe, nil
}
