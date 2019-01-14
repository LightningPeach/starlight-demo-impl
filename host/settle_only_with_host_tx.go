package host

import (
	"github.com/LightningPeach/starlight-demo-impl/script_utils"
	"github.com/LightningPeach/starlight-demo-impl/tools"
	"github.com/stellar/go/xdr"
)

func (host *Account) PublishSettleOnlyWithHostTx(
	guestSettleOnlyWithHostSig *xdr.DecoratedSignature,
	fundingTime uint64,
) error {
	rsn := tools.RoundSequenceNumber(host.BaseSequenceNumber, 1)

	settleOnlyWithHostTx, err := script_utils.CreateSettleOnlyWithHostTx(
		host.selfKeyPair.Address(),
		host.escrowKeyPair.Address(),
		host.guestRatchetAccount.KeyPair.Address(),
		host.hostRatchetAccount.KeyPair.Address(),
		fundingTime,
		rsn,
	)
	if err != nil {
		return err
	}

	txe, err := settleOnlyWithHostTx.Sign(host.escrowKeyPair.Seed())
	if err != nil {
		return err
	}

	txe.E.Signatures = append(txe.E.Signatures, *guestSettleOnlyWithHostSig)

	if err := host.PublishTx(&txe); err != nil {
		tools.ShowDetailError(err)
		return err
	}
	return nil
}
